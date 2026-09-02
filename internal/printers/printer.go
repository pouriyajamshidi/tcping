package printers

import (
	"fmt"

	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
)

// Printer defines a set of methods that any printer implementation must provide.
// Printers are responsible for outputting information, but should not modify data or perform calculations.
type Printer interface {
	// PrintStart prints the first message to indicate the target's address and port.
	// This message is printed only once, at the very beginning.
	PrintStart(s *stats.Statistics)

	// PrintNameResolutionDuration should print how long the initial
	// hostname resolution took. Called once, right after PrintStart -
	// skipped entirely when the target was already a literal IP, since no
	// resolution happened.
	PrintNameResolutionDuration(s *stats.Statistics)

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

	// Shutdown prints statistics and releases any resources the printer
	// holds (e.g. closing files). Statistics must already be finalized
	// (see Prober.finalizeStatistics) by the time this is called. It does
	// not exit the program - callers decide that for themselves.
	Shutdown(s *stats.Statistics)
}

// httpProbeSummary is the short HTTP part of a probe line, e.g. " status=200".
// It is empty for TCP probes and for HTTP probes that never got a response.
func httpProbeSummary(s *stats.Statistics) string {
	if !s.IsHTTP() || !s.HasHTTPResponse() {
		return ""
	}

	return fmt.Sprintf(" status=%s", s.StatusCodeStr())
}

// httpProbeDetails is the indented block printed under a probe line when
// -verbose is on. It is empty unless this is an HTTP probe that got a
// response. The TLS lines are skipped for plain HTTP.
func httpProbeDetails(s *stats.Statistics) string {
	if !s.Verbose || !s.IsHTTP() || !s.HasHTTPResponse() {
		return ""
	}

	msg := fmt.Sprintf("    %s %s\n", s.HTTP.Proto, s.HTTP.Status)

	if s.HTTP.TLSVersion != "" {
		msg += fmt.Sprintf("    %s %s\n", s.HTTP.TLSVersion, s.HTTP.TLSCipherSuite)
		msg += fmt.Sprintf("    certificate expires %s (%d days)\n", s.CertExpiryStr(), s.CertDaysRemaining())
	}

	msg += fmt.Sprintf("    connect=%s ms", s.ConnectDurationStr())
	if s.HTTP.TLSVersion != "" {
		msg += fmt.Sprintf(" tls=%s ms", s.TLSDurationStr())
	}
	msg += fmt.Sprintf(" ttfb=%s ms\n", s.TimeToFirstByteStr())

	return msg
}

// NewPrinter creates and returns an appropriate printer based on configuration
func NewPrinter(cfg config.PrinterConfig) (Printer, error) {
	switch {
	case cfg.OutputJSON:
		return NewJSONPrinter(cfg.PrettyJSON), nil

	case cfg.OutputDBPath != "":
		return NewDatabasePrinter(cfg.Target, cfg.Port, cfg.OutputDBPath)

	case cfg.OutputCSVPath != "":
		return NewCSVPrinter(cfg.OutputCSVPath, cfg.CSVNoTimestamp)

	case cfg.NoColor:
		return NewPlainPrinter(), nil

	default:
		return NewColorPrinter(), nil
	}
}
