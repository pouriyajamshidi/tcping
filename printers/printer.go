// Package printers contains the logic for printing information
package printers

import (
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/utils"
	"github.com/pouriyajamshidi/tcping/v3/types"
)

// PrinterConfig holds all configuration options for Printer creation
type PrinterConfig struct {
	OutputJSON        bool
	PrettyJSON        bool
	NoColor           bool
	WithTimestamp     bool
	WithSourceAddress bool
	OutputDBPath      string
	OutputCSVPath     string
	Target            string
	Port              string
}

// NewPrinter creates and returns an appropriate printer based on configuration
func NewPrinter(cfg PrinterConfig) (types.Printer, error) {
	if cfg.PrettyJSON && !cfg.OutputJSON {
		return nil, fmt.Errorf("--pretty has no effect without the -j flag")
	}

	switch {
	case cfg.OutputJSON:
		return NewJSONPrinter(cfg), nil

	case cfg.OutputDBPath != "":
		return NewDatabasePrinter(cfg)

	case cfg.OutputCSVPath != "":
		return NewCSVPrinter(cfg)

	case cfg.NoColor:
		return NewPlainPrinter(cfg), nil

	default:
		return NewColorPrinter(cfg), nil
	}
}

// PrintStats is a helper method for PrintStatistics of the current printer.
// This should be used instead, as it makes all the necessary calculations beforehand.
func PrintStats(t *types.Tcping) {
	if t.DestWasDown {
		utils.SetLongestDuration(t.StartOfDowntime, time.Since(t.StartOfDowntime), &t.LongestDowntime)
	} else {
		utils.SetLongestDuration(t.StartOfUptime, time.Since(t.StartOfUptime), &t.LongestUptime)
	}

	t.RttResults = calcMinAvgMaxRttTime(t.Rtt)

	t.PrintStatistics(*t)
}

// Shutdown calculates endTime, prints statistics and calls os.Exit(0).
// This should be used as the main exit-point.
func Shutdown(p *types.Tcping) {
	p.EndTime = time.Now()
	PrintStats(p)

	// if the printer type is `database`, close it before exiting
	if db, ok := p.Printer.(*DatabasePrinter); ok {
		db.Done()
	}

	// if the printer type is `csvPrinter`, call the cleanup function before exiting
	if cp, ok := p.Printer.(*CSVPrinter); ok {
		cp.Done()
	}

	os.Exit(0)
}

// SignalHandler catches SIGINT and SIGTERM then prints tcping stats
func SignalHandler(p *types.Tcping) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		Shutdown(p)
	}()
}

// calcMinAvgMaxRttTime calculates min, avg and max RTT values
func calcMinAvgMaxRttTime(timeArr []float32) types.RttResult {
	var result types.RttResult

	arrLen := len(timeArr)
	if arrLen == 0 {
		return result
	}

	var sum float32

	for _, t := range timeArr {
		sum += t
	}

	result.Min = slices.Min(timeArr)
	result.Max = slices.Max(timeArr)
	result.Average = sum / float32(arrLen)
	result.HasResults = true

	return result
}
