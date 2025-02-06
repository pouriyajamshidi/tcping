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

// CsvPrinter is responsible for writing probe results and statistics to CSV files.
type CsvPrinter struct {
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

// NewCSVPrinter initializes a CsvPrinter instance with the given filename and settings.
func NewCSVPrinter(filename string, showTimestamp bool, showSourceAddress bool) (*CsvPrinter, error) {
	filename = addCSVExtension(filename, false)

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, filePermission)
	if err != nil {
		return nil, fmt.Errorf("Failed creating data CSV file: %w", err)
	}

	statsFilename := addCSVExtension(filename, true)

	cp := &CsvPrinter{
		ProbeWriter:       csv.NewWriter(file),
		ProbeFile:         file,
		ProbeFilename:     filename,
		StatsFilename:     statsFilename,
		ShowTimestamp:     showTimestamp,
		ShowSourceAddress: showSourceAddress,
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

func addCSVExtension(filename string, withStats bool) string {
	if withStats {
		return strings.Split(filename, ".")[0] + "_stats.csv"
	}

	if strings.HasSuffix(filename, ".csv") {
		return filename
	}

	return filename + ".csv"
}

func (cp *CsvPrinter) writeHeader() error {
	headers := []string{
		colStatus,
		colHostname,
		colIP,
		colPort,
		colTCPConn,
		colLatency,
	}

	if cp.ShowSourceAddress {
		headers = append(headers, colSourceAddress)
	}

	if cp.ShowTimestamp {
		headers = append(headers, colTimestamp)
	}

	if err := cp.ProbeWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	cp.ProbeWriter.Flush()

	return cp.ProbeWriter.Error()
}

func (cp *CsvPrinter) writeRecord(record []string) error {
	if _, err := os.Stat(cp.ProbeFilename); os.IsNotExist(err) {
		file, err := os.OpenFile(cp.ProbeFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePermission)
		if err != nil {
			return fmt.Errorf("failed to recreate data CSV file: %w", err)
		}

		cp.ProbeFile = file
		cp.ProbeWriter = csv.NewWriter(file)
		cp.HeaderDone = false
	}

	if !cp.HeaderDone {
		if err := cp.writeHeader(); err != nil {
			return err
		}

		cp.HeaderDone = true
	}

	if cp.ShowTimestamp {
		record = append(record, time.Now().Format(consts.TimeFormat))
	}

	if err := cp.ProbeWriter.Write(record); err != nil {
		return fmt.Errorf("failed to write record: %w", err)
	}

	cp.ProbeWriter.Flush()

	return cp.ProbeWriter.Error()
}

// PrintStart logs the beginning of a TCPing session.
func (cp *CsvPrinter) PrintStart(hostname string, port uint16) {
	fmt.Printf("TCPing results for %s on port %d being written to: %s\n", hostname, port, cp.ProbeFilename)
}

// PrintProbeSuccess logs a successful probe attempt to the CSV file.
func (cp *CsvPrinter) PrintProbeSuccess(startTime time.Time, sourceAddr string, opts types.Options, streak uint, rtt float32) {
	record := []string{
		"Reply",
		opts.Hostname,
		opts.IP.String(),
		fmt.Sprint(opts.Port),
		fmt.Sprint(streak),
		fmt.Sprintf("%.3f", rtt),
	}

	if cp.ShowSourceAddress {
		record = append(record, sourceAddr)
	}

	if err := cp.writeRecord(record); err != nil {
		cp.PrintError("failed to write success record: %v", err)
	}
}

// PrintProbeFail logs a failed probe attempt to the CSV file.
func (cp *CsvPrinter) PrintProbeFail(startTime time.Time, opts types.Options, streak uint) {
	record := []string{
		"No reply",
		opts.Hostname,
		opts.IP.String(),
		fmt.Sprint(opts.Port),
		fmt.Sprint(streak),
		"",
	}

	if cp.ShowSourceAddress {
		record = append(record, "")
	}

	if err := cp.writeRecord(record); err != nil {
		cp.PrintError("failed to write failure record: %v", err)
	}
}

// PrintRetryingToResolve logs an attempt to resolve a hostname.
func (cp *CsvPrinter) PrintRetryingToResolve(hostname string) {
	record := []string{
		"Resolving",
		hostname,
		"",
		"",
		"",
		"",
	}

	if err := cp.writeRecord(record); err != nil {
		cp.PrintError("failed to write resolve record: %v", err)
	}
}

// PrintError logs an error message to stderr.
func (cp *CsvPrinter) PrintError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "CSV Error: "+format+"\n", args...)
}

func (cp *CsvPrinter) writeStatsHeader() error {
	headers := []string{
		"Metric",
		"Value",
	}

	if err := cp.StatsWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write statistics headers: %w", err)
	}

	cp.StatsWriter.Flush()

	return cp.StatsWriter.Error()
}

func (cp *CsvPrinter) writeStatsRecord(record []string) error {
	if _, err := os.Stat(cp.StatsFilename); os.IsNotExist(err) {
		statsFile, err := os.OpenFile(cp.StatsFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePermission)
		if err != nil {
			return fmt.Errorf("failed to recreate statistics CSV file: %w", err)
		}
		cp.StatsFile = statsFile
		cp.StatsWriter = csv.NewWriter(statsFile)
		cp.StatsHeaderDone = false
	}

	if !cp.StatsHeaderDone {
		if err := cp.writeStatsHeader(); err != nil {
			return err
		}
		cp.StatsHeaderDone = true
	}

	if err := cp.StatsWriter.Write(record); err != nil {
		return fmt.Errorf("failed to write statistics record: %w", err)
	}

	cp.StatsWriter.Flush()

	return cp.StatsWriter.Error()
}

// PrintStatistics logs TCPing statistics to a CSV file.
func (cp *CsvPrinter) PrintStatistics(t types.Tcping) {
	if cp.StatsFile == nil {
		statsFile, err := os.OpenFile(cp.StatsFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, filePermission)
		if err != nil {
			cp.PrintError("failed to create statistics CSV file: %v", err)
			return
		}
		cp.StatsFile = statsFile
		cp.StatsWriter = csv.NewWriter(statsFile)
		cp.StatsHeaderDone = false
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
		if err := cp.writeStatsRecord(record); err != nil {
			cp.PrintError("failed to write statistics record: %v", err)
			return
		}
	}

	fmt.Printf("TCPing statistics written to: %s\n", cp.StatsFilename)
}

// PrintTotalDownTime is a no-op implementation to satisfy the Printer interface.
func (cp *CsvPrinter) PrintTotalDownTime(_ time.Duration) {}

// PrintInfo is a no-op implementation to satisfy the Printer interface.
func (cp *CsvPrinter) PrintInfo(_ string, _ ...any) {}
