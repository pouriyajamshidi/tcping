package printers

import (
	"fmt"
	"math"
	"net/netip"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/types"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

const (
	probeDataQuery = `SELECT
		type,
		success,
		timestamp,
		ip_address,
		hostname,
		port,
		source_address,
		destination_is_ip,
		time,
		ongoing_successful_probes,
		ongoing_unsuccessful_probes
		FROM %s WHERE type = ?`

	statsDataQuery = `SELECT
		type,
		timestamp,
		ip_address,
		hostname,
		port,
		total_duration,
		total_uptime,
		total_downtime,
		total_packets,
		total_successful_packets,
		total_unsuccessful_packets,
		total_packet_loss_percent,
		longest_uptime,
		longest_downtime,
		hostname_resolve_retries,
		hostname_changes,
		last_successful_probe,
		last_unsuccessful_probe,
		longest_consecutive_uptime_start,
		longest_consecutive_uptime_end,
		longest_consecutive_downtime_start,
		longest_consecutive_downtime_end,
		latency_min,
		latency_avg,
		latency_max,
		start_time,
		end_time
		FROM %s WHERE type = ?`
)

func TestNewDatabasePrinter(t *testing.T) {
	tests := []struct {
		name    string
		cfg     PrinterConfig
		wantErr bool
	}{
		{
			name: "memory database",
			cfg: PrinterConfig{
				OutputDBPath: ":memory:",
				Target:       "localhost",
				Port:         "8001",
			},
			wantErr: false,
		},
		{
			name: "file database",
			cfg: PrinterConfig{
				OutputDBPath: "test",
				Target:       "example.com",
				Port:         "80",
			},
			wantErr: false,
		},
	}

	defer os.Remove("test.db")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewDatabasePrinter(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("NewDatabasePrinter() error = %v", err)
				return
			}
			defer db.Done()

			// Verify tables were created
			query := "SELECT name FROM sqlite_master WHERE type='table';"
			var foundProbe, foundStats bool
			err = sqlitex.Execute(db.Conn, query, &sqlitex.ExecOptions{
				ResultFunc: func(stmt *sqlite.Stmt) error {
					tableName := stmt.ColumnText(0)
					if tableName == db.probeTableName {
						foundProbe = true
					}
					if tableName == db.statsTableName {
						foundStats = true
					}
					return nil
				},
			})
			if err != nil {
				t.Errorf("failed to query tables: %v", err)
			}
			if !foundProbe {
				t.Error("probe table not created")
			}
			if !foundStats {
				t.Error("stats table not created")
			}
		})
	}
}

func TestDatabasePrinter_PrintProbeSuccess(t *testing.T) {
	cfg := PrinterConfig{
		OutputDBPath:      ":memory:",
		Target:            "localhost",
		Port:              "8001",
		WithTimestamp:     true,
		WithSourceAddress: true,
	}

	db, err := NewDatabasePrinter(cfg)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Done()

	tests := []struct {
		name       string
		startTime  time.Time
		sourceAddr string
		opts       types.Options
		streak     uint
		rtt        string
	}{
		{
			name:       "IP destination",
			startTime:  time.Now(),
			sourceAddr: "192.168.1.2",
			opts: types.Options{
				IP:       netip.MustParseAddr("192.168.1.1"),
				Hostname: "192.168.1.1",
				Port:     80,
			},
			streak: 10,
			rtt:    "30ms",
		},
		{
			name:       "hostname destination",
			startTime:  time.Now(),
			sourceAddr: "192.168.1.2",
			opts: types.Options{
				IP:       netip.MustParseAddr("192.168.1.1"),
				Hostname: "example.com",
				Port:     80,
			},
			streak: 10,
			rtt:    "30ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.PrintProbeSuccess(tt.startTime, tt.sourceAddr, tt.opts, tt.streak, tt.rtt)

			query := fmt.Sprintf(probeDataQuery, db.probeTableName)
			err := sqlitex.Execute(db.Conn, query, &sqlitex.ExecOptions{
				Args: []interface{}{eventTypeProbe},
				ResultFunc: func(stmt *sqlite.Stmt) error {
					assertEquals(t, stmt.ColumnText(1), "true")
					assertEquals(t, stmt.ColumnText(3), tt.opts.IP.String())
					assertEquals(t, stmt.ColumnInt64(5), int64(tt.opts.Port))
					assertEquals(t, stmt.ColumnText(6), tt.sourceAddr)
					assertEquals(t, stmt.ColumnText(8), tt.rtt)
					assertEquals(t, stmt.ColumnInt64(9), int64(tt.streak))
					return nil
				},
			})
			if err != nil {
				t.Errorf("failed to query probe data: %v", err)
			}
		})
	}
}

func TestDatabasePrinter_PrintProbeFailure(t *testing.T) {
	cfg := PrinterConfig{
		OutputDBPath:  ":memory:",
		Target:        "localhost",
		Port:          "8001",
		WithTimestamp: true,
	}

	db, err := NewDatabasePrinter(cfg)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Done()

	tests := []struct {
		name      string
		startTime time.Time
		opts      types.Options
		streak    uint
	}{
		{
			name:      "IP destination failure",
			startTime: time.Now(),
			opts: types.Options{
				IP:       netip.MustParseAddr("192.168.1.1"),
				Hostname: "192.168.1.1",
				Port:     80,
			},
			streak: 3,
		},
		{
			name:      "hostname destination failure",
			startTime: time.Now(),
			opts: types.Options{
				IP:       netip.MustParseAddr("192.168.1.1"),
				Hostname: "example.com",
				Port:     80,
			},
			streak: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.PrintProbeFailure(tt.startTime, tt.opts, tt.streak)

			query := fmt.Sprintf(probeDataQuery, db.probeTableName)
			err := sqlitex.Execute(db.Conn, query, &sqlitex.ExecOptions{
				Args: []interface{}{eventTypeProbe},
				ResultFunc: func(stmt *sqlite.Stmt) error {
					assertEquals(t, stmt.ColumnText(1), "false")
					assertEquals(t, stmt.ColumnText(3), tt.opts.IP.String())
					assertEquals(t, stmt.ColumnInt64(5), int64(tt.opts.Port))
					assertEquals(t, stmt.ColumnInt64(10), int64(tt.streak))
					return nil
				},
			})
			if err != nil {
				t.Errorf("failed to query probe data: %v", err)
			}
		})
	}
}

func TestDatabasePrinter_PrintStatistics(t *testing.T) {
	cfg := PrinterConfig{
		OutputDBPath: ":memory:",
		Target:       "localhost",
		Port:         "8001",
	}

	db, err := NewDatabasePrinter(cfg)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Done()

	stats := createMockStats()
	db.PrintStatistics(stats)

	query := fmt.Sprintf(statsDataQuery, db.statsTableName)
	err = sqlitex.Execute(db.Conn, query, &sqlitex.ExecOptions{
		Args: []interface{}{eventTypeStatistics},
		ResultFunc: func(stmt *sqlite.Stmt) error {
			// Verify all statistics fields
			assertEquals(t, stmt.ColumnText(2), stats.Options.IP.String())
			assertEquals(t, stmt.ColumnText(3), stats.Options.Hostname)
			assertEquals(t, stmt.ColumnInt64(4), int64(stats.Options.Port))
			assertEquals(t, stmt.ColumnInt64(9), int64(stats.TotalSuccessfulProbes))
			assertEquals(t, stmt.ColumnInt64(10), int64(stats.TotalUnsuccessfulProbes))

			packetLoss := (float64(stats.TotalUnsuccessfulProbes) / float64(stats.TotalSuccessfulProbes+stats.TotalUnsuccessfulProbes)) * 100
			assertEquals(t, fmt.Sprintf("%.2f", stmt.ColumnFloat(11)), fmt.Sprintf("%.2f", packetLoss))

			// Verify timestamps
			startTime, err := time.Parse(time.DateTime, stmt.ColumnText(25))
			if err != nil {
				t.Errorf("failed to parse start time: %v", err)
			}
			assertEquals(t, startTime.Format(time.DateTime), stats.StartTime.Format(time.DateTime))

			// Verify RTT stats
			assertEquals(t, fmt.Sprintf("%.3f", stmt.ColumnFloat(22)), fmt.Sprintf("%.3f", stats.RttResults.Min))
			assertEquals(t, fmt.Sprintf("%.3f", stmt.ColumnFloat(23)), fmt.Sprintf("%.3f", stats.RttResults.Average))
			assertEquals(t, fmt.Sprintf("%.3f", stmt.ColumnFloat(24)), fmt.Sprintf("%.3f", stats.RttResults.Max))

			return nil
		},
	})
	if err != nil {
		t.Errorf("failed to query statistics data: %v", err)
	}
}

func TestSanitizeTableName(t *testing.T) {
	now := time.Now().Format(time.DateTime)
	now = strings.ReplaceAll(now, " ", "_")
	now = strings.ReplaceAll(now, "-", "_")
	now = strings.ReplaceAll(now, ":", "_")

	tests := []struct {
		name     string
		hostname string
		port     string
		want     string
	}{
		{
			name:     "basic hostname",
			hostname: "example.com",
			port:     "80",
			want:     fmt.Sprintf("example_com_80__%s", now),
		},
		{
			name:     "IP address",
			hostname: "192.168.1.1",
			port:     "443",
			want:     fmt.Sprintf("_192_168_1_1_443__%s", now),
		},
		{
			name:     "hostname with hyphens",
			hostname: "test-server-1",
			port:     "8080",
			want:     fmt.Sprintf("test_server_1_8080__%s", now),
		},
		{
			name:     "numeric hostname",
			hostname: "123server",
			port:     "22",
			want:     fmt.Sprintf("_123server_22__%s", now),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeTableName(tt.hostname, tt.port)
			if !strings.HasPrefix(got, tt.want) {
				t.Errorf("sanitizeTableName() = %v, want prefix %v", got, tt.want)
			}
		})
	}
}

// Helper functions

func createMockStats() types.Tcping {
	now := time.Now()
	return types.Tcping{
		StartTime:              now,
		EndTime:                now.Add(10 * time.Minute),
		LastSuccessfulProbe:    now.Add(1 * time.Minute),
		RetriedHostnameLookups: 10,
		LongestUptime: types.LongestTime{
			Start:    now.Add(20 * time.Second),
			End:      now.Add(80 * time.Second),
			Duration: time.Minute,
		},
		LongestDowntime: types.LongestTime{
			Start:    now.Add(20 * time.Second),
			End:      now.Add(140 * time.Second),
			Duration: time.Minute * 2,
		},
		Options: types.Options{
			IP:       netip.MustParseAddr("192.168.1.1"),
			Hostname: "example.com",
			Port:     1234,
		},
		TotalUptime:             32 * time.Second,
		TotalDowntime:           60 * time.Second,
		TotalSuccessfulProbes:   201,
		TotalUnsuccessfulProbes: 123,
		RttResults: types.RttResult{
			HasResults: true,
			Min:        2.832,
			Average:    3.8123,
			Max:        4.0932,
		},
		HostnameChanges: createMockHostnameChanges(),
	}
}

func createMockHostnameChanges() []types.HostnameChange {
	now := time.Now()
	return []types.HostnameChange{
		{
			Addr: netip.MustParseAddr("192.168.1.1"),
			When: now,
		},
		{
			Addr: netip.MustParseAddr("192.168.1.2"),
			When: now.Add(time.Minute),
		},
		{
			Addr: netip.MustParseAddr("192.168.1.3"),
			When: now.Add(2 * time.Minute),
		},
	}
}

func assertEquals[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func toFixedFloat(input float32, precision int) float32 {
	output := math.Pow10(precision)
	return float32(math.Round(float64(input)*output) / output)
}
