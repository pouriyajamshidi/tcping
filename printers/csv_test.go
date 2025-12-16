package printers_test

import (
	"net/netip"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/printers"
	"github.com/pouriyajamshidi/tcping/v3/statistics"
)

func setupTempCSV(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	return filepath.Join(tempDir, "test_output")
}

func TestNewCSVPrinter(t *testing.T) {
	filePath := setupTempCSV(t)
	p, err := printers.NewCSVPrinter(filePath)
	if err != nil {
		t.Fatalf("NewCSVPrinter: %v", err)
	}
	if p == nil {
		t.Fatal("NewCSVPrinter returned nil")
	}

	defer p.Shutdown(&statistics.Statistics{})

	// verify probe file exists
	probeFile := filePath + ".csv"
	if _, err := os.Stat(probeFile); os.IsNotExist(err) {
		t.Errorf("probe file not created: %s", probeFile)
	}

	// verify stats file exists
	statsFile := filePath + "_stats.csv"
	if _, err := os.Stat(statsFile); os.IsNotExist(err) {
		t.Errorf("stats file not created: %s", statsFile)
	}
}

func TestNewCSVPrinter_WithOptions(t *testing.T) {
	filePath := setupTempCSV(t)
	opts := []printers.CSVPrinterOption{
		printers.WithTimestamp[*printers.CSVPrinter](),
		printers.WithSourceAddress[*printers.CSVPrinter](),
		printers.WithFailuresOnly[*printers.CSVPrinter](),
	}
	p, err := printers.NewCSVPrinter(filePath, opts...)
	if err != nil {
		t.Fatalf("NewCSVPrinter with options: %v", err)
	}
	if p == nil {
		t.Fatal("NewCSVPrinter with options returned nil")
	}
	defer p.Shutdown(&statistics.Statistics{})
}

func TestNewCSVPrinter_InvalidPath(t *testing.T) {
	_, err := printers.NewCSVPrinter("/invalid/path/that/does/not/exist/file")
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}

func TestCSVPrinter_PrintStart(t *testing.T) {
	filePath := setupTempCSV(t)
	p, err := printers.NewCSVPrinter(filePath)
	if err != nil {
		t.Fatalf("NewCSVPrinter: %v", err)
	}
	defer p.Shutdown(&statistics.Statistics{})

	stats := &statistics.Statistics{
		Hostname: "example.com",
		Port:     443,
	}

	// smoke test - should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintStart panicked: %v", r)
		}
	}()
	p.PrintStart(stats)
}

func TestCSVPrinter_PrintProbeSuccess(t *testing.T) {
	filePath := setupTempCSV(t)
	p, err := printers.NewCSVPrinter(filePath)
	if err != nil {
		t.Fatalf("NewCSVPrinter: %v", err)
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
	p.PrintStart(stats)
	p.PrintProbeSuccess(stats)

	// verify file has content
	probeFile := filePath + ".csv"
	info, err := os.Stat(probeFile)
	if err != nil {
		t.Errorf("probe file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("probe file is empty")
	}
}

func TestCSVPrinter_PrintProbeFailure(t *testing.T) {
	filePath := setupTempCSV(t)
	p, err := printers.NewCSVPrinter(filePath)
	if err != nil {
		t.Fatalf("NewCSVPrinter: %v", err)
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
	p.PrintStart(stats)
	p.PrintProbeFailure(stats)

	// verify file has content
	probeFile := filePath + ".csv"
	info, err := os.Stat(probeFile)
	if err != nil {
		t.Errorf("probe file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("probe file is empty")
	}
}

func TestCSVPrinter_PrintStatistics(t *testing.T) {
	filePath := setupTempCSV(t)
	p, err := printers.NewCSVPrinter(filePath)
	if err != nil {
		t.Fatalf("NewCSVPrinter: %v", err)
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
	p.PrintStart(stats)
	p.PrintStatistics(stats)
	p.Shutdown(stats)

	// verify stats file has content
	statsFile := filePath + "_stats.csv"
	info, err := os.Stat(statsFile)
	if err != nil {
		t.Errorf("stats file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("stats file is empty")
	}
}

func TestCSVPrinter_Shutdown(t *testing.T) {
	filePath := setupTempCSV(t)
	p, err := printers.NewCSVPrinter(filePath)
	if err != nil {
		t.Fatalf("NewCSVPrinter: %v", err)
	}

	stats := &statistics.Statistics{}

	// smoke test - should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Shutdown panicked: %v", r)
		}
	}()

	p.Shutdown(stats)

	// verify files exist and are readable after shutdown
	probeFile := filePath + ".csv"
	if _, err := os.Stat(probeFile); os.IsNotExist(err) {
		t.Errorf("probe file not found after shutdown: %s", probeFile)
	}

	statsFile := filePath + "_stats.csv"
	if _, err := os.Stat(statsFile); os.IsNotExist(err) {
		t.Errorf("stats file not found after shutdown: %s", statsFile)
	}
}
