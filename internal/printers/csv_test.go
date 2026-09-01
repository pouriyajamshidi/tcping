package printers

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
)

func TestAddCSVExtension(t *testing.T) {
	timestamp := strings.ReplaceAll(time.Now().Format(time.DateTime), ":", "-")
	timestamp = strings.ReplaceAll(timestamp, " ", "_")
	timestampWithExtension := "_" + timestamp + ".csv"
	timestampWithExtensionStats := "_" + timestamp + "_stats.csv"

	tests := []struct {
		name         string
		filename     string
		withStatsExt bool
		expected     string
	}{
		{"No extension, probe file", "results", false, "results" + timestampWithExtension},
		{"No extension, stats file", "results", true, "results" + timestampWithExtensionStats},
		{"Standard extension, probe file", "results.csv", false, "results" + timestampWithExtension},
		{"Standard extension, stats file", "results.csv", true, "results" + timestampWithExtensionStats},
		{"Multiple dots, probe file", "results.backup.2026", false, "results.backup.2026" + timestampWithExtension},
		{"Multiple dots, stats file", "results.backup.2026.csv", true, "results.backup.2026" + timestampWithExtensionStats},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addDateAndCSVExtension(tt.filename, tt.withStatsExt, false)
			fmt.Println(result)
			if result != tt.expected {
				t.Errorf("addCSVExtension(%q, %v) = %q; expected %q", tt.filename, tt.withStatsExt, result, tt.expected)
			}
		})
	}
}

func TestAddCSVExtension_NoTimestamp(t *testing.T) {
	tests := []struct {
		name         string
		filename     string
		withStatsExt bool
		expected     string
	}{
		{"No extension, probe file", "results", false, "results.csv"},
		{"No extension, stats file", "results", true, "results_stats.csv"},
		{"Standard extension, probe file", "results.csv", false, "results.csv"},
		{"Standard extension, stats file", "results.csv", true, "results_stats.csv"},
		{"Multiple dots, probe file", "results.backup.2026", false, "results.backup.2026.csv"},
		{"Multiple dots, stats file", "results.backup.2026.csv", true, "results.backup.2026_stats.csv"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addDateAndCSVExtension(tt.filename, tt.withStatsExt, true)
			if result != tt.expected {
				t.Errorf("addCSVExtension(%q, %v, true) = %q; expected %q", tt.filename, tt.withStatsExt, result, tt.expected)
			}
		})
	}
}

func TestNewCSVPrinter_CreatesFiles(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_output")

	p, err := NewCSVPrinter(filePath, false)
	if err != nil {
		t.Fatalf("NewCSVPrinter failed: %v", err)
	}
	defer p.Done()

	// Verify Probe File
	if _, err := os.Stat(p.ProbeFile.Name()); os.IsNotExist(err) {
		t.Errorf("Expected probe file to be created at %s, but it was not", p.ProbeFile.Name())
	}

	// Verify Stats File
	if _, err := os.Stat(p.StatsFile.Name()); os.IsNotExist(err) {
		t.Errorf("Expected stats file to be created at %s, but it was not", p.StatsFile.Name())
	}
}

func TestCSVPrinter_PrintStart_WritesHeaders(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "headers_test.csv")

	p, err := NewCSVPrinter(filePath, false)
	if err != nil {
		t.Fatalf("NewCSVPrinter failed: %v", err)
	}
	defer p.Done()

	// Dummy stats object for testing
	dummyStats := &stats.Statistics{
		Hostname:          "example.com",
		Port:              443,
		WithTimestamp:     true,
		WithSourceAddress: true,
	}

	p.PrintStart(dummyStats)
	p.Done() // Close and flush so we can read

	// Verify Probe CSV Headers
	probeFile, err := os.Open(p.ProbeFile.Name())
	if err != nil {
		t.Fatalf("Failed to open probe file: %v", err)
	}
	defer probeFile.Close()

	probeReader := csv.NewReader(probeFile)
	probeHeaders, err := probeReader.Read()
	if err != nil {
		t.Fatalf("Failed to read probe headers: %v", err)
	}

	expectedProbeHeaders := []string{colTimestamp, colStatus, colHostname, colIP, colPort, colSourceAddress, colConnection, colLatency}
	if len(probeHeaders) != len(expectedProbeHeaders) {
		t.Errorf("Expected %d probe headers, got %d", len(expectedProbeHeaders), len(probeHeaders))
	}

	// Verify Stats CSV Headers
	statsFile, err := os.Open(p.StatsFile.Name())
	if err != nil {
		t.Fatalf("Failed to open stats file: %v", err)
	}
	defer statsFile.Close()

	statsReader := csv.NewReader(statsFile)
	statsHeaders, err := statsReader.Read()
	if err != nil {
		t.Fatalf("Failed to read stats headers: %v", err)
	}

	expectedStatsHeaders := []string{"Metric", "Value"}
	if len(statsHeaders) != len(expectedStatsHeaders) {
		t.Errorf("Expected %d stats headers, got %d", len(expectedStatsHeaders), len(statsHeaders))
	}
}

func TestCSVPrinter_ProbeRecords(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "records_test.csv")

	p, err := NewCSVPrinter(filePath, false)
	if err != nil {
		t.Fatalf("NewCSVPrinter failed: %v", err)
	}
	defer p.Done()

	dummyStats := &stats.Statistics{
		Hostname: "example.com",
		Port:     80,
	}

	// Write one success and one failure
	p.PrintProbeSuccess(dummyStats)
	p.PrintProbeFailure(dummyStats)
	p.Done()

	// Read back the records
	probeFile, err := os.Open(p.ProbeFile.Name())
	if err != nil {
		t.Fatalf("Failed to open probe file: %v", err)
	}
	defer probeFile.Close()

	reader := csv.NewReader(probeFile)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read probe records: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(records))
	}

	// Basic validation of record contents
	if records[0][0] != "true" {
		t.Errorf("Expected first record to be a 'true', got %q", records[0][0])
	}
	if records[1][0] != "false" {
		t.Errorf("Expected second record to be a 'false', got %q", records[1][0])
	}
}
