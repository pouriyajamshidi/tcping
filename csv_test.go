package main

import (
	"encoding/csv"
	"net/netip"
	"os"
	"reflect"
	"testing"
	"time"
)

func setupTempCSVFile(t *testing.T) (*csvPrinter, func()) {
	t.Helper()
	tempFile, err := os.CreateTemp("", "test_csv_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	cp := &csvPrinter{
		file:   tempFile,
		writer: csv.NewWriter(tempFile),
	}

	cleanup := func() {
		cp.writer.Flush()
		if err := cp.writer.Error(); err != nil {
			t.Errorf("Error flushing CSV writer: %v", err)
		}
		if err := cp.file.Close(); err != nil {
			t.Errorf("Error closing file: %v", err)
		}
		if err := os.Remove(cp.file.Name()); err != nil {
			t.Errorf("Error removing temp file: %v", err)
		}
	}

	return cp, cleanup
}

func TestNewCSVPrinter(t *testing.T) {
	args := []string{"localhost", "8001"}
	filename := "test.csv"

	cp, err := newCSVPrinter(filename, args)
	if err != nil {
		t.Fatalf("error creating CSV printer: %v", err)
	}

	defer func() {
		if cp != nil && cp.file != nil {
			cp.file.Close()
		}
		os.Remove(filename)
	}()

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("file %s was not created", filename)
	}

	if cp.filename != filename {
		t.Errorf("expected filename %q, got %q", filename, cp.filename)
	}

	if cp.headerDone {
		t.Errorf("expected headerDone to be false, got true")
	}

	if cp.writer == nil {
		t.Errorf("CSV writer was not initialized")
	}
}

func TestCSVPrinterWriteHeader(t *testing.T) {
	cp, cleanup := setupTempCSVFile(t)
	defer cleanup()

	err := cp.writeHeader()
	if err != nil {
		t.Fatalf("writeHeader() failed: %v", err)
	}

	cp.writer.Flush()
	if err := cp.writer.Error(); err != nil {
		t.Fatalf("Error flushing CSV writer: %v", err)
	}

	cp.file.Seek(0, 0)
	reader := csv.NewReader(cp.file)
	header, err := reader.Read()
	if err != nil {
		t.Fatalf("Failed to read CSV header: %v", err)
	}

	expectedHeader := []string{
		"Event Type", "Timestamp", "Address", "Hostname", "Port",
		"Hostname Resolve Retries", "Total Successful Probes", "Total Unsuccessful Probes",
		"Never Succeed Probe", "Never Failed Probe", "Last Successful Probe",
		"Last Unsuccessful Probe", "Total Packets", "Total Packet Loss",
		"Total Uptime", "Total Downtime", "Longest Uptime", "Longest Uptime Start",
		"Longest Uptime End", "Longest Downtime", "Longest Downtime Start",
		"Longest Downtime End", "Latency Min", "Latency Avg", "Latency Max",
		"Start Time", "End Time", "Total Duration",
	}

	if !reflect.DeepEqual(header, expectedHeader) {
		t.Errorf("Header mismatch.\nExpected: %v\nGot: %v", expectedHeader, header)
	}
}

func TestCSVPrinterSaveStats(t *testing.T) {
	cp, cleanup := setupTempCSVFile(t)
	defer cleanup()

	now := time.Now()
	sampleTcping := tcping{
		userInput: userInput{
			ip:       netip.MustParseAddr("192.168.1.1"),
			hostname: "example.com",
			port:     80,
		},
		startTime:               now.Add(-time.Hour),
		endTime:                 now,
		lastSuccessfulProbe:     now.Add(-time.Minute),
		lastUnsuccessfulProbe:   now,
		longestUptime:           longestTime{duration: 50 * time.Second, start: now.Add(-time.Hour), end: now.Add(-59 * time.Minute)},
		longestDowntime:         longestTime{duration: 3 * time.Second, start: now.Add(-30 * time.Minute), end: now.Add(-29*time.Minute - 57*time.Second)},
		totalUptime:             55 * time.Minute,
		totalDowntime:           5 * time.Minute,
		totalSuccessfulProbes:   95,
		totalUnsuccessfulProbes: 5,
		retriedHostnameLookups:  2,
		rttResults:              rttResult{min: 10.5, max: 20.1, average: 15.3, hasResults: true},
	}

	err := cp.saveStats(sampleTcping)
	if err != nil {
		t.Fatalf("saveStats() failed: %v", err)
	}

	cp.writer.Flush()
	if err := cp.writer.Error(); err != nil {
		t.Fatalf("Error flushing CSV writer: %v", err)
	}

	cp.file.Seek(0, 0)
	reader := csv.NewReader(cp.file)

	header, err := reader.Read()
	if err != nil {
		t.Fatalf("Failed to read CSV header: %v", err)
	}
	expectedHeader := []string{
		"Event Type", "Timestamp", "Address", "Hostname", "Port",
		"Hostname Resolve Retries", "Total Successful Probes", "Total Unsuccessful Probes",
		"Never Succeed Probe", "Never Failed Probe", "Last Successful Probe",
		"Last Unsuccessful Probe", "Total Packets", "Total Packet Loss",
		"Total Uptime", "Total Downtime", "Longest Uptime", "Longest Uptime Start",
		"Longest Uptime End", "Longest Downtime", "Longest Downtime Start",
		"Longest Downtime End", "Latency Min", "Latency Avg", "Latency Max",
		"Start Time", "End Time", "Total Duration",
	}
	if !reflect.DeepEqual(header, expectedHeader) {
		t.Errorf("Header mismatch.\nExpected: %v\nGot: %v", expectedHeader, header)
	}

	row, err := reader.Read()
	if err != nil {
		t.Fatalf("Failed to read CSV data row: %v", err)
	}

	if row[0] != "statistics" {
		t.Errorf("Expected event type 'statistics', got '%s'", row[0])
	}
	if row[2] != "192.168.1.1" {
		t.Errorf("Expected IP '192.168.1.1', got '%s'", row[2])
	}
	if row[3] != "example.com" {
		t.Errorf("Expected hostname 'example.com', got '%s'", row[3])
	}
	if row[4] != "80" {
		t.Errorf("Expected port '80', got '%s'", row[4])
	}
	if row[6] != "95" {
		t.Errorf("Expected 95 successful probes, got '%s'", row[6])
	}
	if row[7] != "5" {
		t.Errorf("Expected 5 unsuccessful probes, got '%s'", row[7])
	}
}
func TestCSVPrinterSaveHostNameChange(t *testing.T) {
	cp, cleanup := setupTempCSVFile(t)
	defer cleanup()

	testCases := []struct {
		name     string
		changes  []hostnameChange
		expected int
	}{
		{
			name:     "Empty slice",
			changes:  []hostnameChange{},
			expected: 0,
		},
		{
			name: "Valid entries",
			changes: []hostnameChange{
				{Addr: netip.MustParseAddr("192.168.1.1"), When: time.Now()},
				{Addr: netip.MustParseAddr("192.168.1.2"), When: time.Now().Add(time.Hour)},
			},
			expected: 2,
		},
		{
			name: "Mixed valid and invalid entries",
			changes: []hostnameChange{
				{Addr: netip.MustParseAddr("192.168.1.1"), When: time.Now()},
				{Addr: netip.Addr{}, When: time.Now()},
				{Addr: netip.MustParseAddr("192.168.1.2"), When: time.Now().Add(time.Hour)},
			},
			expected: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset file for each test case
			cp.file.Truncate(0)
			cp.file.Seek(0, 0)

			err := cp.saveHostNameChange(tc.changes)
			if err != nil {
				t.Fatalf("saveHostNameChange failed: %v", err)
			}

			cp.writer.Flush()
			cp.file.Seek(0, 0)

			reader := csv.NewReader(cp.file)
			rows, err := reader.ReadAll()
			if err != nil {
				t.Fatalf("Failed to read CSV: %v", err)
			}

			if len(rows) != tc.expected {
				t.Errorf("Expected %d rows, got %d", tc.expected, len(rows))
			}

			for _, row := range rows {
				if row[0] != "hostname change" {
					t.Errorf("Expected 'hostname change', got %s", row[0])
				}
				if row[2] == "invalid IP" {
					t.Errorf("Invalid IP should not be written to CSV")
				}
			}
		})
	}
}
