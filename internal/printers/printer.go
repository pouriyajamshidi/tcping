package printers

import (
	"fmt"
	"slices"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
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
	Port              uint16
}

func (p PrinterConfig) GetWithTimestamp() bool {
	return p.WithTimestamp
}

func (p PrinterConfig) GetWithSourceAddress() bool {
	return p.WithSourceAddress
}

// Printer defines a set of methods that any printer implementation must provide.
// Printers are responsible for outputting information, but should not modify data or perform calculations.
type Printer interface {
	// PrintStart prints the first message to indicate the target's address and port.
	// This message is printed only once, at the very beginning.
	PrintStart(s *stats.Statistics)

	// PrintProbeSuccess should print a message after each successful probe.
	PrintProbeSuccess(s *stats.Statistics)

	// PrintProbeFailure should print a message after each failed probe.
	PrintProbeFailure(s *stats.Statistics)

	// PrintStatistics should print all the statistics.
	// This is called on exit and when a user hits the "Enter" key.
	PrintStatistics(s *stats.Statistics)

	// PrintRetryingToResolve should print a message with the hostname
	// it is trying to resolve an IP for.
	// This is only called when the -r flag is provided.
	PrintRetryingToResolve(hostname string)

	// PrintDownTimeDuration should print a downtime duration.
	// This is called when target was unavailable for some time
	// but it has become available now.
	PrintDownTimeDuration(s *stats.Statistics)

	// PrintError prints an error message in red. It takes a print verb and then the arguments.
	PrintError(format string, args ...any)

	// Shutdown sets the end time, prints statistics, and exits the program.
	// This will be removed soon
	Shutdown(s *stats.Statistics)
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
func PrintStats(p Printer, s *stats.Statistics) {
	if s.LastProbeHadFailed {
		stats.SetLongestDuration(s.StartOfDowntime, time.Since(s.StartOfDowntime), &s.LongestDowntime)
	} else {
		stats.SetLongestDuration(s.StartOfUptime, time.Since(s.StartOfUptime), &s.LongestUptime)
	}

	s.RTTResults = calcMinAvgMaxRttTime(s.RTT)

	p.PrintStatistics(s)
}

// calcMinAvgMaxRttTime calculates min, avg and max RTT values
func calcMinAvgMaxRttTime(timeArr []float32) stats.RTTResult {
	var result stats.RTTResult

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
