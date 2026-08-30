package printers

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

const (
	filePermission os.FileMode = 0644
	fileFlag       int         = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
)

// CSVPrinter is responsible for writing probe results and statistics to CSV files.
type CSVPrinter struct {
	ProbeWriter *csv.Writer
	StatsWriter *csv.Writer
	ProbeFile   *os.File
	StatsFile   *os.File
}

// NewCSVPrinter initializes a CSVPrinter instance with the given filename and settings.
func NewCSVPrinter(filePath string) (*CSVPrinter, error) {
	probeFilename := addCSVExtension(filePath, false)

	probeFile, err := os.OpenFile(probeFilename, fileFlag, filePermission)
	if err != nil {
		return nil, fmt.Errorf("error creating the probe CSV file %s: %w", probeFilename, err)
	}

	statsFilename := addCSVExtension(filePath, true)

	statsFile, err := os.OpenFile(statsFilename, fileFlag, filePermission)
	if err != nil {
		probeFile.Close() // Clean up on failure
		return nil, fmt.Errorf("error creating the stats CSV file %s: %w", statsFilename, err)
	}

	return &CSVPrinter{
		ProbeWriter: csv.NewWriter(probeFile),
		StatsWriter: csv.NewWriter(statsFile),
		ProbeFile:   probeFile,
		StatsFile:   statsFile,
	}, nil
}

func addCSVExtension(filename string, withStatsExt bool) string {
	ext := filepath.Ext(filename)
	base := filename[:len(filename)-len(ext)]

	if withStatsExt {
		return base + "_stats.csv"
	}

	return base + ".csv"
}

// Done flushes the buffer of writers and closes the probe and stats files.
func (p *CSVPrinter) Done() {
	if p.ProbeWriter != nil {
		p.ProbeWriter.Flush()
	}
	if p.ProbeFile != nil {
		p.ProbeFile.Close()
	}

	if p.StatsWriter != nil {
		p.StatsWriter.Flush()
	}
	if p.StatsFile != nil {
		p.StatsFile.Close()
	}
}

func (p *CSVPrinter) writeProbeHeader(s *stats.Statistics) error {
	var headers []string

	if s.WithTimestamp {
		headers = append(headers, colTimestamp)
	}

	headers = append(headers, colStatus, colHostname, colIP, colPort)

	if s.WithSourceAddress {
		headers = append(headers, colSourceAddress)
	}

	headers = append(headers, colConnection, colLatency)

	if err := p.ProbeWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	p.ProbeWriter.Flush()
	return p.ProbeWriter.Error()
}

func (p *CSVPrinter) writeStatsHeader() error {
	headers := []string{"Statistic", "Value"}

	if err := p.StatsWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write statistics headers: %w", err)
	}

	p.StatsWriter.Flush()
	return p.StatsWriter.Error()
}

// PrintStart logs the beginning of a TCPing session.
func (p *CSVPrinter) PrintStart(s *stats.Statistics) {
	p.writeProbeHeader(s)
	p.writeStatsHeader()

	fmt.Printf("TCPinging %s on port %d - saving the results to: %s\n", s.Hostname, s.Port, p.ProbeFile.Name())
}

// PrintProbeSuccess logs a successful probe to the CSV file.
func (p *CSVPrinter) PrintProbeSuccess(s *stats.Statistics) {
	var record []string

	if s.WithTimestamp {
		record = append(record, s.CurrentTimestamp())
	}

	record = append(
		record,
		"true",
		s.Hostname,
		s.IPStr(),
		s.PortStr(),
	)

	if s.WithSourceAddress {
		record = append(record, s.SourceAddr())
	}

	record = append(record, strconv.Itoa(int(s.OngoingSuccessfulProbes)), s.RTTStr())

	if err := p.ProbeWriter.Write(record); err != nil {
		p.PrintError("failed to write success record: %w", err)
	}

	p.ProbeWriter.Flush()
}

// PrintProbeFailure logs a failed probe attempt to the CSV file.
func (p *CSVPrinter) PrintProbeFailure(s *stats.Statistics) {
	var record []string

	if s.WithTimestamp {
		record = append(record, s.CurrentTimestamp())
	}

	record = append(
		record,
		"false",
		s.Hostname,
		s.IPStr(),
		strconv.Itoa(int(s.Port)),
	)

	if s.WithSourceAddress {
		record = append(record, s.SourceAddr())
	}

	record = append(record, strconv.Itoa(int(s.OngoingUnsuccessfulProbes)), "")

	if err := p.ProbeWriter.Write(record); err != nil {
		p.PrintError("failed to write failure record: %v", err)
	}

	p.ProbeWriter.Flush()
}

// PrintStatistics logs TCPing statistics to a CSV file.
func (p *CSVPrinter) PrintStatistics(s *stats.Statistics) {
	timestamp := s.CurrentTimestamp()

	statistics := [][]string{
		{"Timestamp", timestamp},
		{"IP Address", s.IPStr()},
	}

	if !s.DestIsIP {
		statistics = append(statistics, []string{"Hostname", s.Hostname})
	}

	statistics = append(statistics,
		[]string{"Port", s.PortStr()},
		[]string{"Total Duration", s.RuntimeDuration()},
		[]string{"Total Uptime", s.TotalUptimeDuration()},
		[]string{"Total Downtime", s.TotalDowntimeDuration()},
		[]string{"Total Packets", strconv.Itoa(int(s.TotalProbes()))},
		[]string{"Total Successful Packets", strconv.Itoa(int(s.TotalSuccessfulProbes))},
		[]string{"Total Unsuccessful Packets", strconv.Itoa(int(s.TotalUnsuccessfulProbes))},
		[]string{"Total Packet Loss Percentage", fmt.Sprintf("%.2f", s.PacketLoss())},
	)

	if s.LongestUp.Duration != 0 {
		statistics = append(statistics,
			[]string{"Longest Uptime", s.LongestUptimeDuration()},
			[]string{"Longest Consecutive Uptime Start", s.LongestUptimeStartTime()},
			[]string{"Longest Consecutive Uptime End", s.LongestUptimeEndTime()},
		)
	} else {
		statistics = append(statistics,
			[]string{"Longest Uptime", "Never"},
			[]string{"Longest Consecutive Uptime Start", "Never"},
			[]string{"Longest Consecutive Uptime End", "Never"},
		)
	}

	if s.LongestDown.Duration != 0 {
		statistics = append(statistics,
			[]string{"Longest Downtime", s.LongestDowntimeDuration()},
			[]string{"Longest Consecutive Downtime Start", s.LongestDowntimeStartTime()},
			[]string{"Longest Consecutive Downtime End", s.LongestDowntimeEndTime()},
		)
	} else {
		statistics = append(statistics,
			[]string{"Longest Downtime", "Never"},
			[]string{"Longest Consecutive Downtime Start", "Never"},
			[]string{"Longest Consecutive Downtime End", "Never"},
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

			fmt.Fprintf(&hostnameChanges, "from %s to %s at %s - ",
				from,
				to,
				s.HostnameChanges[i+1].WhenFormatted(),
			)
		}
		statistics = append(statistics, []string{"Hostname Changes", hostnameChanges.String()})
	} else {
		statistics = append(statistics, []string{"Hostname Changes", "0"})
	}

	if s.LastSuccessfulProbe.IsZero() {
		statistics = append(statistics, []string{"Last Successful Probe", "Never succeeded"})
	} else {
		statistics = append(statistics, []string{"Last Successful Probe", s.LastSuccessfulProbeFormatted()})
	}

	if s.LastUnsuccessfulProbe.IsZero() {
		statistics = append(statistics, []string{"Last Unsuccessful Probe", "Never failed"})
	} else {
		statistics = append(statistics, []string{"Last Unsuccessful Probe", s.LastUnsuccessfulProbeFormatted()})
	}

	if s.RTTResults.HasResults {
		statistics = append(statistics,
			[]string{"Latency Min", fmt.Sprintf("%.3f", s.RTTResults.Min)},
			[]string{"Latency Avg", fmt.Sprintf("%.3f", s.RTTResults.Average)},
			[]string{"Latency Max", fmt.Sprintf("%.3f", s.RTTResults.Max)},
		)
	} else {
		statistics = append(statistics,
			[]string{"Latency Min", "N/A"},
			[]string{"Latency Avg", "N/A"},
			[]string{"Latency Max", "N/A"},
		)
	}

	statistics = append(statistics, []string{"Start Timestamp", s.StartTimeFormatted()})

	if !s.EndTime.IsZero() {
		statistics = append(statistics, []string{"End Timestamp", s.EndTimeFormatted()})
	} else {
		statistics = append(statistics, []string{"End Timestamp", "In progress"})
	}

	for _, record := range statistics {
		if err := p.StatsWriter.Write(record); err != nil {
			p.PrintError("failed to write statistics record: %v", err)
			return
		}
	}

	p.StatsWriter.Flush()

	fmt.Printf("\nStatistics have been saved to: %s\n", p.StatsFile.Name())
}

// PrintRetryingToResolve logs an attempt to resolve a hostname.
func (p *CSVPrinter) PrintRetryingToResolve(hostname string) {
	fmt.Printf("Retrying to resolve %s\n", hostname)
}

// PrintDownTimeDuration prints the total duration of downtime when no response was received.
func (p *CSVPrinter) PrintDownTimeDuration(s *stats.Statistics) {
	fmt.Printf("No response received for %s\n", s.DowntimeDuration())
}

// PrintError logs an error message to stderr.
func (p *CSVPrinter) PrintError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "CSV Error: "+format+"\n", args...)
}

// Shutdown sets the end time, prints statistics, calls Done() and exits the program.
func (p *CSVPrinter) Shutdown(s *stats.Statistics) {
	s.EndTime = time.Now()
	PrintStats(p, s)
	p.Done()
	os.Exit(0)
}
