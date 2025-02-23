// Package printers contains the logic for printing information
package printers

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/internal/utils"
	"github.com/pouriyajamshidi/tcping/v2/types"
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
	cfg         PrinterConfig
}

// NewCSVPrinter initializes a CSVPrinter instance with the given filename and settings.
func NewCSVPrinter(cfg PrinterConfig) (*CSVPrinter, error) {
	probeFilename := addCSVExtension(cfg.OutputCSVPath, false)

	probeFile, err := os.OpenFile(probeFilename, fileFlag, filePermission)
	if err != nil {
		return nil, fmt.Errorf("Error creating the probe CSV file %s: %w", probeFilename, err)
	}

	statsFilename := addCSVExtension(cfg.OutputCSVPath, true)

	statsFile, err := os.OpenFile(statsFilename, fileFlag, filePermission)
	if err != nil {
		return nil, fmt.Errorf("Error creating the probe CSV file %s: %w", statsFilename, err)
	}

	p := &CSVPrinter{
		ProbeWriter: csv.NewWriter(probeFile),
		StatsWriter: csv.NewWriter(statsFile),
		ProbeFile:   probeFile,
		StatsFile:   statsFile,
		cfg:         cfg,
	}

	writeProbeHeader(p.cfg, p.ProbeWriter)
	writeStatsHeader(p.StatsWriter)

	return p, nil
}

func addCSVExtension(filename string, withStatsExt bool) string {
	if withStatsExt {
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

func writeProbeHeader(cfg PrinterConfig, w *csv.Writer) error {
	headers := []string{}

	if cfg.WithTimestamp {
		headers = append(headers, colTimestamp)
	}

	headers = append(headers, colStatus, colHostname, colIP, colPort)

	if cfg.WithSourceAddress {
		headers = append(headers, colSourceAddress)
	}

	headers = append(headers, colConnection, colLatency)

	if err := w.Write(headers); err != nil {
		return fmt.Errorf("Failed to write headers: %w", err)
	}

	w.Flush()

	return w.Error()
}

func writeStatsHeader(w *csv.Writer) error {
	headers := []string{
		"Metric",
		"Value",
	}

	if err := w.Write(headers); err != nil {
		return fmt.Errorf("Failed to write statistics headers: %w", err)
	}

	w.Flush()

	return w.Error()
}

// PrintStart logs the beginning of a TCPing session.
func (p *CSVPrinter) PrintStart(hostname string, port uint16) {
	fmt.Printf("TCPinging %s on port %d - saving results to file: %s\n", hostname, port, p.ProbeFile.Name())
}

// PrintProbeSuccess logs a successful probe to the CSV file.
func (p *CSVPrinter) PrintProbeSuccess(startTime time.Time, sourceAddr string, opts types.Options, streak uint, rtt string) {
	if p.cfg.ShowFailuresOnly {
		return
	}

	record := []string{}

	if p.cfg.WithTimestamp {
		record = append(record, startTime.Format(time.DateTime))
	}

	record = append(
		record,
		"Reply",
		opts.Hostname,
		opts.IP.String(),
		fmt.Sprint(opts.Port),
	)

	if p.cfg.WithSourceAddress {
		record = append(record, sourceAddr, fmt.Sprint(streak), rtt)
	}

	record = append(record, fmt.Sprint(streak), rtt)

	if err := p.ProbeWriter.Write(record); err != nil {
		p.PrintError("Failed to write success record: %w", err)
	}

	p.ProbeWriter.Flush()
}

// PrintProbeFailure logs a failed probe attempt to the CSV file.
func (p *CSVPrinter) PrintProbeFailure(startTime time.Time, opts types.Options, streak uint) {
	record := []string{}

	if p.cfg.WithTimestamp {
		record = append(record, startTime.Format(time.DateTime))
	}

	record = append(
		record,
		"No Reply",
		opts.Hostname,
		opts.IP.String(),
		fmt.Sprint(opts.Port),
		fmt.Sprint(streak),
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
func (p *CSVPrinter) PrintRetryingToResolve(hostname string) {
	fmt.Printf("Retrying to resolve %s\n", hostname)
}

// PrintStatistics logs TCPing statistics to a CSV file.
func (p *CSVPrinter) PrintStatistics(t types.Tcping) {
	timestamp := time.Now().Format(time.DateTime)

	statistics := [][]string{
		{"Timestamp", timestamp},
		{"IP Address", t.Options.IP.String()},
	}

	if t.Options.IP.String() != t.Options.Hostname {
		statistics = append(statistics, []string{"Hostname", t.Options.Hostname})
	}

	statistics = append(statistics, []string{"Port", fmt.Sprintf("%d", t.Options.Port)})

	totalDuration := t.TotalDowntime + t.TotalUptime
	statistics = append(statistics, []string{"Total Duration",
		fmt.Sprintf("%.0f", totalDuration.Seconds())},
	)

	statistics = append(statistics, []string{"Total Uptime",
		utils.DurationToString(t.TotalUptime)},
	)
	statistics = append(statistics, []string{"Total Downtime",
		utils.DurationToString(t.TotalDowntime)},
	)

	totalPackets := t.TotalSuccessfulProbes + t.TotalUnsuccessfulProbes
	packetLoss := (float32(t.TotalUnsuccessfulProbes) / float32(totalPackets)) * 100

	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	statistics = append(statistics, []string{"Total Packets", fmt.Sprintf("%d", totalPackets)})
	statistics = append(statistics, []string{"Total Successful Packets", fmt.Sprintf("%d", t.TotalSuccessfulProbes)})
	statistics = append(statistics, []string{"Total Unsuccessful Packets", fmt.Sprintf("%d", t.TotalUnsuccessfulProbes)})
	statistics = append(statistics, []string{"Total Packet Loss Percentage", fmt.Sprintf("%.2f", packetLoss)})

	if t.LongestUptime.Duration != 0 {
		longestUptime := fmt.Sprintf("%.0f", t.LongestUptime.Duration.Seconds())
		longestConsecutiveUptimeStart := t.LongestUptime.Start.Format(time.DateTime)
		longestConsecutiveUptimeEnd := t.LongestUptime.End.Format(time.DateTime)

		statistics = append(statistics, []string{"Longest Uptime", longestUptime})
		statistics = append(statistics, []string{"Longest Consecutive Uptime Start", longestConsecutiveUptimeStart})
		statistics = append(statistics, []string{"Longest Consecutive Uptime End", longestConsecutiveUptimeEnd})
	} else {
		statistics = append(statistics, []string{"Longest Uptime", "Never"})
		statistics = append(statistics, []string{"Longest Consecutive Uptime Start", "Never"})
		statistics = append(statistics, []string{"Longest Consecutive Uptime End", "Never"})
	}

	if t.LongestDowntime.Duration != 0 {
		longestDowntime := fmt.Sprintf("%.0f", t.LongestDowntime.Duration.Seconds())
		longestConsecutiveDowntimeStart := t.LongestDowntime.Start.Format(time.DateTime)
		longestConsecutiveDowntimeEnd := t.LongestDowntime.End.Format(time.DateTime)

		statistics = append(statistics, []string{"Longest Downtime", longestDowntime})
		statistics = append(statistics, []string{"Longest Consecutive Downtime Start", longestConsecutiveDowntimeStart})
		statistics = append(statistics, []string{"Longest Consecutive Downtime End", longestConsecutiveDowntimeEnd})
	} else {
		statistics = append(statistics, []string{"Longest Downtime", "Never"})
		statistics = append(statistics, []string{"Longest Consecutive Downtime Start", "Never"})
		statistics = append(statistics, []string{"Longest Consecutive Downtime End", "Never"})
	}

	if t.RetriedHostnameLookups > 0 {
		statistics = append(statistics, []string{"Hostname Resolve Retries", fmt.Sprintf("%d", t.RetriedHostnameLookups)})
	}

	if len(t.HostnameChanges) > 1 {
		hostnameChanges := ""

		for i := 0; i < len(t.HostnameChanges)-1; i++ {
			if t.HostnameChanges[i].Addr.String() == "" {
				continue
			}

			hostnameChanges += fmt.Sprintf("from %s to %s at %v - ",
				t.HostnameChanges[i].Addr.String(),
				t.HostnameChanges[i+1].Addr.String(),
				t.HostnameChanges[i+1].When.Format(time.DateTime),
			)
		}
	} else {
		statistics = append(statistics, []string{"Hostname Changes", "Never changed"})
	}

	if t.LastSuccessfulProbe.IsZero() {
		statistics = append(statistics, []string{"Last Successful Probe", "Never succeeded"})
	} else {
		statistics = append(statistics, []string{"Last Successful Probe", t.LastSuccessfulProbe.Format(time.DateTime)})
	}

	if t.LastUnsuccessfulProbe.IsZero() {
		statistics = append(statistics, []string{"Last Unsuccessful Probe", "Never failed"})
	} else {
		statistics = append(statistics, []string{"Last Unsuccessful Probe", t.LastUnsuccessfulProbe.Format(time.DateTime)})
	}

	if t.RttResults.HasResults {
		statistics = append(statistics, []string{"Latency Min", fmt.Sprintf("%.3f", t.RttResults.Min)})
		statistics = append(statistics, []string{"Latency Avg", fmt.Sprintf("%.3f", t.RttResults.Average)})
		statistics = append(statistics, []string{"Latency Max", fmt.Sprintf("%.3f", t.RttResults.Max)})
	} else {
		statistics = append(statistics, []string{"Latency Min", "N/A"})
		statistics = append(statistics, []string{"Latency Avg", "N/A"})
		statistics = append(statistics, []string{"Latency Max", "N/A"})
	}

	statistics = append(statistics, []string{"Start Timestamp", t.StartTime.Format(time.DateTime)})

	if !t.EndTime.IsZero() {
		statistics = append(statistics, []string{"End Timestamp", t.EndTime.Format(time.DateTime)})
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
func (p *CSVPrinter) PrintTotalDownTime(_ time.Duration) {}
