// Package printers contains the logic for printing information
package printers

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/utils"
	"github.com/pouriyajamshidi/tcping/v3/probes/statistics"
)

const (
	colTimestamp     string = "Timestamp"
	colStatus        string = "Status"
	colHostname      string = "Hostname"
	colIP            string = "IP"
	colPort          string = "Port"
	colConnection    string = "Connection"
	colLatency       string = "Latency(ms)"
	colSourceAddress string = "Source Address"
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
		return nil, fmt.Errorf("Error creating the probe CSV file %s: %w", probeFilename, err)
	}

	statsFilename := addCSVExtension(filePath, true)

	statsFile, err := os.OpenFile(statsFilename, fileFlag, filePermission)
	if err != nil {
		return nil, fmt.Errorf("Error creating the probe CSV file %s: %w", statsFilename, err)
	}

	p := &CSVPrinter{
		ProbeWriter: csv.NewWriter(probeFile),
		StatsWriter: csv.NewWriter(statsFile),
		ProbeFile:   probeFile,
		StatsFile:   statsFile,
	}

	return p, nil
}

func addCSVExtension(filename string, withStatsExt bool) string {
	if withStatsExt {
		// TODO: account for when there are more than one dots
		return strings.Split(filename, ".")[0] + "_stats.csv"
	}

	if strings.HasSuffix(filename, ".csv") {
		return filename
	}

	return filename + ".csv"
}

// Done flushes the buffer of writers and closes the probe and stats file
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

// Shutdown sets the end time, prints statistics, calls Done() and exits the program.
func (p *CSVPrinter) Shutdown(s *statistics.Statistics) {
	s.EndTime = time.Now()
	PrintStats(p, s)
	p.Done()
	os.Exit(0)
}

func (p *CSVPrinter) writeProbeHeader(s *statistics.Statistics) error {
	headers := []string{}

	if s.WithTimestamp {
		headers = append(headers, colTimestamp)
	}

	headers = append(headers, colStatus, colHostname, colIP, colPort)

	if s.WithSourceAddress {
		headers = append(headers, colSourceAddress)
	}

	headers = append(headers, colConnection, colLatency)

	if err := p.ProbeWriter.Write(headers); err != nil {
		return fmt.Errorf("Failed to write headers: %w", err)
	}

	p.ProbeWriter.Flush()

	return p.ProbeWriter.Error()
}

func (p *CSVPrinter) writeStatsHeader() error {
	headers := []string{
		"Metric",
		"Value",
	}

	if err := p.ProbeWriter.Write(headers); err != nil {
		return fmt.Errorf("Failed to write statistics headers: %w", err)
	}

	p.ProbeWriter.Flush()

	return p.ProbeWriter.Error()
}

// PrintStart logs the beginning of a TCPing session.
func (p *CSVPrinter) PrintStart(s *statistics.Statistics) {
	// TODO: Is this a good place to put these?
	p.writeProbeHeader(s)
	p.writeStatsHeader()

	fmt.Printf("TCPinging %s on port %d - saving the results to: %s\n", s.Hostname, s.Port, p.ProbeFile.Name())
}

// PrintProbeSuccess logs a successful probe to the CSV file.
func (p *CSVPrinter) PrintProbeSuccess(s *statistics.Statistics) {
	record := []string{}

	if s.WithTimestamp {
		record = append(record, s.StartTimeFormatted())
	}

	record = append(
		record,
		"Reply",
		s.Hostname,
		s.IP.String(),
		fmt.Sprint(s.Port),
	)

	if s.WithSourceAddress {
		// TODO: Is there a better way than Sprint?
		record = append(record, s.SourceAddr(), fmt.Sprint(s.OngoingSuccessfulProbes), s.RTTStr())
	}

	// TODO: Is there a better way than Sprint?
	record = append(record, fmt.Sprint(s.OngoingSuccessfulProbes), s.RTTStr())

	if err := p.ProbeWriter.Write(record); err != nil {
		p.PrintError("Failed to write success record: %w", err)
	}

	p.ProbeWriter.Flush()
}

// PrintProbeFailure logs a failed probe attempt to the CSV file.
func (p *CSVPrinter) PrintProbeFailure(s *statistics.Statistics) {
	record := []string{}

	if s.WithTimestamp {
		record = append(record, s.StartTimeFormatted())
	}

	record = append(
		record,
		"No Reply",
		s.Hostname,
		s.IP.String(),
		fmt.Sprint(s.Port),
		fmt.Sprint(s.OngoingUnsuccessfulProbes),
	)

	if err := p.ProbeWriter.Write(record); err != nil {
		p.PrintError("Failed to write failure record: %v", err)
	}

	p.ProbeWriter.Flush()
}

// PrintError logs an error message to stderr.
func (p *CSVPrinter) PrintError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "CSV Error: "+format+"\n", args...)
}

// PrintRetryingToResolve logs an attempt to resolve a hostname.
func (p *CSVPrinter) PrintRetryingToResolve(s *statistics.Statistics) {
	fmt.Printf("Retrying to resolve %s\n", s.Hostname)
}

// PrintStatistics logs TCPing statistics to a CSV file.
func (p *CSVPrinter) PrintStatistics(s *statistics.Statistics) {
	timestamp := time.Now().Format(time.DateTime)

	statistics := [][]string{
		{"Timestamp", timestamp},
		{"IP Address", s.IPStr()},
	}

	if s.IPStr() != s.Hostname {
		statistics = append(statistics, []string{"Hostname", s.Hostname})
	}

	statistics = append(statistics, []string{"Port", fmt.Sprintf("%d", s.Port)})

	totalDuration := s.TotalDowntime + s.TotalUptime
	statistics = append(statistics, []string{"Total Duration",
		fmt.Sprintf("%.0f", totalDuration.Seconds())},
	)

	statistics = append(statistics, []string{"Total Uptime",
		utils.DurationToString(s.TotalUptime)},
	)
	statistics = append(statistics, []string{"Total Downtime",
		utils.DurationToString(s.TotalDowntime)},
	)

	totalPackets := s.TotalSuccessfulProbes + s.TotalUnsuccessfulProbes
	packetLoss := (float32(s.TotalUnsuccessfulProbes) / float32(totalPackets)) * 100

	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	statistics = append(statistics, []string{"Total Packets", fmt.Sprintf("%d", totalPackets)})
	statistics = append(statistics, []string{"Total Successful Packets", fmt.Sprintf("%d", s.TotalSuccessfulProbes)})
	statistics = append(statistics, []string{"Total Unsuccessful Packets", fmt.Sprintf("%d", s.TotalUnsuccessfulProbes)})
	statistics = append(statistics, []string{"Total Packet Loss Percentage", fmt.Sprintf("%.2f", packetLoss)})

	if s.LongestUp.Duration != 0 {
		longestUptime := fmt.Sprintf("%.0f", s.LongestUp.Duration.Seconds())
		longestConsecutiveUptimeStart := s.LongestUp.Start.Format(time.DateTime)
		longestConsecutiveUptimeEnd := s.LongestUp.End.Format(time.DateTime)

		statistics = append(statistics, []string{"Longest Uptime", longestUptime})
		statistics = append(statistics, []string{"Longest Consecutive Uptime Start", longestConsecutiveUptimeStart})
		statistics = append(statistics, []string{"Longest Consecutive Uptime End", longestConsecutiveUptimeEnd})
	} else {
		statistics = append(statistics, []string{"Longest Uptime", "Never"})
		statistics = append(statistics, []string{"Longest Consecutive Uptime Start", "Never"})
		statistics = append(statistics, []string{"Longest Consecutive Uptime End", "Never"})
	}

	if s.LongestDown.Duration != 0 {
		longestDowntime := fmt.Sprintf("%.0f", s.LongestDown.Duration.Seconds())
		longestConsecutiveDowntimeStart := s.LongestDown.Start.Format(time.DateTime)
		longestConsecutiveDowntimeEnd := s.LongestDown.End.Format(time.DateTime)

		statistics = append(statistics, []string{"Longest Downtime", longestDowntime})
		statistics = append(statistics, []string{"Longest Consecutive Downtime Start", longestConsecutiveDowntimeStart})
		statistics = append(statistics, []string{"Longest Consecutive Downtime End", longestConsecutiveDowntimeEnd})
	} else {
		statistics = append(statistics, []string{"Longest Downtime", "Never"})
		statistics = append(statistics, []string{"Longest Consecutive Downtime Start", "Never"})
		statistics = append(statistics, []string{"Longest Consecutive Downtime End", "Never"})
	}

	if s.RetriedHostnameLookups > 0 {
		statistics = append(statistics, []string{"Hostname Resolve Retries", fmt.Sprintf("%d", s.RetriedHostnameLookups)})
	}

	if len(s.HostnameChanges) > 1 {
		hostnameChanges := ""

		for i := 0; i < len(s.HostnameChanges)-1; i++ {
			if s.HostnameChanges[i].Addr.String() == "" {
				continue
			}

			hostnameChanges += fmt.Sprintf("from %s to %s at %v - ",
				s.HostnameChanges[i].Addr.String(),
				s.HostnameChanges[i+1].Addr.String(),
				s.HostnameChanges[i+1].When.Format(time.DateTime),
			)
		}
	} else {
		statistics = append(statistics, []string{"Hostname Changes", "Never changed"})
	}

	if s.LastSuccessfulProbe.IsZero() {
		statistics = append(statistics, []string{"Last Successful Probe", "Never succeeded"})
	} else {
		statistics = append(statistics, []string{"Last Successful Probe", s.LastSuccessfulProbe.Format(time.DateTime)})
	}

	if s.LastUnsuccessfulProbe.IsZero() {
		statistics = append(statistics, []string{"Last Unsuccessful Probe", "Never failed"})
	} else {
		statistics = append(statistics, []string{"Last Unsuccessful Probe", s.LastUnsuccessfulProbe.Format(time.DateTime)})
	}

	if s.RTTResults.HasResults {
		statistics = append(statistics, []string{"Latency Min", fmt.Sprintf("%.3f", s.RTTResults.Min)})
		statistics = append(statistics, []string{"Latency Avg", fmt.Sprintf("%.3f", s.RTTResults.Average)})
		statistics = append(statistics, []string{"Latency Max", fmt.Sprintf("%.3f", s.RTTResults.Max)})
	} else {
		statistics = append(statistics, []string{"Latency Min", "N/A"})
		statistics = append(statistics, []string{"Latency Avg", "N/A"})
		statistics = append(statistics, []string{"Latency Max", "N/A"})
	}

	statistics = append(statistics, []string{"Start Timestamp", s.StartTime.Format(time.DateTime)})

	if !s.EndTime.IsZero() {
		statistics = append(statistics, []string{"End Timestamp", s.EndTime.Format(time.DateTime)})
	} else {
		statistics = append(statistics, []string{"End Timestamp", "In progress"})
	}

	for _, record := range statistics {
		if err := p.StatsWriter.Write(record); err != nil {
			p.PrintError("Failed to write statistics record: %v", err)
			return
		}
	}

	fmt.Printf("\nStatistics have been saved to: %s\n", p.StatsFile.Name())
}

// PrintTotalDownTime is a no-op implementation to satisfy the Printer interface.
func (p *CSVPrinter) PrintTotalDownTime(_ *statistics.Statistics) {}
