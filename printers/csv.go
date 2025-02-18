// Package printers contains the logic for printing information
package printers

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/consts"
	"github.com/pouriyajamshidi/tcping/v2/internal/utils"
	"github.com/pouriyajamshidi/tcping/v2/types"
)

const (
	colStatus        = "Status"
	colTimestamp     = "Timestamp"
	colHostname      = "Hostname"
	colIP            = "IP"
	colPort          = "Port"
	colTCPConn       = "TCP_Conn"
	colLatency       = "Latency(ms)"
	colSourceAddress = "Source Address"
)

const (
	filePermission os.FileMode = 0644
)

// CSVPrinter is responsible for writing probe results and statistics to CSV files.
type CSVPrinter struct {
	ProbeWriter       *csv.Writer
	StatsWriter       *csv.Writer
	ProbeFile         *os.File
	StatsFile         *os.File
	StatsFilename     string
	ProbeFilename     string
	HeaderDone        bool
	StatsHeaderDone   bool
	ShowTimestamp     bool
	ShowSourceAddress bool
	Cleanup           func()
}

// NewCSVPrinter initializes a CSVPrinter instance with the given filename and settings.
func NewCSVPrinter(cfg PrinterConfig) (*CSVPrinter, error) {
	filename := addCSVExtension(cfg.OutputCSVPath, false)

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, filePermission)
	if err != nil {
		return nil, fmt.Errorf("Failed creating %s data CSV file: %w", cfg.OutputCSVPath, err)
	}

	statsFilename := addCSVExtension(filename, true)

	cp := &CSVPrinter{
		ProbeWriter:       csv.NewWriter(file),
		ProbeFile:         file,
		ProbeFilename:     filename,
		StatsFilename:     statsFilename,
		ShowTimestamp:     cfg.WithTimestamp,
		ShowSourceAddress: cfg.WithSourceAddress,
	}

	cp.Cleanup = func() {
		if cp.ProbeWriter != nil {
			cp.ProbeWriter.Flush()
		}

		if cp.ProbeFile != nil {
			cp.ProbeFile.Close()
		}

		if cp.StatsWriter != nil {
			cp.StatsWriter.Flush()
		}

		if cp.StatsFile != nil {
			cp.StatsFile.Close()
		}
	}

	return cp, nil
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

func (p *CSVPrinter) writeHeader() error {
	headers := []string{
		colStatus,
		colHostname,
		colIP,
		colPort,
		colTCPConn,
		colLatency,
	}

	if p.ShowSourceAddress {
		headers = append(headers, colSourceAddress)
	}

	if p.ShowTimestamp {
		headers = append(headers, colTimestamp)
	}

	if err := p.ProbeWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	p.ProbeWriter.Flush()

	return p.ProbeWriter.Error()
}

func (p *CSVPrinter) writeRecord(record []string) error {
	if _, err := os.Stat(p.ProbeFilename); os.IsNotExist(err) {
		file, err := os.OpenFile(p.ProbeFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePermission)
		if err != nil {
			return fmt.Errorf("failed to recreate data CSV file: %w", err)
		}

		p.ProbeFile = file
		p.ProbeWriter = csv.NewWriter(file)
		p.HeaderDone = false
	}

	if !p.HeaderDone {
		if err := p.writeHeader(); err != nil {
			return err
		}

		p.HeaderDone = true
	}

	if p.ShowTimestamp {
		record = append(record, time.Now().Format(consts.TimeFormat))
	}

	if err := p.ProbeWriter.Write(record); err != nil {
		return fmt.Errorf("failed to write record: %w", err)
	}

	p.ProbeWriter.Flush()

	return p.ProbeWriter.Error()
}

// PrintStart logs the beginning of a TCPing session.
func (p *CSVPrinter) PrintStart(hostname string, port uint16) {
	fmt.Printf("TCPing results for %s on port %d being written to: %s\n", hostname, port, p.ProbeFilename)
}

// PrintProbeSuccess logs a successful probe attempt to the CSV file.
func (p *CSVPrinter) PrintProbeSuccess(startTime time.Time, sourceAddr string, opts types.Options, streak uint, rtt string) {
	record := []string{
		"Reply",
		opts.Hostname,
		opts.IP.String(),
		fmt.Sprint(opts.Port),
		fmt.Sprint(streak),
		rtt,
	}

	if p.ShowSourceAddress {
		record = append(record, sourceAddr)
	}

	if err := p.writeRecord(record); err != nil {
		p.PrintError("Failed to write success record: %v", err)
	}
}

// PrintProbeFail logs a failed probe attempt to the CSV file.
func (p *CSVPrinter) PrintProbeFail(startTime time.Time, opts types.Options, streak uint) {
	record := []string{
		"No reply",
		opts.Hostname,
		opts.IP.String(),
		fmt.Sprint(opts.Port),
		fmt.Sprint(streak),
		"",
	}

	if p.ShowSourceAddress {
		record = append(record, "")
	}

	if err := p.writeRecord(record); err != nil {
		p.PrintError("Failed to write failure record: %v", err)
	}
}

// PrintRetryingToResolve logs an attempt to resolve a hostname.
func (p *CSVPrinter) PrintRetryingToResolve(hostname string) {
	record := []string{
		"Resolving",
		hostname,
		"",
		"",
		"",
		"",
	}

	if err := p.writeRecord(record); err != nil {
		p.PrintError("Failed to write resolve record: %v", err)
	}
}

// PrintError logs an error message to stderr.
func (p *CSVPrinter) PrintError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "CSV Error: "+format+"\n", args...)
}

func (p *CSVPrinter) writeStatsHeader() error {
	headers := []string{
		"Metric",
		"Value",
	}

	if err := p.StatsWriter.Write(headers); err != nil {
		return fmt.Errorf("Failed to write statistics headers: %w", err)
	}

	p.StatsWriter.Flush()

	return p.StatsWriter.Error()
}

func (p *CSVPrinter) writeStatsRecord(record []string) error {
	if _, err := os.Stat(p.StatsFilename); os.IsNotExist(err) {
		statsFile, err := os.OpenFile(p.StatsFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePermission)
		if err != nil {
			return fmt.Errorf("Failed to recreate statistics CSV file: %w", err)
		}
		p.StatsFile = statsFile
		p.StatsWriter = csv.NewWriter(statsFile)
		p.StatsHeaderDone = false
	}

	if !p.StatsHeaderDone {
		if err := p.writeStatsHeader(); err != nil {
			return err
		}
		p.StatsHeaderDone = true
	}

	if err := p.StatsWriter.Write(record); err != nil {
		return fmt.Errorf("Failed to write statistics record: %w", err)
	}

	p.StatsWriter.Flush()

	return p.StatsWriter.Error()
}

// PrintStatistics logs TCPing statistics to a CSV file.
func (p *CSVPrinter) PrintStatistics(t types.Tcping) {
	if p.StatsFile == nil {
		statsFile, err := os.OpenFile(p.StatsFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, filePermission)
		if err != nil {
			p.PrintError("Failed to create statistics CSV file: %v", err)
			return
		}
		p.StatsFile = statsFile
		p.StatsWriter = csv.NewWriter(statsFile)
		p.StatsHeaderDone = false
	}

	var packetLoss float32
	totalPackets := t.TotalSuccessfulProbes + t.TotalUnsuccessfulProbes
	if totalPackets == 0 {
		packetLoss = 0
	} else {
		packetLoss = (float32(t.TotalUnsuccessfulProbes) / float32(totalPackets)) * 100
	}

	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	// Collect statistics data
	timestamp := time.Now().Format(consts.TimeFormat)
	statistics := [][]string{
		{"Timestamp", timestamp},
		{"Total Packets", fmt.Sprint(totalPackets)},
		{"Successful Probes", fmt.Sprint(t.TotalSuccessfulProbes)},
		{"Unsuccessful Probes", fmt.Sprint(t.TotalUnsuccessfulProbes)},
		{"Packet Loss", fmt.Sprintf("%.2f%%", packetLoss)},
	}

	if t.LastSuccessfulProbe.IsZero() {
		statistics = append(statistics, []string{"Last Successful Probe", "Never succeeded"})
	} else {
		statistics = append(statistics, []string{"Last Successful Probe", t.LastSuccessfulProbe.Format(consts.TimeFormat)})
	}

	if t.LastUnsuccessfulProbe.IsZero() {
		statistics = append(statistics, []string{"Last Unsuccessful Probe", "Never failed"})
	} else {
		statistics = append(statistics, []string{"Last Unsuccessful Probe", t.LastUnsuccessfulProbe.Format(consts.TimeFormat)})
	}

	statistics = append(statistics, []string{"Total Uptime", utils.DurationToString(t.TotalUptime)})
	statistics = append(statistics, []string{"Total Downtime", utils.DurationToString(t.TotalDowntime)})

	if t.LongestUptime.Duration != 0 {
		statistics = append(statistics,
			[]string{"Longest Uptime Duration", utils.DurationToString(t.LongestUptime.Duration)},
			[]string{"Longest Uptime From", t.LongestUptime.Start.Format(consts.TimeFormat)},
			[]string{"Longest Uptime To", t.LongestUptime.End.Format(consts.TimeFormat)},
		)
	}

	if t.LongestDowntime.Duration != 0 {
		statistics = append(statistics,
			[]string{"Longest Downtime Duration", utils.DurationToString(t.LongestDowntime.Duration)},
			[]string{"Longest Downtime From", t.LongestDowntime.Start.Format(consts.TimeFormat)},
			[]string{"Longest Downtime To", t.LongestDowntime.End.Format(consts.TimeFormat)},
		)
	}

	if !t.DestIsIP {
		statistics = append(statistics, []string{"Retried Hostname Lookups", fmt.Sprint(t.RetriedHostnameLookups)})

		if len(t.HostnameChanges) >= 2 {
			for i := 0; i < len(t.HostnameChanges)-1; i++ {
				statistics = append(statistics,
					[]string{"IP Change", t.HostnameChanges[i].Addr.String()},
					[]string{"To", t.HostnameChanges[i+1].Addr.String()},
					[]string{"At", t.HostnameChanges[i+1].When.Format(consts.TimeFormat)},
				)
			}
		}
	}

	if t.RttResults.HasResults {
		statistics = append(statistics,
			[]string{"RTT Min", fmt.Sprintf("%.3f ms", t.RttResults.Min)},
			[]string{"RTT Avg", fmt.Sprintf("%.3f ms", t.RttResults.Average)},
			[]string{"RTT Max", fmt.Sprintf("%.3f ms", t.RttResults.Max)},
		)
	}

	statistics = append(statistics, []string{"TCPing Started At", t.StartTime.Format(consts.TimeFormat)})

	if !t.EndTime.IsZero() {
		statistics = append(statistics, []string{"TCPing Ended At", t.EndTime.Format(consts.TimeFormat)})
	}

	durationTime := time.Time{}.Add(t.TotalDowntime + t.TotalUptime)
	statistics = append(statistics, []string{"Duration (HH:MM:SS)", durationTime.Format(consts.HourFormat)})

	for _, record := range statistics {
		if err := p.writeStatsRecord(record); err != nil {
			p.PrintError("Failed to write statistics record: %v", err)
			return
		}
	}

	fmt.Printf("TCPing statistics written to: %s\n", p.StatsFilename)
}

// PrintTotalDownTime is a no-op implementation to satisfy the Printer interface.
func (p *CSVPrinter) PrintTotalDownTime(_ time.Duration) {}
