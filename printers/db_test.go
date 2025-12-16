package printers_test

import (
	"net/netip"
	"path/filepath"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/printers"
	"github.com/pouriyajamshidi/tcping/v3/statistics"
)

func setupTempDB(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	return filepath.Join(tempDir, "test.db")
}

func TestNewDatabasePrinter(t *testing.T) {
	dbPath := setupTempDB(t)
	p, err := printers.NewDatabasePrinter("example.com", "443", dbPath)
	if err != nil {
		t.Fatalf("NewDatabasePrinter: %v", err)
	}
	if p == nil {
		t.Fatal("NewDatabasePrinter returned nil")
	}
	defer p.Shutdown(&statistics.Statistics{})
}

func TestNewDatabasePrinter_WithOptions(t *testing.T) {
	dbPath := setupTempDB(t)
	opts := []printers.DatabasePrinterOption{
		printers.WithTimestamp[*printers.DatabasePrinter](),
		printers.WithSourceAddress[*printers.DatabasePrinter](),
		printers.WithFailuresOnly[*printers.DatabasePrinter](),
	}
	p, err := printers.NewDatabasePrinter("example.com", "443", dbPath, opts...)
	if err != nil {
		t.Fatalf("NewDatabasePrinter with options: %v", err)
	}
	if p == nil {
		t.Fatal("NewDatabasePrinter with options returned nil")
	}
	defer p.Shutdown(&statistics.Statistics{})
}

func TestNewDatabasePrinter_TableNameSanitization(t *testing.T) {
	tests := []struct {
		name   string
		target string
		port   string
	}{
		{
			name:   "hostname with dots",
			target: "example.com",
			port:   "443",
		},
		{
			name:   "IP address",
			target: "192.168.1.1",
			port:   "8080",
		},
		{
			name:   "hostname with special chars",
			target: "test-server.example.com",
			port:   "443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath := setupTempDB(t)
			p, err := printers.NewDatabasePrinter(tt.target, tt.port, dbPath)
			if err != nil {
				t.Fatalf("NewDatabasePrinter: %v", err)
			}
			defer p.Shutdown(&statistics.Statistics{})

			// smoke test - database should be created without errors
			if p == nil {
				t.Fatal("NewDatabasePrinter returned nil")
			}
		})
	}
}

func TestDatabasePrinter_PrintProbeSuccess(t *testing.T) {
	dbPath := setupTempDB(t)
	p, err := printers.NewDatabasePrinter("example.com", "443", dbPath)
	if err != nil {
		t.Fatalf("NewDatabasePrinter: %v", err)
	}
	defer p.Shutdown(&statistics.Statistics{})

	stats := &statistics.Statistics{
		IP:                      netip.MustParseAddr("192.168.1.1"),
		Hostname:                "example.com",
		Port:                    443,
		OngoingSuccessfulProbes: 5,
		LatestRTT:               12.345,
	}

	// smoke test - should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintProbeSuccess panicked: %v", r)
		}
	}()
	p.PrintProbeSuccess(stats)
}

func TestDatabasePrinter_PrintProbeFailure(t *testing.T) {
	dbPath := setupTempDB(t)
	p, err := printers.NewDatabasePrinter("example.com", "443", dbPath)
	if err != nil {
		t.Fatalf("NewDatabasePrinter: %v", err)
	}
	defer p.Shutdown(&statistics.Statistics{})

	stats := &statistics.Statistics{
		IP:                        netip.MustParseAddr("192.168.1.1"),
		Hostname:                  "example.com",
		Port:                      443,
		OngoingUnsuccessfulProbes: 3,
	}

	// smoke test - should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintProbeFailure panicked: %v", r)
		}
	}()
	p.PrintProbeFailure(stats)
}

func TestDatabasePrinter_PrintStatistics(t *testing.T) {
	dbPath := setupTempDB(t)
	p, err := printers.NewDatabasePrinter("example.com", "443", dbPath)
	if err != nil {
		t.Fatalf("NewDatabasePrinter: %v", err)
	}

	stats := &statistics.Statistics{
		IP:                        netip.MustParseAddr("192.168.1.1"),
		Hostname:                  "example.com",
		Port:                      443,
		TotalSuccessfulProbes:     10,
		TotalUnsuccessfulProbes:   2,
		LastSuccessfulProbe:       time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		LastUnsuccessfulProbe:     time.Date(2024, 1, 15, 12, 0, 30, 0, time.UTC),
		TotalUptime:               50 * time.Second,
		TotalDowntime:             10 * time.Second,
		StartTime:                 time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		EndTime:                   time.Date(2024, 1, 15, 12, 1, 0, 0, time.UTC),
		RTTResults:                statistics.RttResult{HasResults: true, Min: 10.5, Average: 15.2, Max: 20.8},
	}

	// smoke test - should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintStatistics panicked: %v", r)
		}
	}()
	p.PrintStatistics(stats)
	p.Shutdown(stats)
}

func TestDatabasePrinter_Shutdown(t *testing.T) {
	dbPath := setupTempDB(t)
	p, err := printers.NewDatabasePrinter("example.com", "443", dbPath)
	if err != nil {
		t.Fatalf("NewDatabasePrinter: %v", err)
	}

	stats := &statistics.Statistics{}

	// smoke test - should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Shutdown panicked: %v", r)
		}
	}()

	p.Shutdown(stats)
}
