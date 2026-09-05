package printers

import (
	"encoding/csv"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/config"
	"github.com/pouriyajamshidi/tcping/v3/stats"
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

	p, err := NewCSVPrinter(Config{OutputCSVPath: filePath})
	if err != nil {
		t.Fatalf("NewCSVPrinter failed: %v", err)
	}
	defer p.done()

	// Verify Probe File
	if _, err := os.Stat(p.probeFile.Name()); os.IsNotExist(err) {
		t.Errorf("Expected probe file to be created at %s, but it was not", p.probeFile.Name())
	}

	// Verify Stats File
	if _, err := os.Stat(p.statsFile.Name()); os.IsNotExist(err) {
		t.Errorf("Expected stats file to be created at %s, but it was not", p.statsFile.Name())
	}
}

func TestCSVPrinter_PrintStart_WritesHeaders(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "headers_test.csv")

	p, err := NewCSVPrinter(Config{
		OutputCSVPath:     filePath,
		WithTimestamp:     true,
		WithSourceAddress: true,
	})
	if err != nil {
		t.Fatalf("NewCSVPrinter failed: %v", err)
	}
	defer p.done()

	// Dummy stats object for testing
	dummyStats := &stats.Statistics{
		Hostname: "example.com",
		Port:     443,
	}

	p.PrintStart(dummyStats)
	p.done() // Close and flush so we can read

	// Verify Probe CSV Headers
	probeFile, err := os.Open(p.probeFile.Name())
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
	statsFile, err := os.Open(p.statsFile.Name())
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

	p, err := NewCSVPrinter(Config{OutputCSVPath: filePath})
	if err != nil {
		t.Fatalf("NewCSVPrinter failed: %v", err)
	}
	defer p.done()

	dummyStats := &stats.Statistics{
		Hostname: "example.com",
		Port:     80,
	}

	// Write one success and one failure
	p.PrintProbeSuccess(dummyStats)
	p.PrintProbeFailure(dummyStats)
	p.done()

	// Read back the records
	probeFile, err := os.Open(p.probeFile.Name())
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

// A failed probe usually has no source address to report. The column is
// still in the header, so the row has to carry an empty value for it,
// otherwise every column after it shifts left by one.
func TestCSVPrinter_FailureRowKeepsSourceAddressColumn(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "failure_alignment.csv")

	p, err := NewCSVPrinter(Config{
		OutputCSVPath:     filePath,
		WithSourceAddress: true,
	})
	if err != nil {
		t.Fatalf("NewCSVPrinter failed: %v", err)
	}
	defer p.done()

	s := &stats.Statistics{
		Hostname:                  "example.com",
		Port:                      443,
		OngoingUnsuccessfulProbes: 1,
	}

	p.PrintStart(s)
	p.PrintProbeFailure(s)
	p.done()

	probeFile, err := os.Open(p.probeFile.Name())
	if err != nil {
		t.Fatalf("Failed to open probe file: %v", err)
	}
	defer probeFile.Close()

	// csv.Reader defaults to requiring every record to match the header's
	// field count, so a short row fails right here.
	records, err := csv.NewReader(probeFile).ReadAll()
	if err != nil {
		t.Fatalf("Failed to read probe CSV: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("got %d rows, want 2 (header + one probe)", len(records))
	}

	header, row := records[0], records[1]
	if len(row) != len(header) {
		t.Fatalf("row has %d fields, header has %d", len(row), len(header))
	}

	for i, name := range header {
		if name != colSourceAddress {
			continue
		}
		if row[i] != "" {
			t.Errorf("%s = %q, want empty", colSourceAddress, row[i])
		}
	}
}

// csvTestPrinter is a CSV printer writing into a temporary directory.
func csvTestPrinter(t *testing.T, cfg Config) *CSVPrinter {
	t.Helper()

	cfg.OutputCSVPath = filepath.Join(t.TempDir(), "results.csv")
	cfg.CSVNoTimestamp = true

	p, err := NewCSVPrinter(cfg)
	if err != nil {
		t.Fatalf("NewCSVPrinter failed: %v", err)
	}
	t.Cleanup(p.done)

	return p
}

// readCSV flushes the printer and reads one of its files back.
func readCSV(t *testing.T, p *CSVPrinter, path string) [][]string {
	t.Helper()

	p.done()

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("opening %s: %v", path, err)
	}
	defer f.Close()

	// The probe rows and the statistics rows have different widths, so let
	// the reader accept whatever each file holds.
	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}

	return records
}

// csvTestStats is a successful TCP probe against a hostname. Tests change
// only the fields they care about.
func csvTestStats() *stats.Statistics {
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

func TestCSVHTTPProbeColumns(t *testing.T) {
	p := csvTestPrinter(t, Config{})

	probeStats := csvTestStats()
	probeStats.Protocol = config.HTTPS
	probeStats.HTTP = stats.HTTPInfo{
		StatusCode:      200,
		Proto:           "HTTP/2.0",
		TLSVersion:      "TLS 1.3",
		CertExpiry:      time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC),
		ConnectDuration: 12 * time.Millisecond,
		TLSDuration:     34 * time.Millisecond,
		TimeToFirstByte: 56 * time.Millisecond,
	}

	p.PrintStart(probeStats)
	p.PrintProbeSuccess(probeStats)

	records := readCSV(t, p, p.probeFile.Name())
	if len(records) != 2 {
		t.Fatalf("got %d rows, want a header and one probe", len(records))
	}

	headers, row := records[0], records[1]

	// The HTTP columns are appended after the common ones, so line them up
	// by name rather than by a hard-coded index.
	got := make(map[string]string, len(headers))
	for i, name := range headers {
		got[name] = row[i]
	}

	want := map[string]string{
		colStatusCode:  "200",
		colHTTPVersion: "HTTP/2.0",
		colTLSVersion:  "TLS 1.3",
		colConnectTime: "12.000",
		colTLSTime:     "34.000",
		colTTFB:        "56.000",
	}
	for name, value := range want {
		if got[name] != value {
			t.Errorf("column %q = %q, want %q", name, got[name], value)
		}
	}

	if got[colCertExpiry] == "" {
		t.Errorf("column %q is empty, want the certificate expiry", colCertExpiry)
	}
}

func TestCSVTCPProbeHasNoHTTPOrUDPColumns(t *testing.T) {
	p := csvTestPrinter(t, Config{})

	p.PrintStart(csvTestStats())
	p.PrintProbeSuccess(csvTestStats())

	records := readCSV(t, p, p.probeFile.Name())
	headers := records[0]

	// A TCP run keeps the CSV shape it has always had.
	unwanted := append([]string{colUDPProbeNumber, colUDPResult}, httpColumns...)
	for _, name := range unwanted {
		if slices.Contains(headers, name) {
			t.Errorf("headers %q contain %q, want a TCP run to leave it out", headers, name)
		}
	}
}

func TestCSVHTTPProbeWithoutAResponseLeavesItsColumnsEmpty(t *testing.T) {
	p := csvTestPrinter(t, Config{})

	probeStats := csvTestStats()
	probeStats.Protocol = config.HTTPS
	probeStats.OngoingSuccessfulProbes = 0
	probeStats.OngoingUnsuccessfulProbes = 1

	p.PrintStart(probeStats)
	p.PrintProbeFailure(probeStats)

	records := readCSV(t, p, p.probeFile.Name())
	headers, row := records[0], records[1]

	if len(row) != len(headers) {
		t.Fatalf("the row has %d fields but the header has %d, so later columns are shifted", len(row), len(headers))
	}

	for i, name := range headers {
		if slices.Contains(httpColumns, name) && row[i] != "" {
			t.Errorf("column %q = %q, want it empty when nothing answered", name, row[i])
		}
	}
}

func TestCSVUDPProbeColumns(t *testing.T) {
	tests := []struct {
		name      string
		udp       stats.UDPInfo
		reachable bool
		want      string
	}{
		{
			name:      "the reply carried our own payload back",
			udp:       stats.UDPInfo{Echoed: true, ReplySize: 8, ProbeNumber: 3},
			reachable: true,
			want:      "echoed",
		},
		{
			name:      "something answered with a payload of its own",
			udp:       stats.UDPInfo{ReplySize: 40, ProbeNumber: 4},
			reachable: true,
			want:      "replied",
		},
		{
			name: "the target refused us",
			udp:  stats.UDPInfo{Rejected: true, ProbeNumber: 5},
			want: "port unreachable",
		},
		{
			name: "nothing came back at all",
			udp:  stats.UDPInfo{ProbeNumber: 6},
			want: "no reply",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := csvTestPrinter(t, Config{})

			probeStats := csvTestStats()
			probeStats.Protocol = config.UDP
			probeStats.UDP = tt.udp

			p.PrintStart(probeStats)
			if tt.reachable {
				p.PrintProbeSuccess(probeStats)
			} else {
				p.PrintProbeFailure(probeStats)
			}

			records := readCSV(t, p, p.probeFile.Name())
			headers, row := records[0], records[1]

			got := make(map[string]string, len(headers))
			for i, name := range headers {
				got[name] = row[i]
			}

			wantNumber := strconv.FormatUint(tt.udp.ProbeNumber, 10)
			if got[colUDPProbeNumber] != wantNumber {
				t.Errorf("column %q = %q, want %q", colUDPProbeNumber, got[colUDPProbeNumber], wantNumber)
			}
			if got[colUDPResult] != tt.want {
				t.Errorf("column %q = %q, want %q", colUDPResult, got[colUDPResult], tt.want)
			}
		})
	}
}

func TestUDPResult(t *testing.T) {
	tests := []struct {
		name string
		udp  stats.UDPInfo
		want string
	}{
		{name: "an echo beats a plain reply", udp: stats.UDPInfo{Echoed: true, ReplySize: 8}, want: "echoed"},
		{name: "a reply that is not our payload", udp: stats.UDPInfo{ReplySize: 40}, want: "replied"},
		{name: "an ICMP refusal", udp: stats.UDPInfo{Rejected: true}, want: "port unreachable"},
		{name: "silence", udp: stats.UDPInfo{}, want: "no reply"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := udpResult(&stats.Statistics{UDP: tt.udp}); got != tt.want {
				t.Errorf("udpResult = %q, want %q", got, tt.want)
			}
		})
	}
}

// statsRows turns the statistics file into a name/value map, since it is
// written as one "Statistic,Value" pair per line.
func statsRows(t *testing.T, p *CSVPrinter) map[string]string {
	t.Helper()

	records := readCSV(t, p, p.statsFile.Name())

	if len(records) == 0 || records[0][0] != "Statistic" {
		t.Fatalf("the statistics file does not start with its header: %q", records)
	}

	rows := make(map[string]string, len(records))
	for _, record := range records[1:] {
		if len(record) != 2 {
			t.Fatalf("statistics row %q has %d fields, want 2", record, len(record))
		}
		rows[record[0]] = record[1]
	}

	return rows
}

func TestCSVStatisticsFile(t *testing.T) {
	p := csvTestPrinter(t, Config{})

	start := time.Now().Add(-time.Hour)

	probeStats := csvTestStats()
	probeStats.StartTime = start
	probeStats.EndTime = start.Add(time.Hour)
	probeStats.TotalSuccessfulProbes = 8
	probeStats.TotalUnsuccessfulProbes = 2
	probeStats.RTTResults = stats.RTTResult{Min: 1.5, Average: 3.25, Max: 9, Mdev: 0.75}
	probeStats.RetriedHostnameLookups = 3

	p.PrintStart(probeStats)
	// PrintStatistics reports to the terminal as well as the file.
	captureStdout(t, func() { p.PrintStatistics(probeStats) })

	rows := statsRows(t, p)

	want := map[string]string{
		"IP Address":                   "93.184.216.34",
		"Hostname":                     "example.com",
		"Port":                         "443",
		"Protocol":                     "TCP",
		"Total Packets":                "10",
		"Total Successful Packets":     "8",
		"Total Unsuccessful Packets":   "2",
		"Total Packet Loss Percentage": "20.00",
		"Hostname Resolve Retries":     "3",
		"Latency Min":                  "1.500",
		"Latency Avg":                  "3.250",
		"Latency Max":                  "9.000",
		"Latency Mdev":                 "0.750",
	}
	for name, value := range want {
		if rows[name] != value {
			t.Errorf("statistic %q = %q, want %q", name, rows[name], value)
		}
	}
}

func TestCSVStatisticsLeaveOutWhatDidNotHappen(t *testing.T) {
	p := csvTestPrinter(t, Config{})

	// A run where every probe failed, and against a literal IP so there is
	// no hostname to report either.
	probeStats := csvTestStats()
	probeStats.DestIsIP = true
	probeStats.OngoingSuccessfulProbes = 0
	probeStats.TotalSuccessfulProbes = 0
	probeStats.TotalUnsuccessfulProbes = 5

	p.PrintStart(probeStats)
	captureStdout(t, func() { p.PrintStatistics(probeStats) })

	rows := statsRows(t, p)

	if _, ok := rows["Hostname"]; ok {
		t.Error("the statistics name a hostname for an IP target")
	}

	if _, ok := rows["Hostname Resolve Retries"]; ok {
		t.Error("the statistics report resolve retries with none recorded")
	}

	// Without a successful probe there is no latency to summarize, and a
	// zero would read as a very fast target.
	for _, name := range []string{"Latency Min", "Latency Avg", "Latency Max", "Latency Mdev"} {
		if rows[name] != "" {
			t.Errorf("statistic %q = %q, want it empty without a successful probe", name, rows[name])
		}
	}

	for _, name := range []string{"Last Successful Probe", "End Timestamp", "Longest Consecutive Uptime Start"} {
		if rows[name] != "" {
			t.Errorf("statistic %q = %q, want it empty for something that never happened", name, rows[name])
		}
	}

	if rows["Longest Uptime"] != "0" {
		t.Errorf("statistic %q = %q, want 0", "Longest Uptime", rows["Longest Uptime"])
	}
	if rows["Hostname Changes"] != "0" {
		t.Errorf("statistic %q = %q, want 0", "Hostname Changes", rows["Hostname Changes"])
	}
}

func TestCSVStatisticsReportTheLongestStreaks(t *testing.T) {
	p := csvTestPrinter(t, Config{})

	start := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	probeStats := csvTestStats()
	probeStats.TotalSuccessfulProbes = 8
	probeStats.LongestUptime = stats.LongestTime{
		Start: start, End: start.Add(90 * time.Second), Duration: 90 * time.Second,
	}
	probeStats.LongestDowntime = stats.LongestTime{
		Start: start, End: start.Add(30 * time.Second), Duration: 30 * time.Second,
	}

	p.PrintStart(probeStats)
	captureStdout(t, func() { p.PrintStatistics(probeStats) })

	rows := statsRows(t, p)

	if rows["Longest Uptime"] != "1 minute 30 seconds" {
		t.Errorf("statistic %q = %q, want %q", "Longest Uptime", rows["Longest Uptime"], "1 minute 30 seconds")
	}
	if rows["Longest Downtime"] != "30 seconds" {
		t.Errorf("statistic %q = %q, want %q", "Longest Downtime", rows["Longest Downtime"], "30 seconds")
	}

	for _, name := range []string{
		"Longest Consecutive Uptime Start", "Longest Consecutive Uptime End",
		"Longest Consecutive Downtime Start", "Longest Consecutive Downtime End",
	} {
		if rows[name] == "" {
			t.Errorf("statistic %q is empty, want a timestamp", name)
		}
	}
}

func TestCSVStatisticsReportHostnameChanges(t *testing.T) {
	p := csvTestPrinter(t, Config{})

	when := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	probeStats := csvTestStats()
	probeStats.HostnameChanges = []stats.HostnameChange{
		{Addr: netip.MustParseAddr("1.2.3.4"), When: when},
		{Addr: netip.MustParseAddr("5.6.7.8"), When: when, Duration: 12 * time.Millisecond},
	}

	p.PrintStart(probeStats)
	captureStdout(t, func() { p.PrintStatistics(probeStats) })

	rows := statsRows(t, p)

	if !strings.Contains(rows["Hostname Changes"], "from 1.2.3.4 to 5.6.7.8") {
		t.Errorf("statistic %q = %q, want it to name both addresses", "Hostname Changes", rows["Hostname Changes"])
	}
}

func TestCSVShutdownWritesTheStatisticsFile(t *testing.T) {
	p := csvTestPrinter(t, Config{})

	probeStats := csvTestStats()
	p.PrintStart(probeStats)

	out := captureStdout(t, func() { p.Shutdown(probeStats) })

	if !strings.Contains(out, p.statsFile.Name()) {
		t.Errorf("Shutdown output = %q, want it to name the statistics file", out)
	}

	rows := statsRows(t, p)
	if rows["IP Address"] != "93.184.216.34" {
		t.Errorf("the statistics file was not written by Shutdown: %q", rows)
	}
}
