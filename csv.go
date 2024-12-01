package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"time"
)

type csvPrinter struct {
	writer          *csv.Writer
	file            *os.File
	dataFilename    string
	headerDone      bool
	showTimestamp   bool
	statsWriter     *csv.Writer
	statsFile       *os.File
	statsFilename   string
	statsHeaderDone bool
	cleanup         func()
}

const (
	csvTimeFormat = "2006-01-02 15:04:05.000"
)

const (
	colStatus    = "Status"
	colTimestamp = "Timestamp"
	colHostname  = "Hostname"
	colIP        = "IP"
	colPort      = "Port"
	colTCPConn   = "TCP_Conn"
	colLatency   = "Latency(ms)"
)

func newCSVPrinter(dataFilename string, showTimestamp bool) (*csvPrinter, error) {
	// Open the data file with the os.O_TRUNC flag to truncate it
	file, err := os.OpenFile(dataFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("error creating data CSV file: %w", err)
	}

	// Append _stats before the .csv extension
	statsFilename := dataFilename
	if len(dataFilename) > 4 && dataFilename[len(dataFilename)-4:] == ".csv" {
		statsFilename = dataFilename[:len(dataFilename)-4] + "_stats.csv"
	} else {
		statsFilename = dataFilename + "_stats"
	}

	cp := &csvPrinter{
		writer:        csv.NewWriter(file),
		file:          file,
		dataFilename:  dataFilename,
		statsFilename: statsFilename,
		showTimestamp: showTimestamp,
	}

	cp.cleanup = func() {
		if cp.writer != nil {
			cp.writer.Flush()
		}
		if cp.file != nil {
			cp.file.Close()
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

	if cp.showTimestamp {
		headers = append(headers, colTimestamp)
	}

	if err := cp.writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	cp.writer.Flush()
	return cp.writer.Error()
}

func (cp *csvPrinter) writeRecord(record []string) error {
	if _, err := os.Stat(cp.dataFilename); os.IsNotExist(err) {
		file, err := os.OpenFile(cp.dataFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to recreate data CSV file: %w", err)
		}
		cp.file = file
		cp.writer = csv.NewWriter(file)
		cp.headerDone = false
	}

	if !cp.headerDone {
		if err := cp.writeHeader(); err != nil {
			return err
		}
		cp.headerDone = true
	}

	if cp.showTimestamp {
		record = append(record, time.Now().Format(csvTimeFormat))
	}

	if err := cp.writer.Write(record); err != nil {
		return fmt.Errorf("failed to write record: %w", err)
	}

	cp.writer.Flush()
	return cp.writer.Error()
}

func (cp *csvPrinter) printStart(hostname string, port uint16) {
	fmt.Printf("TCPing results being written to: %s\n", cp.dataFilename)
}

func (cp *csvPrinter) printProbeSuccess(hostname, ip string, port uint16, streak uint, rtt float32) {
	record := []string{
		"Success",
		hostname,
		ip,
		fmt.Sprint(port),
		fmt.Sprint(streak),
		fmt.Sprintf("%.3f", rtt),
	}

	if err := cp.writeRecord(record); err != nil {
		cp.printError("Failed to write success record: %v", err)
	}
}

func (cp *csvPrinter) printProbeFail(hostname, ip string, port uint16, streak uint) {
	record := []string{
		"Failed",
		hostname,
		ip,
		fmt.Sprint(port),
		fmt.Sprint(streak),
		"-",
	}

	if err := cp.writeRecord(record); err != nil {
		cp.printError("Failed to write failure record: %v", err)
	}
}

func (cp *csvPrinter) printRetryingToResolve(hostname string) {
	record := []string{
		"Resolving",
		hostname,
		"-",
		"-",
		"-",
		"-",
	}

	if err := cp.writeRecord(record); err != nil {
		cp.printError("Failed to write resolve record: %v", err)
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
		statsFile, err := os.OpenFile(cp.statsFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to recreate statistics CSV file: %w", err)
		}
		cp.statsFile = statsFile
		cp.statsWriter = csv.NewWriter(statsFile)
		cp.statsHeaderDone = false
	}

	// Write header if not done
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
	// Initialize stats file if not already done
	if cp.statsFile == nil {
		statsFile, err := os.OpenFile(cp.statsFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0644)
		if err != nil {
			cp.printError("Failed to create statistics CSV file: %v", err)
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
	timestamp := time.Now().Format(time.RFC3339)
	statistics := [][]string{
		{"Timestamp", timestamp},
		{"Total Packets", fmt.Sprint(totalPackets)},
		{"Successful Probes", fmt.Sprint(t.totalSuccessfulProbes)},
		{"Unsuccessful Probes", fmt.Sprint(t.totalUnsuccessfulProbes)},
		{"Packet Loss", fmt.Sprintf("%.2f%%", packetLoss)},
		{"Last Successful Probe", t.lastSuccessfulProbe.Format(timeFormat)},
		{"Last Unsuccessful Probe", t.lastUnsuccessfulProbe.Format(timeFormat)},
		{"Total Uptime", durationToString(t.totalUptime)},
		{"Total Downtime", durationToString(t.totalDowntime)},
	}

	if t.longestUptime.duration != 0 {
		statistics = append(statistics, []string{
			"Longest Uptime", durationToString(t.longestUptime.duration),
			"From", t.longestUptime.start.Format(timeFormat),
			"To", t.longestUptime.end.Format(timeFormat),
		})
	}

	if t.longestDowntime.duration != 0 {
		statistics = append(statistics, []string{
			"Longest Downtime", durationToString(t.longestDowntime.duration),
			"From", t.longestDowntime.start.Format(timeFormat),
			"To", t.longestDowntime.end.Format(timeFormat),
		})
	}

	if !t.destIsIP {
		statistics = append(statistics, []string{
			"Retried Hostname Lookups", fmt.Sprint(t.retriedHostnameLookups),
		})

		if len(t.hostnameChanges) >= 2 {
			for i := 0; i < len(t.hostnameChanges)-1; i++ {
				statistics = append(statistics, []string{
					"IP Change", t.hostnameChanges[i].Addr.String(),
					"To", t.hostnameChanges[i+1].Addr.String(),
					"At", t.hostnameChanges[i+1].When.Format(timeFormat),
				})
			}
		}
	}

	if t.rttResults.hasResults {
		statistics = append(statistics, []string{
			"RTT Min", fmt.Sprintf("%.3f ms", t.rttResults.min),
			"RTT Avg", fmt.Sprintf("%.3f ms", t.rttResults.average),
			"RTT Max", fmt.Sprintf("%.3f ms", t.rttResults.max),
		})
	}

	statistics = append(statistics, []string{
		"TCPing Started At", t.startTime.Format(timeFormat),
	})

	if !t.endTime.IsZero() {
		statistics = append(statistics, []string{
			"TCPing Ended At", t.endTime.Format(timeFormat),
		})
	}

	durationTime := time.Time{}.Add(t.totalDowntime + t.totalUptime)
	statistics = append(statistics, []string{
		"Duration (HH:MM:SS)", durationTime.Format(hourFormat),
	})

	// Write statistics to CSV
	for _, record := range statistics {
		if err := cp.writeStatsRecord(record); err != nil {
			cp.printError("Failed to write statistics record: %v", err)
			return
		}
	}

	// Print the message only if statistics are written
	fmt.Printf("TCPing statistics written to: %s\n", cp.statsFilename)
}

// Satisfying remaining printer interface methods
func (cp *csvPrinter) printTotalDownTime(downtime time.Duration) {}
func (cp *csvPrinter) printVersion()                             {}
func (cp *csvPrinter) printInfo(format string, args ...any)      {}
