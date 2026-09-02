package printers

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
)

const (
	colTimestamp     string = "Timestamp"
	colStatus        string = "Reachable"
	colHostname      string = "Hostname"
	colIP            string = "IP"
	colPort          string = "Port"
	colSourceAddress string = "Source Address"
	colConnection    string = "Connection"
	colLatency       string = "Latency(ms)"
)

// Extra columns written for HTTP(S) targets only, so a TCP run keeps the
// same CSV shape it has always had.
const (
	colStatusCode  string = "Status Code"
	colHTTPVersion string = "HTTP Version"
	colTLSVersion  string = "TLS Version"
	colCertExpiry  string = "Certificate Expiry"
	colConnectTime string = "Connect(ms)"
	colTLSTime     string = "TLS Handshake(ms)"
	colTTFB        string = "TTFB(ms)"
)

// httpColumns are the HTTP(S) header names in the order httpRecord fills them.
var httpColumns = []string{
	colStatusCode,
	colHTTPVersion,
	colTLSVersion,
	colCertExpiry,
	colConnectTime,
	colTLSTime,
	colTTFB,
}

// httpRecord returns the HTTP(S) values for one probe, in httpColumns order.
// A probe that got no response leaves them all empty.
func httpRecord(s *stats.Statistics) []string {
	if !s.HasHTTPResponse() {
		return make([]string, len(httpColumns))
	}

	return []string{
		s.StatusCodeStr(),
		s.HTTP.Proto,
		s.HTTP.TLSVersion,
		s.CertExpiryStr(),
		s.ConnectDurationStr(),
		s.TLSDurationStr(),
		s.TimeToFirstByteStr(),
	}
}

// Extra columns written for UDP targets only, for the same reason: a TCP run
// keeps the CSV shape it has always had.
const (
	colUDPProbeNumber string = "UDP Probe"
	colUDPResult      string = "UDP Result"
)

// udpResult describes what one UDP probe learned, since "Reachable" alone
// cannot tell a refusal apart from silence.
func udpResult(s *stats.Statistics) string {
	switch {
	case s.UDP.Echoed:
		return "echoed"
	case s.UDP.ReplySize > 0:
		return "replied"
	case s.UDP.Rejected:
		return "port unreachable"
	default:
		return "no reply"
	}
}

const (
	filePermission os.FileMode = 0644
	fileFlag       int         = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
)

// CSVPrinter is responsible for writing probe results and statistics to CSV files.
type CSVPrinter struct {
	probeWriter *csv.Writer
	statsWriter *csv.Writer
	probeFile   *os.File
	statsFile   *os.File
	cfg         config.PrinterConfig
}

// NewCSVPrinter initializes a CSVPrinter instance from the printer config.
// When cfg.CSVNoTimestamp is true, cfg.OutputCSVPath is used as-is (plus a
// "_stats" suffix for the stats file) instead of getting a date/time suffix,
// so repeated runs overwrite the same file rather than creating a new one
// each time.
func NewCSVPrinter(cfg config.PrinterConfig) (*CSVPrinter, error) {
	filePath, noTimestamp := cfg.OutputCSVPath, cfg.CSVNoTimestamp

	probeFilename := addDateAndCSVExtension(filePath, false, noTimestamp)

	probeFile, err := os.OpenFile(probeFilename, fileFlag, filePermission)
	if err != nil {
		return nil, fmt.Errorf("error creating the probe CSV file %s: %w", probeFilename, err)
	}

	statsFilename := addDateAndCSVExtension(filePath, true, noTimestamp)

	statsFile, err := os.OpenFile(statsFilename, fileFlag, filePermission)
	if err != nil {
		probeFile.Close()
		return nil, fmt.Errorf("error creating the stats CSV file %s: %w", statsFilename, err)
	}

	return &CSVPrinter{
		probeWriter: csv.NewWriter(probeFile),
		statsWriter: csv.NewWriter(statsFile),
		probeFile:   probeFile,
		statsFile:   statsFile,
		cfg:         cfg,
	}, nil
}

func addDateAndCSVExtension(filename string, withStatsExt, noTimestamp bool) string {
	ext := filepath.Ext(filename)
	// don't mistake example.com with example.csv
	if ext != ".csv" {
		ext = ""
	}
	base := filename[:len(filename)-len(ext)]

	if noTimestamp {
		if withStatsExt {
			return base + "_stats.csv"
		}
		return base + ".csv"
	}

	timestamp := strings.ReplaceAll(time.Now().Format(time.DateTime), ":", "-")

	if withStatsExt {
		return strings.ReplaceAll(base+"_"+timestamp+"_stats.csv", " ", "_")
	}

	return strings.ReplaceAll(base+"_"+timestamp+".csv", " ", "_")
}

// Done flushes the buffer of writers and closes the probe and stats files.
func (p *CSVPrinter) done() {
	if p.probeWriter != nil {
		p.probeWriter.Flush()
	}
	if p.probeFile != nil {
		p.probeFile.Close()
	}

	if p.statsWriter != nil {
		p.statsWriter.Flush()
	}
	if p.statsFile != nil {
		p.statsFile.Close()
	}
}

func (p *CSVPrinter) writeProbeHeader(s *stats.Statistics) error {
	var headers []string

	if p.cfg.WithTimestamp {
		headers = append(headers, colTimestamp)
	}

	headers = append(headers, colStatus, colHostname, colIP, colPort)

	if p.cfg.WithSourceAddress {
		headers = append(headers, colSourceAddress)
	}

	headers = append(headers, colConnection, colLatency)

	if s.IsHTTP() {
		headers = append(headers, httpColumns...)
	}

	if s.IsUDP() {
		headers = append(headers, colUDPProbeNumber, colUDPResult)
	}

	if err := p.probeWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	p.probeWriter.Flush()
	return p.probeWriter.Error()
}

func (p *CSVPrinter) writeStatsHeader() error {
	headers := []string{"Statistic", "Value"}

	if err := p.statsWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write statistics headers: %w", err)
	}

	p.statsWriter.Flush()
	return p.statsWriter.Error()
}

// PrintStart logs the beginning of a TCPing session.
func (p *CSVPrinter) PrintStart(s *stats.Statistics) {
	p.writeProbeHeader(s)
	p.writeStatsHeader()

	if s.DestIsIP {
		fmt.Printf("Probing %s on port %d over %s - saving the results to: %s\n", s.Hostname, s.Port, s.ProtocolStr(), p.probeFile.Name())
		return
	}
	fmt.Printf("Probing %s on port %d over %s (resolved in %s ms) - saving the results to: %s\n",
		s.Hostname, s.Port, s.ProtocolStr(), s.NameResolutionDurationStr(), p.probeFile.Name())
}

// PrintNameResolutionDuration prints how long a hostname resolution retry took.
func (p *CSVPrinter) PrintNameResolutionDuration(s *stats.Statistics) {
	fmt.Printf("Resolved in %s ms\n", s.NameResolutionDurationStr())
}

// PrintProbeSuccess logs a successful probe to the CSV file.
func (p *CSVPrinter) PrintProbeSuccess(s *stats.Statistics) {
	var record []string

	if p.cfg.WithTimestamp {
		record = append(record, s.CurrentTimestamp())
	}

	record = append(
		record,
		"true",
		s.Hostname,
		s.IPStr(),
		s.PortStr(),
	)

	if p.cfg.WithSourceAddress {
		record = append(record, s.SourceAddr())
	}

	record = append(record, strconv.Itoa(int(s.OngoingSuccessfulProbes)), s.RTTStr())

	if s.IsHTTP() {
		record = append(record, httpRecord(s)...)
	}

	if s.IsUDP() {
		record = append(record, s.ProbeNumberStr(), udpResult(s))
	}

	if err := p.probeWriter.Write(record); err != nil {
		p.PrintError("failed to write success record: %v", err)
	}

	p.probeWriter.Flush()
}

// PrintProbeFailure logs a failed probe attempt to the CSV file.
func (p *CSVPrinter) PrintProbeFailure(s *stats.Statistics) {
	var record []string

	if p.cfg.WithTimestamp {
		record = append(record, s.CurrentTimestamp())
	}

	record = append(
		record,
		"false",
		s.Hostname,
		s.IPStr(),
		s.PortStr(),
	)

	// Always write the column when the header has one, even if the address
	// is empty, otherwise the row is short and every later column shifts.
	if p.cfg.WithSourceAddress {
		record = append(record, s.SourceAddr())
	}

	record = append(record, strconv.Itoa(int(s.OngoingUnsuccessfulProbes)), "")

	if s.IsHTTP() {
		record = append(record, httpRecord(s)...)
	}

	if s.IsUDP() {
		record = append(record, s.ProbeNumberStr(), udpResult(s))
	}

	if err := p.probeWriter.Write(record); err != nil {
		p.PrintError("failed to write failure record: %v", err)
	}

	p.probeWriter.Flush()
}

// PrintStatistics logs TCPing statistics to a CSV file.
func (p *CSVPrinter) PrintStatistics(s *stats.Statistics) {
	statistics := [][]string{
		{"IP Address", s.IPStr()},
	}

	if !s.DestIsIP {
		statistics = append(statistics, []string{"Hostname", s.Hostname})
	}

	statistics = append(statistics,
		[]string{"Port", s.PortStr()},
		[]string{"Protocol", s.ProtocolStr()},
		[]string{"Total Duration", s.RuntimeDuration()},
		[]string{"Total Uptime", s.TotalUptimeDuration()},
		[]string{"Total Downtime", s.TotalDowntimeDuration()},
		[]string{"Total Packets", strconv.Itoa(int(s.TotalProbes()))},
		[]string{"Total Successful Packets", strconv.Itoa(int(s.TotalSuccessfulProbes))},
		[]string{"Total Unsuccessful Packets", strconv.Itoa(int(s.TotalUnsuccessfulProbes))},
		[]string{"Total Packet Loss Percentage", fmt.Sprintf("%.2f", s.PacketLoss())},
	)

	if s.LongestUptime.Duration != 0 {
		statistics = append(statistics,
			[]string{"Longest Uptime", s.LongestUptimeDuration()},
			[]string{"Longest Consecutive Uptime Start", s.LongestUptimeStartTime()},
			[]string{"Longest Consecutive Uptime End", s.LongestUptimeEndTime()},
		)
	} else {
		statistics = append(statistics,
			[]string{"Longest Uptime", "0"},
			[]string{"Longest Consecutive Uptime Start", ""},
			[]string{"Longest Consecutive Uptime End", ""},
		)
	}

	if s.LongestDowntime.Duration != 0 {
		statistics = append(statistics,
			[]string{"Longest Downtime", s.LongestDowntimeDuration()},
			[]string{"Longest Consecutive Downtime Start", s.LongestDowntimeStartTime()},
			[]string{"Longest Consecutive Downtime End", s.LongestDowntimeEndTime()},
		)
	} else {
		statistics = append(statistics,
			[]string{"Longest Downtime", "0"},
			[]string{"Longest Consecutive Downtime Start", ""},
			[]string{"Longest Consecutive Downtime End", ""},
		)
	}

	if s.RetriedHostnameLookups > 0 {
		statistics = append(statistics, []string{"Hostname Resolve Retries", strconv.Itoa(int(s.RetriedHostnameLookups))})
	}

	if len(s.HostnameChanges) > 1 {
		var hostnameChanges strings.Builder

		for i := 0; i < len(s.HostnameChanges)-1; i++ {
			from := s.HostnameChanges[i].Addr.String()
			if from == "" {
				continue
			}

			to := s.HostnameChanges[i+1].Addr.String()

			fmt.Fprintf(&hostnameChanges, "from %s to %s at %s took %s ms - ",
				from,
				to,
				s.HostnameChanges[i+1].WhenFormatted(),
				s.HostnameChanges[i+1].DurationStr(),
			)
		}
		statistics = append(statistics, []string{"Hostname Changes", hostnameChanges.String()})
	} else {
		statistics = append(statistics, []string{"Hostname Changes", "0"})
	}

	if s.LastSuccessfulProbe.IsZero() {
		statistics = append(statistics, []string{"Last Successful Probe", ""})
	} else {
		statistics = append(statistics, []string{"Last Successful Probe", s.LastSuccessfulProbeFormatted()})
	}

	if s.LastUnsuccessfulProbe.IsZero() {
		statistics = append(statistics, []string{"Last Unsuccessful Probe", ""})
	} else {
		statistics = append(statistics, []string{"Last Unsuccessful Probe", s.LastUnsuccessfulProbeFormatted()})
	}

	if s.TotalSuccessfulProbes > 0 {
		statistics = append(statistics,
			[]string{"Latency Min", fmt.Sprintf("%.3f", s.RTTResults.Min)},
			[]string{"Latency Avg", fmt.Sprintf("%.3f", s.RTTResults.Average)},
			[]string{"Latency Max", fmt.Sprintf("%.3f", s.RTTResults.Max)},
			[]string{"Latency Mdev", fmt.Sprintf("%.3f", s.RTTResults.Mdev)},
		)
	} else {
		statistics = append(statistics,
			[]string{"Latency Min", ""},
			[]string{"Latency Avg", ""},
			[]string{"Latency Max", ""},
			[]string{"Latency Mdev", ""},
		)
	}

	statistics = append(statistics, []string{"Start Timestamp", s.StartTimeFormatted()})

	if !s.EndTime.IsZero() {
		statistics = append(statistics, []string{"End Timestamp", s.EndTimeFormatted()})
	} else {
		statistics = append(statistics, []string{"End Timestamp", ""})
	}

	for _, record := range statistics {
		if err := p.statsWriter.Write(record); err != nil {
			p.PrintError("failed to write statistics record: %v", err)
			return
		}
	}

	p.statsWriter.Flush()

	fmt.Printf("\nStatistics have been saved to: %s\n", p.statsFile.Name())
}

// PrintRetryingToResolve logs an attempt to resolve a hostname.
func (p *CSVPrinter) PrintRetryingToResolve(hostname string) {
	fmt.Printf("Retrying to resolve %s\n", hostname)
}

// PrintDownTimeDuration prints the total duration of downtime when no response was received.
func (p *CSVPrinter) PrintDownTimeDuration(s *stats.Statistics) {
	fmt.Printf("No response received for %s\n", s.DowntimeDuration())
}

// PrintUpTimeDuration prints how long the target was up for, right as it stops responding.
func (p *CSVPrinter) PrintUpTimeDuration(s *stats.Statistics) {
	fmt.Printf("No response received after %s of uptime\n", s.UptimeDuration())
}

// PrintError logs an error message to stderr.
func (p *CSVPrinter) PrintError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "CSV Error: "+format+"\n", args...)
}

// Shutdown prints statistics and calls Done() to flush and close the CSV
// files. Statistics are already finalized by finalizeStatistics by the time
// this runs. It does not exit the program - that decision belongs to the
// caller, not the printer.
func (p *CSVPrinter) Shutdown(s *stats.Statistics) {
	p.PrintStatistics(s)
	p.done()
}
