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
	// PrintError should NOT exit the program - exit decisions belong to the application layer.
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
		opts := []printers.JSONPrinterOption{}
		if cfg.PrettyJSON {
			opts = append(opts, printers.WithPrettyJSON())
		}
		if cfg.WithTimestamp {
			opts = append(opts, printers.WithTimestamp[*printers.JSONPrinter]())
		}
		if cfg.WithSourceAddress {
			opts = append(opts, printers.WithSourceAddress[*printers.JSONPrinter]())
		}
		return printers.NewJSONPrinter(opts...), nil

	case cfg.OutputDBPath != "":
		opts := []printers.DatabasePrinterOption{}
		if cfg.WithTimestamp {
			opts = append(opts, printers.WithTimestamp[*printers.DatabasePrinter]())
		}
		if cfg.WithSourceAddress {
			opts = append(opts, printers.WithSourceAddress[*printers.DatabasePrinter]())
		}
		return printers.NewDatabasePrinter(cfg.Target, cfg.Port, cfg.OutputDBPath, opts...)

	case cfg.OutputCSVPath != "":
		opts := []printers.CSVPrinterOption{}
		if cfg.WithTimestamp {
			opts = append(opts, printers.WithTimestamp[*printers.CSVPrinter]())
		}
		if cfg.WithSourceAddress {
			opts = append(opts, printers.WithSourceAddress[*printers.CSVPrinter]())
		}
		return printers.NewCSVPrinter(cfg.OutputCSVPath, opts...)

	case cfg.NoColor:
		opts := []printers.PlainPrinterOption{}
		if cfg.WithTimestamp {
			opts = append(opts, printers.WithTimestamp[*printers.PlainPrinter]())
		}
		if cfg.WithSourceAddress {
			opts = append(opts, printers.WithSourceAddress[*printers.PlainPrinter]())
		}
		return printers.NewPlainPrinter(opts...), nil

	default:
		opts := []printers.ColorPrinterOption{}
		if cfg.WithTimestamp {
			opts = append(opts, printers.WithTimestamp[*printers.ColorPrinter]())
		}
		if cfg.WithSourceAddress {
			opts = append(opts, printers.WithSourceAddress[*printers.ColorPrinter]())
		}
		return printers.NewColorPrinter(opts...), nil
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
