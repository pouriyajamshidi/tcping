// csv.go outputs data in CSV format
package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strings"
	"time"
)

type csvPrinter struct {
	probeWriter       *csv.Writer
	statsWriter       *csv.Writer
	probeFile         *os.File
	statsFile         *os.File
	statsFilename     string
	probeFilename     string
	headerDone        bool
	statsHeaderDone   bool
	showTimestamp     *bool
	showSourceAddress *bool
	cleanup           func()
}

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

func addCSVExtension(filename string, withStats bool) string {
	if withStats {
		return strings.Split(filename, ".")[0] + "_stats.csv"
	}
	if strings.HasSuffix(filename, ".csv") {
		return filename
	}
	return filename + ".csv"
}

func newCSVPrinter(filename string, showTimestamp *bool, showSourceAddress *bool) (*csvPrinter, error) {
	filename = addCSVExtension(filename, false)

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, filePermission)
	if err != nil {
		return nil, fmt.Errorf("error creating data CSV file: %w", err)
	}

	statsFilename := addCSVExtension(filename, true)
	cp := &csvPrinter{
		probeWriter:       csv.NewWriter(file),
		probeFile:         file,
		probeFilename:     filename,
		statsFilename:     statsFilename,
		showTimestamp:     showTimestamp,
		showSourceAddress: showSourceAddress,
	}

	cp.cleanup = func() {
		if cp.probeWriter != nil {
			cp.probeWriter.Flush()
		}
		if cp.probeFile != nil {
			cp.probeFile.Close()
		}
		if cp.statsWriter != nil {
			cp.statsWriter.Flush()
		}
		if cp.statsFile != nil {
			cp.statsFile.Close()
		}
	}

	return cp, nil
}

func (cp *csvPrinter) writeHeader() error {
	headers := []string{
		colStatus,
		colHostname,
		colIP,
		colPort,
		colTCPConn,
		colLatency,
	}

	if *cp.showSourceAddress {
		headers = append(headers, colSourceAddress)
	}

	if *cp.showTimestamp {
		headers = append(headers, colTimestamp)
	}

	if err := cp.probeWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	cp.probeWriter.Flush()

	return cp.probeWriter.Error()
}

func (cp *csvPrinter) writeRecord(record []string) error {
	if _, err := os.Stat(cp.probeFilename); os.IsNotExist(err) {
		file, err := os.OpenFile(cp.probeFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePermission)
		if err != nil {
			return fmt.Errorf("failed to recreate data CSV file: %w", err)
		}
		cp.probeFile = file
		cp.probeWriter = csv.NewWriter(file)
		cp.headerDone = false
	}

	if !cp.headerDone {
		if err := cp.writeHeader(); err != nil {
			return err
		}
		cp.headerDone = true
	}

	if *cp.showTimestamp {
		record = append(record, time.Now().Format(timeFormat))
	}

	if err := cp.probeWriter.Write(record); err != nil {
		return fmt.Errorf("failed to write record: %w", err)
	}

	cp.probeWriter.Flush()

	return cp.probeWriter.Error()
}

func (cp *csvPrinter) printStart(hostname string, port uint16) {
	fmt.Printf("TCPing results for %s on port %d being written to: %s\n", hostname, port, cp.probeFilename)
}

func (cp *csvPrinter) printProbeSuccess(sourceAddr string, userInput userInput, streak uint, rtt float32) {
	record := []string{
		"Reply",
		userInput.hostname,
		userInput.ip.String(),
		fmt.Sprint(userInput.port),
		fmt.Sprint(streak),
		fmt.Sprintf("%.3f", rtt),
	}

	if *cp.showSourceAddress {
		record = append(record, sourceAddr)
	}

	if err := cp.writeRecord(record); err != nil {
		cp.printError("failed to write success record: %v", err)
	}
}

func (cp *csvPrinter) printProbeFail(userInput userInput, streak uint) {
	record := []string{
		"No reply",
		userInput.hostname,
		userInput.ip.String(),
		fmt.Sprint(userInput.port),
		fmt.Sprint(streak),
		"",
	}

	if *cp.showSourceAddress {
		record = append(record, "")
	}

	if err := cp.writeRecord(record); err != nil {
		cp.printError("failed to write failure record: %v", err)
	}
}

func (cp *csvPrinter) printRetryingToResolve(hostname string) {
	record := []string{
		"Resolving",
		hostname,
		"",
		"",
		"",
		"",
	}

	if err := cp.writeRecord(record); err != nil {
		cp.printError("failed to write resolve record: %v", err)
	}
}

func (cp *csvPrinter) printError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "CSV Error: "+format+"\n", args...)
}

func (cp *csvPrinter) writeStatsHeader() error {
	headers := []string{
		"Metric",
		"Value",
	}

	if err := cp.statsWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write statistics headers: %w", err)
	}

	cp.statsWriter.Flush()

	return cp.statsWriter.Error()
}

func (cp *csvPrinter) writeStatsRecord(record []string) error {
	if _, err := os.Stat(cp.statsFilename); os.IsNotExist(err) {
		statsFile, err := os.OpenFile(cp.statsFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePermission)
		if err != nil {
			return fmt.Errorf("failed to recreate statistics CSV file: %w", err)
		}
		cp.statsFile = statsFile
		cp.statsWriter = csv.NewWriter(statsFile)
		cp.statsHeaderDone = false
	}

	if !cp.statsHeaderDone {
		if err := cp.writeStatsHeader(); err != nil {
			return err
		}
		cp.statsHeaderDone = true
	}

	if err := cp.statsWriter.Write(record); err != nil {
		return fmt.Errorf("failed to write statistics record: %w", err)
	}

	cp.statsWriter.Flush()

	return cp.statsWriter.Error()
}

func (cp *csvPrinter) printStatistics(t tcping) {
	if cp.statsFile == nil {
		statsFile, err := os.OpenFile(cp.statsFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, filePermission)
		if err != nil {
			cp.printError("failed to create statistics CSV file: %v", err)
			return
		}
		cp.statsFile = statsFile
		cp.statsWriter = csv.NewWriter(statsFile)
		cp.statsHeaderDone = false
	}

	totalPackets := t.totalSuccessfulProbes + t.totalUnsuccessfulProbes
	packetLoss := (float32(t.totalUnsuccessfulProbes) / float32(totalPackets)) * 100
	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	// Collect statistics data
	timestamp := time.Now().Format(timeFormat)
	statistics := [][]string{
		{"Timestamp", timestamp},
		{"Total Packets", fmt.Sprint(totalPackets)},
		{"Successful Probes", fmt.Sprint(t.totalSuccessfulProbes)},
		{"Unsuccessful Probes", fmt.Sprint(t.totalUnsuccessfulProbes)},
		{"Packet Loss", fmt.Sprintf("%.2f%%", packetLoss)},
	}

	if t.lastSuccessfulProbe.IsZero() {
		statistics = append(statistics, []string{"Last Successful Probe", "Never succeeded"})
	} else {
		statistics = append(statistics, []string{"Last Successful Probe", t.lastSuccessfulProbe.Format(timeFormat)})
	}

	if t.lastUnsuccessfulProbe.IsZero() {
		statistics = append(statistics, []string{"Last Unsuccessful Probe", "Never failed"})
	} else {
		statistics = append(statistics, []string{"Last Unsuccessful Probe", t.lastUnsuccessfulProbe.Format(timeFormat)})
	}

	statistics = append(statistics, []string{"Total Uptime", durationToString(t.totalUptime)})
	statistics = append(statistics, []string{"Total Downtime", durationToString(t.totalDowntime)})

	if t.longestUptime.duration != 0 {
		statistics = append(statistics,
			[]string{"Longest Uptime Duration", durationToString(t.longestUptime.duration)},
			[]string{"Longest Uptime From", t.longestUptime.start.Format(timeFormat)},
			[]string{"Longest Uptime To", t.longestUptime.end.Format(timeFormat)},
		)
	}

	if t.longestDowntime.duration != 0 {
		statistics = append(statistics,
			[]string{"Longest Downtime Duration", durationToString(t.longestDowntime.duration)},
			[]string{"Longest Downtime From", t.longestDowntime.start.Format(timeFormat)},
			[]string{"Longest Downtime To", t.longestDowntime.end.Format(timeFormat)},
		)
	}

	if !t.destIsIP {
		statistics = append(statistics, []string{"Retried Hostname Lookups", fmt.Sprint(t.retriedHostnameLookups)})

		if len(t.hostnameChanges) >= 2 {
			for i := 0; i < len(t.hostnameChanges)-1; i++ {
				statistics = append(statistics,
					[]string{"IP Change", t.hostnameChanges[i].Addr.String()},
					[]string{"To", t.hostnameChanges[i+1].Addr.String()},
					[]string{"At", t.hostnameChanges[i+1].When.Format(timeFormat)},
				)
			}
		}
	}

	if t.rttResults.hasResults {
		statistics = append(statistics,
			[]string{"RTT Min", fmt.Sprintf("%.3f ms", t.rttResults.min)},
			[]string{"RTT Avg", fmt.Sprintf("%.3f ms", t.rttResults.average)},
			[]string{"RTT Max", fmt.Sprintf("%.3f ms", t.rttResults.max)},
		)
	}

	statistics = append(statistics, []string{"TCPing Started At", t.startTime.Format(timeFormat)})

	if !t.endTime.IsZero() {
		statistics = append(statistics, []string{"TCPing Ended At", t.endTime.Format(timeFormat)})
	}

	durationTime := time.Time{}.Add(t.totalDowntime + t.totalUptime)
	statistics = append(statistics, []string{"Duration (HH:MM:SS)", durationTime.Format(hourFormat)})

	for _, record := range statistics {
		if err := cp.writeStatsRecord(record); err != nil {
			cp.printError("failed to write statistics record: %v", err)
			return
		}
	}

	fmt.Printf("TCPing statistics written to: %s\n", cp.statsFilename)
}

// Satisfying remaining printer interface methods
func (cp *csvPrinter) printTotalDownTime(_ time.Duration) {}
func (cp *csvPrinter) printVersion()                      {}
func (cp *csvPrinter) printInfo(_ string, _ ...any)       {}
