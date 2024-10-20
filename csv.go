package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

type csvPrinter struct {
	writer     *csv.Writer
	file       *os.File
	filename   string
	headerDone bool
}

const (
	csvTimeFormat = "2006-01-02 15:04:05.000"
)

func newCSVPrinter(filename string, args []string) (*csvPrinter, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("error creating CSV file: %w", err)
	}

	writer := csv.NewWriter(file)
	return &csvPrinter{
		writer:     writer,
		file:       file,
		filename:   filename,
		headerDone: false,
	}, nil
}

func (cp *csvPrinter) writeHeader() error {
	header := []string{
		"Event Type", "Timestamp", "Address", "Hostname", "Port",
		"Hostname Resolve Retries", "Total Successful Probes", "Total Unsuccessful Probes",
		"Never Succeed Probe", "Never Failed Probe", "Last Successful Probe",
		"Last Unsuccessful Probe", "Total Packets", "Total Packet Loss",
		"Total Uptime", "Total Downtime", "Longest Uptime", "Longest Uptime Start",
		"Longest Uptime End", "Longest Downtime", "Longest Downtime Start",
		"Longest Downtime End", "Latency Min", "Latency Avg", "Latency Max",
		"Start Time", "End Time", "Total Duration",
	}
	return cp.writer.Write(header)
}

func (cp *csvPrinter) saveStats(tcping tcping) error {
	if !cp.headerDone {
		if err := cp.writeHeader(); err != nil {
			return fmt.Errorf("error writing CSV header: %w", err)
		}
		cp.headerDone = true
	}

	totalPackets := tcping.totalSuccessfulProbes + tcping.totalUnsuccessfulProbes
	packetLoss := float64(tcping.totalUnsuccessfulProbes) / float64(totalPackets) * 100

	lastSuccessfulProbe := tcping.lastSuccessfulProbe.Format(csvTimeFormat)
	lastUnsuccessfulProbe := tcping.lastUnsuccessfulProbe.Format(csvTimeFormat)
	if tcping.lastSuccessfulProbe.IsZero() {
		lastSuccessfulProbe = ""
	}
	if tcping.lastUnsuccessfulProbe.IsZero() {
		lastUnsuccessfulProbe = ""
	}

	longestUptimeDuration := tcping.longestUptime.duration.String()
	longestUptimeStart := tcping.longestUptime.start.Format(csvTimeFormat)
	longestUptimeEnd := tcping.longestUptime.end.Format(csvTimeFormat)
	longestDowntimeDuration := tcping.longestDowntime.duration.String()
	longestDowntimeStart := tcping.longestDowntime.start.Format(csvTimeFormat)
	longestDowntimeEnd := tcping.longestDowntime.end.Format(csvTimeFormat)

	totalDuration := time.Since(tcping.startTime).String()
	if !tcping.endTime.IsZero() {
		totalDuration = tcping.endTime.Sub(tcping.startTime).String()
	}

	row := []string{
		"statistics",
		time.Now().Format(csvTimeFormat),
		tcping.userInput.ip.String(),
		tcping.userInput.hostname,
		strconv.Itoa(int(tcping.userInput.port)),
		strconv.Itoa(int(tcping.retriedHostnameLookups)),
		strconv.Itoa(int(tcping.totalSuccessfulProbes)),
		strconv.Itoa(int(tcping.totalUnsuccessfulProbes)),
		strconv.FormatBool(tcping.lastSuccessfulProbe.IsZero()),
		strconv.FormatBool(tcping.lastUnsuccessfulProbe.IsZero()),
		lastSuccessfulProbe,
		lastUnsuccessfulProbe,
		strconv.Itoa(int(totalPackets)),
		fmt.Sprintf("%.2f", packetLoss),
		tcping.totalUptime.String(),
		tcping.totalDowntime.String(),
		longestUptimeDuration,
		longestUptimeStart,
		longestUptimeEnd,
		longestDowntimeDuration,
		longestDowntimeStart,
		longestDowntimeEnd,
		fmt.Sprintf("%.3f", tcping.rttResults.min),
		fmt.Sprintf("%.3f", tcping.rttResults.average),
		fmt.Sprintf("%.3f", tcping.rttResults.max),
		tcping.startTime.Format(csvTimeFormat),
		tcping.endTime.Format(csvTimeFormat),
		totalDuration,
	}

	return cp.writer.Write(row)
}

func (cp *csvPrinter) saveHostNameChange(h []hostnameChange) error {
	for _, host := range h {
		if host.Addr.String() == "" {
			continue
		}
		row := []string{
			"hostname change",
			host.When.Format(csvTimeFormat),
			host.Addr.String(),
			"", // Empty fields for consistency with stats rows
			"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
		}
		if err := cp.writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func (cp *csvPrinter) printStart(hostname string, port uint16) {
	fmt.Printf("TCPinging %s on port %d\n", hostname, port)
}

func (cp *csvPrinter) printStatistics(tcping tcping) {
	err := cp.saveStats(tcping)
	if err != nil {
		cp.printError("\nError while writing stats to the CSV file %q\nerr: %s", cp.filename, err)
	}

	if !tcping.endTime.IsZero() {
		err = cp.saveHostNameChange(tcping.hostnameChanges)
		if err != nil {
			cp.printError("\nError while writing hostname changes to the CSV file %q\nerr: %s", cp.filename, err)
		}
	}

	cp.writer.Flush()
	if err := cp.writer.Error(); err != nil {
		cp.printError("Error flushing CSV writer: %v", err)
	}

	fmt.Printf("\nStatistics for %q have been saved to %q\n", tcping.userInput.hostname, cp.filename)
}

func (cp *csvPrinter) printError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func (cp *csvPrinter) Close() error {
	cp.writer.Flush()
	return cp.file.Close()
}

// Satisfying the "printer" interface.
func (db *csvPrinter) printProbeSuccess(hostname, ip string, port uint16, streak uint, rtt float32) {}
func (db *csvPrinter) printProbeFail(hostname, ip string, port uint16, streak uint)                 {}
func (db *csvPrinter) printRetryingToResolve(hostname string)                                       {}
func (db *csvPrinter) printTotalDownTime(downtime time.Duration)                                    {}
func (db *csvPrinter) printVersion()                                                                {}
func (db *csvPrinter) printInfo(format string, args ...any)                                         {}
