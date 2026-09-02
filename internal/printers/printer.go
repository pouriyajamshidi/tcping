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
// verbose is on. It is empty unless this is an HTTP probe that got a
// response. The TLS lines are skipped for plain HTTP.
func httpProbeDetails(s *stats.Statistics, verbose bool) string {
	if !verbose || !s.IsHTTP() || !s.HasHTTPResponse() {
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

// udpProbeFailureReason says what a failed UDP probe actually told us, e.g.
// " (port unreachable)". It is empty for the other probe types.
func udpProbeFailureReason(s *stats.Statistics) string {
	if !s.IsUDP() {
		return ""
	}

	if s.UDP.Rejected {
		return " (port unreachable)"
	}

	// Nothing came back at all. UDP gives us no way to tell an open port
	// that stays quiet from a packet that never arrived.
	return " (port may still be open)"
}

// udpProbeDetails is the indented block printed under a UDP probe line when
// verbose is on. Every probe sends its own number as the payload, so an
// echoed reply names the probe it belongs to and a gap in those numbers is
// exactly which probe was lost.
func udpProbeDetails(s *stats.Statistics, verbose bool) string {
	if !verbose || !s.IsUDP() {
		return ""
	}

	switch {
	case s.UDP.Echoed:
		return fmt.Sprintf("    reply echoed back probe %s\n", s.ProbeNumberStr())

	case s.UDP.ReplySize > 0:
		return fmt.Sprintf("    reply of %d bytes did not match probe %s\n", s.UDP.ReplySize, s.ProbeNumberStr())

	case s.UDP.Rejected:
		return fmt.Sprintf("    probe %s was refused\n", s.ProbeNumberStr())

	default:
		return fmt.Sprintf("    probe %s got no reply\n", s.ProbeNumberStr())
	}
}

// NewPrinter creates and returns an appropriate printer based on configuration
func NewPrinter(cfg config.PrinterConfig) (Printer, error) {
	switch {
	case cfg.OutputJSON:
		return NewJSONPrinter(cfg), nil

	case cfg.OutputDBPath != "":
		return NewDatabasePrinter(cfg)

	case cfg.OutputCSVPath != "":
		return NewCSVPrinter(cfg)

	case cfg.AlloyURL != "":
		return NewAlloyPrinter(cfg), nil

	case cfg.InfluxDBURL != "":
		return NewInfluxDBPrinter(cfg)

	case cfg.NoColor:
		return NewPlainPrinter(cfg), nil

	default:
		return NewColorPrinter(cfg), nil
	}
}
