// Package printers contains the logic for printing information
package printers

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/types"
)

// PrinterConfig holds all configuration options for Printer creation
type PrinterConfig struct {
	OutputJSON        bool
	PrettyJSON        bool
	NoColor           bool
	WithTimestamp     bool
	WithSourceAddress bool
	ShowFailuresOnly  bool
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

// PrintStats is a helper method for PrintStatistics
// for the current printer.
// This should be used instead, as it makes
// all the necessary calculations beforehand.
func PrintStats(t *types.Tcping) {
	if t.DestWasDown {
		SetLongestDuration(t.StartOfDowntime, time.Since(t.StartOfDowntime), &t.LongestDowntime)
	} else {
		SetLongestDuration(t.StartOfUptime, time.Since(t.StartOfUptime), &t.LongestUptime)
	}

	t.RttResults = calcMinAvgMaxRttTime(t.Rtt)

	t.PrintStatistics(*t)
}

// SetLongestDuration updates the longest uptime or downtime based on the given type.
func SetLongestDuration(start time.Time, duration time.Duration, longest *types.LongestTime) {
	if start.IsZero() || duration == 0 {
		return
	}

	newLongest := types.NewLongestTime(start, duration)

	if longest.End.IsZero() || newLongest.Duration >= longest.Duration {
		*longest = newLongest
	}
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
	var sum float32
	var result types.RttResult

	arrLen := len(timeArr)
	if arrLen > 0 {
		result.Min = timeArr[0]
	}

	for i := 0; i < arrLen; i++ {
		sum += timeArr[i]

		if timeArr[i] > result.Max {
			result.Max = timeArr[i]
		}

		if timeArr[i] < result.Min {
			result.Min = timeArr[i]
		}
	}

	if arrLen > 0 {
		result.HasResults = true
		result.Average = sum / float32(arrLen)
	}

	return result
}
