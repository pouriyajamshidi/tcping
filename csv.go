package main

import (
	"encoding/csv"
	"fmt"
	"os"
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

func (cp *csvPrinter) writeRecord(record []string) error {
	return cp.writer.Write(record)
}

func (cp *csvPrinter) close() {
	cp.writer.Flush()
	cp.file.Close()
}

func (cp *csvPrinter) printError(format string, args ...any) {
	colorRed("Error: "+format, args...)
}

func (cp *csvPrinter) printStart(hostname string, port uint16) {
	header := []string{
		"status", "hostname", "ip", "port", "TCP_conn", "time",
	}
	cp.writer.Write(header)
}

func (cp *csvPrinter) printProbeSuccess(hostname, ip string, port uint16, streak uint, rtt float32) {
	cp.writeRecord([]string{
		"reply", hostname, ip, fmt.Sprint(port), fmt.Sprint(streak), fmt.Sprintf("%.3f", rtt),
	})
}

func (cp *csvPrinter) printProbeFail(hostname, ip string, port uint16, streak uint) {
	cp.writeRecord([]string{
		"no reply", hostname, ip, fmt.Sprint(port), fmt.Sprint(streak), "",
	})
}

// Satisfying the "printer" interface.
func (cp *csvPrinter) printStatistics(s tcping)                  {}
func (cp *csvPrinter) printRetryingToResolve(hostname string)    {}
func (cp *csvPrinter) printTotalDownTime(downtime time.Duration) {}
func (cp *csvPrinter) printVersion()                             {}
func (cp *csvPrinter) printInfo(format string, args ...any)      {}
