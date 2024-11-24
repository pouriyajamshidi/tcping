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

func (cp *csvPrinter) writeHeader() error {
	header := []string{
		"probe status", "hostname", "ip", "port", "TCP_conn", "time",
	}
	return cp.writer.Write(header)
}

func (cp *csvPrinter) writeRecord(record []string) error {
	return cp.writer.Write(record)
}

func (cp *csvPrinter) close() {
	cp.writer.Flush()
	cp.file.Close()
}

// Satisfying the "printer" interface.
func (db *csvPrinter) printProbeSuccess(hostname, ip string, port uint16, streak uint, rtt float32) {}
func (db *csvPrinter) printProbeFail(hostname, ip string, port uint16, streak uint)                 {}
func (db *csvPrinter) printRetryingToResolve(hostname string)                                       {}
func (db *csvPrinter) printTotalDownTime(downtime time.Duration)                                    {}
func (db *csvPrinter) printVersion()                                                                {}
func (db *csvPrinter) printInfo(format string, args ...any)                                         {}
