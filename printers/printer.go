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
	"github.com/pouriyajamshidi/tcping/v3/probes/statistics"
	"github.com/pouriyajamshidi/tcping/v3/types"
)

// Printer defines a set of methods that any printer implementation must provide.
// Printers are responsible for outputting information, but should not modify data or perform calculations.
type Printer interface {
	// PrintStart prints the first message to indicate the target's address and port.
	// This message is printed only once, at the very beginning.
	PrintStart(s *statistics.Statistics)

	// PrintProbeSuccess should print a message after each successful probe.
	// hostname could be empty, meaning it's pinging an address.
	// streak is the number of successful consecuti`ve probes.
	PrintProbeSuccess(s *statistics.Statistics)

	// PrintProbeFailure should print a message after each failed probe.
	// hostname could be empty, meaning it's pinging an address.
	// streak is the number of successful consecutive probes.
	PrintProbeFailure(s *statistics.Statistics)

	// PrintRetryingToResolve should print a message with the hostname
	// it is trying to resolve an IP for.
	//
	// This is only being printed when the -r flag is applied.
	PrintRetryingToResolve(s *statistics.Statistics)

	// PrintTotalDownTime should print a downtime duration.
	//
	// This is being called when host was unavailable for some time
	// but the latest probe was successful (became available).
	PrintTotalDownTime(s *statistics.Statistics)

	// PrintStatistics should print a message with
	// helpful statistics information.
	//
	// This is being called on exit and when user hits "Enter".
	PrintStatistics(s *statistics.Statistics)

	// PrintError should print an error message.
	// Printer should also apply \n to the given string, if needed.
	PrintError(format string, args ...any)

	// Shutdown sets the EndTime, calls PrintStatistics() and Done() then exits the program.
	Shutdown(s *statistics.Statistics)
}

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
func NewPrinter(cfg PrinterConfig) (Printer, error) {
	if cfg.PrettyJSON && !cfg.OutputJSON {
		return nil, fmt.Errorf("--pretty has no effect without the -j flag")
	}

	switch {
	case cfg.OutputJSON:
		return NewJSONPrinter(cfg.PrettyJSON), nil

	case cfg.OutputDBPath != "":
		return NewDatabasePrinter(cfg.Target, cfg.Port, cfg.OutputDBPath)

	case cfg.OutputCSVPath != "":
		return NewCSVPrinter(cfg.OutputCSVPath)

	case cfg.NoColor:
		return NewPlainPrinter(), nil

	default:
		return NewColorPrinter(), nil
	}
}

// PrintStats is a helper method for PrintStatistics of the current printer.
// This should be used instead of directly calling the PrintStatistics
// as it makes the common calculations beforehand.
func PrintStats(p Printer, s *statistics.Statistics) {
	if s.DestWasDown {
		utils.SetLongestDuration(s.StartOfDowntime, time.Since(s.StartOfDowntime), &s.LongestDowntime)
	} else {
		utils.SetLongestDuration(s.StartOfUptime, time.Since(s.StartOfUptime), &s.LongestUptime)
	}

	s.RTTResults = calcMinAvgMaxRttTime(s.RTT)

	p.PrintStatistics(s)
}

// SignalHandler catches SIGINT and SIGTERM then prints tcping stats
func SignalHandler(p Printer, s *statistics.Statistics) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		p.Shutdown(s)
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
