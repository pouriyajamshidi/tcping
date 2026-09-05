// Package printers holds the places a run's results can go: the terminal,
// with or without color, a JSON, CSV or SQLite file, and the Alloy and
// InfluxDB endpoints. NewPrinter picks one based on the flags given.
package printers

import (
	"fmt"

	"github.com/pouriyajamshidi/tcping/v3/probe"
	"github.com/pouriyajamshidi/tcping/v3/stats"
)

// NewPrinter creates and returns an appropriate printer based on configuration
func NewPrinter(cfg Config) (probe.Printer, error) {
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
