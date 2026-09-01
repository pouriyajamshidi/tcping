package printers

import (
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

	// PrintUpTimeDuration should print an uptime duration.
	// This is called when target was available for some time
	// but it has just become unavailable.
	PrintUpTimeDuration(s *stats.Statistics)

	// PrintError prints an error message in red. It takes a print verb and then the arguments.
	PrintError(format string, args ...any)

	// Shutdown prints statistics and exits the program. Statistics must
	// already be finalized (see Prober.finalizeStatistics) by the time
	// this is called.
	Shutdown(s *stats.Statistics)
}

// NewPrinter creates and returns an appropriate printer based on configuration
func NewPrinter(cfg PrinterConfig) (Printer, error) {
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
