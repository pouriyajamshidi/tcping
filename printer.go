package tcping

import (
	"fmt"

	"github.com/pouriyajamshidi/tcping/v3/printers"
	"github.com/pouriyajamshidi/tcping/v3/statistics"
)

var (
	_ Printer = (*printers.ColorPrinter)(nil)
	_ Printer = (*printers.JSONPrinter)(nil)
	_ Printer = (*printers.CSVPrinter)(nil)
	_ Printer = (*printers.DatabasePrinter)(nil)
	_ Printer = (*printers.PlainPrinter)(nil)
)

// Printer defines a set of methods that any printer implementation must provide.
// Printers are responsible for outputting information, but should not modify data or perform calculations.
type Printer interface {
	// PrintStart prints the first message to indicate the target's address and port.
	// This message is printed only once, at the very beginning.
	PrintStart(s *statistics.Statistics)

	// PrintProbeSuccess should print a message after each successful probe.
	// hostname could be empty, meaning it's pinging an address.
	// streak is the number of successful consecutive probes.
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

// NewPrinter creates and returns an appropriate printer based on configuration
func NewPrinter(cfg PrinterConfig) (Printer, error) {
	if cfg.PrettyJSON && !cfg.OutputJSON {
		return nil, fmt.Errorf("--pretty has no effect without the -j flag")
	}

	switch {
	case cfg.OutputJSON:
		return printers.NewJSONPrinter(cfg.PrettyJSON), nil

	case cfg.OutputDBPath != "":
		return printers.NewDatabasePrinter(cfg.Target, cfg.Port, cfg.OutputDBPath)

	case cfg.OutputCSVPath != "":
		return printers.NewCSVPrinter(cfg.OutputCSVPath)

	case cfg.NoColor:
		return printers.NewPlainPrinter(), nil

	default:
		return printers.NewColorPrinter(), nil
	}
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
