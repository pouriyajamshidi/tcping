//go:build !windows

package printers

import (
	"fmt"
	"net"
	"net/netip"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"

	"github.com/pouriyajamshidi/tcping/v3/config"
	"github.com/pouriyajamshidi/tcping/v3/stats"
)

func TestAddDBExtension(t *testing.T) {
	tests := map[string]string{
		":memory:":   ":memory:",
		"results.db": "results.db",
		"results":    "results.db",
	}

	for input, want := range tests {
		got := addDbExtension(input)
		if got != want {
			t.Fatalf("addDbExtension(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestBoolToInt(t *testing.T) {
	if got := boolToInt(true); got != 1 {
		t.Fatalf("boolToInt(true) = %d, want 1", got)
	}
	if got := boolToInt(false); got != 0 {
		t.Fatalf("boolToInt(false) = %d, want 0", got)
	}
}

func TestSanitizeTableName(t *testing.T) {
	name := sanitizeTableName("example.com-1", "443")
	if !regexp.MustCompile(`^example_com_1_443__[0-9]{4}_[0-9]{2}_[0-9]{2}_[0-9]{2}_[0-9]{2}_[0-9]{2}$`).MatchString(name) {
		t.Fatalf("unexpected table name %q", name)
	}

	name = sanitizeTableName("1.2.3.4", "443")
	if name[0] != '_' {
		t.Fatalf("table name %q should be prefixed when target starts with a number", name)
	}
}

func TestNewDatabasePrinterConfiguresSQLite(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "results.db")

	printer, err := NewDatabasePrinter(Config{Target: "example.com", Port: 443, OutputDBPath: dbPath})
	if err != nil {
		t.Fatal(err)
	}
	defer printer.done()

	var journalMode, synchronous, busyTimeout string
	err = sqlitex.Execute(printer.conn,
		"SELECT (SELECT journal_mode FROM pragma_journal_mode), (SELECT synchronous FROM pragma_synchronous), (SELECT timeout FROM pragma_busy_timeout)",
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				journalMode = stmt.ColumnText(0)
				synchronous = stmt.ColumnText(1)
				busyTimeout = stmt.ColumnText(2)
				return nil
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if journalMode != "wal" {
		t.Errorf("journal_mode = %q, want wal", journalMode)
	}
	if synchronous != "1" {
		t.Errorf("synchronous = %q, want 1 (NORMAL)", synchronous)
	}
	if busyTimeout != "5000" {
		t.Errorf("busy_timeout = %q, want 5000", busyTimeout)
	}
}

func TestNewDatabasePrinterSchema(t *testing.T) {
	printer, err := NewDatabasePrinter(Config{Target: "example.com", Port: 443, OutputDBPath: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	defer printer.done()

	assertColumns := func(table string, expected map[string]string) {
		t.Helper()

		got := make(map[string]string)
		err := sqlitex.Execute(printer.conn,
			"SELECT name, type FROM pragma_table_info(?)",
			&sqlitex.ExecOptions{
				Args: []any{table},
				ResultFunc: func(stmt *sqlite.Stmt) error {
					got[stmt.ColumnText(0)] = stmt.ColumnText(1)
					return nil
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		for column, wantType := range expected {
			if got[column] != wantType {
				t.Fatalf("column %q has type %q, want %q", column, got[column], wantType)
			}
		}
	}

	assertColumns(printer.probeTableName, map[string]string{
		"reachable":                   "INTEGER",
		"destination_is_ip":           "INTEGER",
		"latency":                     "REAL",
		"ongoing_successful_probes":   "INTEGER",
		"ongoing_unsuccessful_probes": "INTEGER",
	})

	assertColumns(printer.statsTableName, map[string]string{
		"total_packet_loss_percent": "REAL",
		"latency_min":               "REAL",
		"latency_avg":               "REAL",
		"latency_max":               "REAL",
		"hostname_changes":          "TEXT",
	})
}

func TestInsertProbeStoresSQLiteTypes(t *testing.T) {
	printer, err := NewDatabasePrinter(Config{Target: "example.com", Port: 443, OutputDBPath: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	defer printer.done()

	ip := netip.MustParseAddr("192.0.2.10")

	probeStats := &stats.Statistics{
		Hostname:                "example.com",
		IP:                      ip,
		Port:                    443,
		DestIsIP:                true,
		OngoingSuccessfulProbes: 7,
	}
	printer.insertProbe(true, probeStats, "12.345", 7, 0)

	var (
		reachable        string
		destinationIsIP  string
		latency          string
		successfulProbes string
	)

	if err := sqlitex.Execute(printer.conn,
		fmt.Sprintf(`SELECT typeof(reachable), typeof(destination_is_ip), typeof(latency), typeof(ongoing_successful_probes) FROM %s`, printer.probeTableName),
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				reachable = stmt.ColumnText(0)
				destinationIsIP = stmt.ColumnText(1)
				latency = stmt.ColumnText(2)
				successfulProbes = stmt.ColumnText(3)
				return nil
			},
		},
	); err != nil {
		t.Fatal(err)
	}

	if reachable != "integer" {
		t.Errorf("typeof(reachable) = %q, want integer", reachable)
	}
	if destinationIsIP != "integer" {
		t.Errorf("typeof(destination_is_ip) = %q, want integer", destinationIsIP)
	}
	if latency != "real" {
		t.Errorf("typeof(latency) = %q, want real", latency)
	}
	if successfulProbes != "integer" {
		t.Errorf("typeof(ongoing_successful_probes) = %q, want integer", successfulProbes)
	}
}

// dbTestPrinter is a printer writing to an in-memory database, along with a
// query helper for reading back the single row a test just wrote.
func dbTestPrinter(t *testing.T, cfg Config) *DatabasePrinter {
	t.Helper()

	cfg.OutputDBPath = ":memory:"
	if cfg.Target == "" {
		cfg.Target = "example.com"
	}
	if cfg.Port == 0 {
		cfg.Port = 443
	}

	printer, err := NewDatabasePrinter(cfg)
	if err != nil {
		t.Fatalf("NewDatabasePrinter failed: %v", err)
	}
	t.Cleanup(printer.done)

	return printer
}

// queryRow runs a one-row query and hands each column back as text, with
// "null" for the columns SQLite stored as NULL. Reading everything as text
// keeps the assertions in the tests readable.
func queryRow(t *testing.T, printer *DatabasePrinter, query string) []string {
	t.Helper()

	var row []string
	rows := 0

	if err := sqlitex.Execute(printer.conn, query, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			rows++
			row = make([]string, stmt.ColumnCount())
			for i := range row {
				if stmt.ColumnType(i) == sqlite.TypeNull {
					row[i] = "null"
					continue
				}
				row[i] = stmt.ColumnText(i)
			}
			return nil
		},
	}); err != nil {
		t.Fatalf("query %q failed: %v", query, err)
	}

	if rows != 1 {
		t.Fatalf("query %q returned %d rows, want exactly 1", query, rows)
	}

	return row
}

// dbTestStats is a successful TCP probe against a hostname. Tests change only
// the fields they care about.
func dbTestStats() *stats.Statistics {
	return &stats.Statistics{
		Hostname:                "example.com",
		IP:                      netip.MustParseAddr("93.184.216.34"),
		Port:                    443,
		Protocol:                config.TCP,
		LatestRTT:               3.5,
		OngoingSuccessfulProbes: 2,
		StartTime:               time.Now(),
	}
}

func TestDatabaseProbeSuccessRow(t *testing.T) {
	printer := dbTestPrinter(t, Config{})

	printer.PrintProbeSuccess(dbTestStats())

	got := queryRow(t, printer, fmt.Sprintf(
		`SELECT reachable, hostname, ip_address, port, latency, protocol,
		        ongoing_successful_probes, ongoing_unsuccessful_probes
		 FROM %s`, printer.probeTableName))

	want := []string{"1", "example.com", "93.184.216.34", "443", "3.5", "TCP", "2", "0"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("column %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDatabaseProbeFailureRow(t *testing.T) {
	printer := dbTestPrinter(t, Config{})

	probeStats := dbTestStats()
	probeStats.OngoingSuccessfulProbes = 0
	probeStats.OngoingUnsuccessfulProbes = 4

	printer.PrintProbeFailure(probeStats)

	got := queryRow(t, printer, fmt.Sprintf(
		`SELECT reachable, latency, ongoing_successful_probes, ongoing_unsuccessful_probes
		 FROM %s`, printer.probeTableName))

	// A failed probe has no round trip time, so latency must be NULL rather
	// than a zero that would drag an average down.
	want := []string{"0", "null", "0", "4"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("column %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDatabaseTCPProbeLeavesHTTPAndUDPColumnsNull(t *testing.T) {
	printer := dbTestPrinter(t, Config{})

	printer.PrintProbeSuccess(dbTestStats())

	got := queryRow(t, printer, fmt.Sprintf(
		`SELECT status_code, http_version, tls_version, tls_cipher_suite,
		        certificate_expiry, connect_ms, tls_handshake_ms, ttfb_ms,
		        udp_probe_number, udp_result
		 FROM %s`, printer.probeTableName))

	for i, column := range got {
		if column != "null" {
			t.Errorf("column %d = %q, want null for a TCP probe", i, column)
		}
	}
}

func TestDatabaseHTTPSProbeColumns(t *testing.T) {
	printer := dbTestPrinter(t, Config{Target: "example.com"})

	probeStats := dbTestStats()
	probeStats.Protocol = config.HTTPS
	probeStats.HTTP = stats.HTTPInfo{
		StatusCode:      200,
		Status:          "200 OK",
		Proto:           "HTTP/2.0",
		TLSVersion:      "TLS 1.3",
		TLSCipherSuite:  "TLS_AES_128_GCM_SHA256",
		CertExpiry:      time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC),
		ConnectDuration: 12 * time.Millisecond,
		TLSDuration:     34 * time.Millisecond,
		TimeToFirstByte: 56 * time.Millisecond,
	}

	printer.PrintProbeSuccess(probeStats)

	got := queryRow(t, printer, fmt.Sprintf(
		`SELECT status_code, http_version, tls_version, tls_cipher_suite,
		        connect_ms, tls_handshake_ms, ttfb_ms
		 FROM %s`, printer.probeTableName))

	want := []string{"200", "HTTP/2.0", "TLS 1.3", "TLS_AES_128_GCM_SHA256", "12.0", "34.0", "56.0"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("column %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDatabasePlainHTTPProbeHasNoTLSColumns(t *testing.T) {
	printer := dbTestPrinter(t, Config{})

	probeStats := dbTestStats()
	probeStats.Protocol = config.HTTP
	probeStats.HTTP = stats.HTTPInfo{
		StatusCode:      200,
		Proto:           "HTTP/1.1",
		ConnectDuration: 12 * time.Millisecond,
		TimeToFirstByte: 56 * time.Millisecond,
	}

	printer.PrintProbeSuccess(probeStats)

	got := queryRow(t, printer, fmt.Sprintf(
		`SELECT status_code, tls_version, tls_cipher_suite, certificate_expiry
		 FROM %s`, printer.probeTableName))

	// Plain HTTP has no handshake and no certificate, so those stay NULL
	// while the status code is still recorded.
	want := []string{"200", "null", "null", "null"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("column %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDatabaseHTTPProbeWithoutAResponseIsAllNull(t *testing.T) {
	printer := dbTestPrinter(t, Config{})

	probeStats := dbTestStats()
	probeStats.Protocol = config.HTTPS
	probeStats.OngoingSuccessfulProbes = 0
	probeStats.OngoingUnsuccessfulProbes = 1

	printer.PrintProbeFailure(probeStats)

	got := queryRow(t, printer, fmt.Sprintf(
		`SELECT status_code, http_version, tls_version, connect_ms, ttfb_ms
		 FROM %s`, printer.probeTableName))

	for i, column := range got {
		if column != "null" {
			t.Errorf("column %d = %q, want null when nothing answered", i, column)
		}
	}
}

func TestDatabaseUDPProbeColumns(t *testing.T) {
	tests := []struct {
		name       string
		udp        stats.UDPInfo
		reachable  bool
		wantResult string
	}{
		{
			name:       "the reply carried our own payload back",
			udp:        stats.UDPInfo{Echoed: true, ReplySize: 8, ProbeNumber: 3},
			reachable:  true,
			wantResult: "echoed",
		},
		{
			name:       "something answered with a payload of its own",
			udp:        stats.UDPInfo{ReplySize: 40, ProbeNumber: 4},
			reachable:  true,
			wantResult: "replied",
		},
		{
			name:       "the target refused us",
			udp:        stats.UDPInfo{Rejected: true, ProbeNumber: 5},
			wantResult: "port unreachable",
		},
		{
			name:       "nothing came back at all",
			udp:        stats.UDPInfo{ProbeNumber: 6},
			wantResult: "no reply",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printer := dbTestPrinter(t, Config{})

			probeStats := dbTestStats()
			probeStats.Protocol = config.UDP
			probeStats.UDP = tt.udp

			if tt.reachable {
				printer.PrintProbeSuccess(probeStats)
			} else {
				printer.PrintProbeFailure(probeStats)
			}

			got := queryRow(t, printer, fmt.Sprintf(
				`SELECT udp_probe_number, udp_result FROM %s`, printer.probeTableName))

			wantNumber := strconv.FormatUint(tt.udp.ProbeNumber, 10)
			if got[0] != wantNumber {
				t.Errorf("udp_probe_number = %q, want %q", got[0], wantNumber)
			}
			if got[1] != tt.wantResult {
				t.Errorf("udp_result = %q, want %q", got[1], tt.wantResult)
			}
		})
	}
}

func TestDatabaseTimestampAndSourceAddressFollowTheConfig(t *testing.T) {
	t.Run("left empty unless asked for", func(t *testing.T) {
		printer := dbTestPrinter(t, Config{})

		probeStats := dbTestStats()
		probeStats.LocalAddr = &net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 12345}

		printer.PrintProbeSuccess(probeStats)

		got := queryRow(t, printer, fmt.Sprintf(
			`SELECT timestamp, source_address FROM %s`, printer.probeTableName))

		if got[0] != "" || got[1] != "" {
			t.Errorf("timestamp = %q and source_address = %q, want both empty", got[0], got[1])
		}
	})

	t.Run("written when asked for", func(t *testing.T) {
		printer := dbTestPrinter(t, Config{WithTimestamp: true, WithSourceAddress: true})

		probeStats := dbTestStats()
		probeStats.LocalAddr = &net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 12345}

		printer.PrintProbeSuccess(probeStats)

		got := queryRow(t, printer, fmt.Sprintf(
			`SELECT timestamp, source_address FROM %s`, printer.probeTableName))

		if got[0] == "" {
			t.Error("timestamp is empty, want the probe time")
		}
		if got[1] != "10.0.0.1:12345" {
			t.Errorf("source_address = %q, want 10.0.0.1:12345", got[1])
		}
	})
}

func TestDatabasePrintStatisticsRow(t *testing.T) {
	printer := dbTestPrinter(t, Config{})

	start := time.Now().Add(-time.Hour)

	probeStats := dbTestStats()
	probeStats.StartTime = start
	probeStats.EndTime = start.Add(time.Hour)
	probeStats.TotalSuccessfulProbes = 8
	probeStats.TotalUnsuccessfulProbes = 2
	probeStats.TotalUptime = 8 * time.Second
	probeStats.TotalDowntime = 2 * time.Second
	probeStats.RTTResults = stats.RTTResult{Min: 1.5, Average: 3.25, Max: 9, Mdev: 0.75}
	probeStats.RetriedHostnameLookups = 3

	// PrintStatistics reports to the terminal as well as the database.
	captureStdout(t, func() { printer.PrintStatistics(probeStats) })

	got := queryRow(t, printer, fmt.Sprintf(
		`SELECT hostname, ip_address, port, protocol, total_packets,
		        total_successful_packets, total_unsuccessful_packets,
		        total_packet_loss_percent, hostname_resolve_retries,
		        latency_min, latency_avg, latency_max, latency_mdev
		 FROM %s`, printer.statsTableName))

	want := []string{
		"example.com", "93.184.216.34", "443", "TCP", "10",
		"8", "2",
		"20.0", "3",
		"1.5", "3.25", "9.0", "0.75",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("column %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDatabaseStatisticsLeaveOutWhatDidNotHappen(t *testing.T) {
	printer := dbTestPrinter(t, Config{})

	// A run where every probe failed: there is no latency to summarize and
	// no uptime streak to point at.
	probeStats := dbTestStats()
	probeStats.OngoingSuccessfulProbes = 0
	probeStats.TotalSuccessfulProbes = 0
	probeStats.TotalUnsuccessfulProbes = 5

	captureStdout(t, func() { printer.PrintStatistics(probeStats) })

	got := queryRow(t, printer, fmt.Sprintf(
		`SELECT latency_min, latency_avg, latency_max, latency_mdev,
		        longest_uptime, longest_consecutive_uptime_start,
		        last_successful_probe, end_time
		 FROM %s`, printer.statsTableName))

	// The latency columns must be NULL rather than zero, which would read
	// as a very fast target.
	for i := range 4 {
		if got[i] != "null" {
			t.Errorf("latency column %d = %q, want null without a successful probe", i, got[i])
		}
	}

	for i := 4; i < len(got); i++ {
		if got[i] != "" {
			t.Errorf("column %d = %q, want it empty for something that never happened", i, got[i])
		}
	}
}

func TestDatabaseStatisticsRetriesAreLeftOutForAnIPTarget(t *testing.T) {
	printer := dbTestPrinter(t, Config{})

	// A literal IP is never resolved, so a retry count would be meaningless
	// even if the field happened to hold one.
	probeStats := dbTestStats()
	probeStats.DestIsIP = true
	probeStats.RetriedHostnameLookups = 7

	captureStdout(t, func() { printer.PrintStatistics(probeStats) })

	got := queryRow(t, printer, fmt.Sprintf(
		`SELECT hostname_resolve_retries FROM %s`, printer.statsTableName))

	if got[0] != "0" {
		t.Errorf("hostname_resolve_retries = %q, want 0 for an IP target", got[0])
	}
}

func TestDatabaseShutdownWritesStatisticsAndClosesTheDatabase(t *testing.T) {
	printer, err := NewDatabasePrinter(Config{Target: "example.com", Port: 443, OutputDBPath: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}

	out := captureStdout(t, func() { printer.Shutdown(dbTestStats()) })

	if !strings.Contains(out, printer.statsTableName) {
		t.Errorf("Shutdown output = %q, want it to name the statistics table", out)
	}

	// A second Close is what tells us the first one already happened.
	if err := printer.conn.Close(); err == nil {
		t.Error("the connection is still open after Shutdown")
	}
}

func TestUDPResultText(t *testing.T) {
	tests := []struct {
		name string
		udp  stats.UDPInfo
		want string
	}{
		{
			name: "an echo beats a plain reply",
			udp:  stats.UDPInfo{Echoed: true, ReplySize: 8},
			want: "echoed",
		},
		{
			name: "a reply that is not our payload",
			udp:  stats.UDPInfo{ReplySize: 40},
			want: "replied",
		},
		{
			name: "an ICMP refusal",
			udp:  stats.UDPInfo{Rejected: true},
			want: "port unreachable",
		},
		{
			name: "silence",
			udp:  stats.UDPInfo{},
			want: "no reply",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := udpResultText(&stats.Statistics{UDP: tt.udp}); got != tt.want {
				t.Errorf("udpResultText = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNullIfEmpty(t *testing.T) {
	if got := nullIfEmpty(""); got != nil {
		t.Errorf("nullIfEmpty(\"\") = %v, want nil", got)
	}
	if got := nullIfEmpty("TLS 1.3"); got != "TLS 1.3" {
		t.Errorf("nullIfEmpty(%q) = %v, want the string back", "TLS 1.3", got)
	}
}

func TestMillisecondsFloat(t *testing.T) {
	tests := []struct {
		name string
		in   time.Duration
		want float64
	}{
		{name: "whole milliseconds", in: 12 * time.Millisecond, want: 12},
		{name: "zero", in: 0, want: 0},
		{name: "kept to three decimals", in: 1234 * time.Microsecond, want: 1.234},
		{name: "rounded at the fourth decimal", in: 1234500 * time.Nanosecond, want: 1.235},
		{name: "sub-millisecond is not floored to zero", in: 500 * time.Microsecond, want: 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := millisecondsFloat(tt.in); got != tt.want {
				t.Errorf("millisecondsFloat(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestHostnameChanges(t *testing.T) {
	when := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	t.Run("a single entry is not a change", func(t *testing.T) {
		s := &stats.Statistics{HostnameChanges: []stats.HostnameChange{
			{Addr: netip.MustParseAddr("1.2.3.4"), When: when},
		}}

		if got := hostnameChanges(s); got != "" {
			t.Errorf("hostnameChanges = %q, want empty", got)
		}
	})

	t.Run("each pair becomes a line", func(t *testing.T) {
		s := &stats.Statistics{HostnameChanges: []stats.HostnameChange{
			{Addr: netip.MustParseAddr("1.2.3.4"), When: when},
			{Addr: netip.MustParseAddr("5.6.7.8"), When: when, Duration: 12 * time.Millisecond},
		}}

		got := hostnameChanges(s)
		if !strings.Contains(got, "from 1.2.3.4 to 5.6.7.8") {
			t.Errorf("hostnameChanges = %q, want it to name both addresses", got)
		}
	})
}
