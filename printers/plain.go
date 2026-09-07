package printers

import (
	"fmt"
	"strings"

	"github.com/pouriyajamshidi/tcping/v3/stats"
)

// PlainPrinter provides functionality for printing messages in plain text (colorless).
type PlainPrinter struct {
	cfg Config
}

// NewPlainPrinter creates a new PlainPrinter instance.
func NewPlainPrinter(cfg Config) *PlainPrinter {
	return &PlainPrinter{cfg: cfg}
}

// PrintStart prints the first message to indicate the target's address and port.
func (p *PlainPrinter) PrintStart(s *stats.Statistics) {
	if s.DestIsIP {
		fmt.Printf("Probing %s on port %d over %s\n", s.Hostname, s.Port, s.ProtocolStr())
		return
	}
	fmt.Printf("Probing %s on port %d over %s (resolved in %s ms)\n", s.Hostname, s.Port, s.ProtocolStr(), s.NameResolutionDurationStr())
}

// PrintNameResolutionDuration prints how long a hostname resolution retry took.
func (p *PlainPrinter) PrintNameResolutionDuration(s *stats.Statistics) {
	if s.ResolvedThisProbe {
		// Shown inline in PrintProbeSuccess/PrintProbeFailure instead.
		return
	}
	fmt.Printf("Resolved in %s ms\n", s.NameResolutionDurationStr())
}

// PrintProbeSuccess prints a message when there is a successful probe response.
func (p *PlainPrinter) PrintProbeSuccess(s *stats.Statistics) {
	msg := "Reply from "

	if p.cfg.WithTimestamp {
		msg = fmt.Sprintf("%s %s", s.CurrentTimestamp(), msg)
	}

	target := s.IPStr()
	if s.Hostname != target {
		target = fmt.Sprintf("%s (%s)", s.Hostname, s.IPStr())
	}

	msg += fmt.Sprintf("%s on port %d ", target, s.Port)

	if p.cfg.WithSourceAddress && s.SourceAddr() != "" {
		msg += fmt.Sprintf("using %s ", s.SourceAddr())
	}

	msg += fmt.Sprintf("%s_conn=%d", s.ProtocolStr(), s.OngoingSuccessfulProbes)
	msg += httpProbeSummary(s)
	msg += fmt.Sprintf(" time=%s ms", s.RTTStr())

	if s.ResolvedThisProbe {
		msg += fmt.Sprintf(" (resolved in %s ms)", s.NameResolutionDurationStr())
	}

	if s.EndedDowntime != 0 {
		msg += fmt.Sprintf(" (after %s of downtime)", s.EndedDowntimeDuration())
	}
	msg += "\n"
	msg += httpProbeDetails(s, p.cfg.Verbose)
	msg += udpProbeDetails(s, p.cfg.Verbose)

	fmt.Print(msg)
}

// PrintProbeFailure prints a message the probe has failed.
func (p *PlainPrinter) PrintProbeFailure(s *stats.Statistics) {
	msg := "No reply from "

	if p.cfg.WithTimestamp {
		msg = fmt.Sprintf("%s %s", s.CurrentTimestamp(), msg)
	}

	target := s.IPStr()
	if s.Hostname != target {
		target = fmt.Sprintf("%s (%s)", s.Hostname, s.IPStr())
	}

	msg += fmt.Sprintf("%s on port %d ", target, s.Port)

	if p.cfg.WithSourceAddress && s.SourceAddr() != "" {
		msg += fmt.Sprintf("using %s ", s.SourceAddr())
	}

	msg += fmt.Sprintf("%s_conn=%d", s.ProtocolStr(), s.OngoingUnsuccessfulProbes)
	msg += httpProbeSummary(s)
	msg += udpProbeFailureReason(s)

	if s.ResolvedThisProbe {
		msg += fmt.Sprintf(" (resolved in %s ms)", s.NameResolutionDurationStr())
	}

	if s.EndedUptime != 0 {
		msg += fmt.Sprintf(" (after %s of uptime)", s.EndedUptimeDuration())
	}
	msg += "\n"
	msg += httpProbeDetails(s, p.cfg.Verbose)
	msg += udpProbeDetails(s, p.cfg.Verbose)

	fmt.Print(msg)
}

// PrintStatistics prints the summary of all probe statistics.
func (p *PlainPrinter) PrintStatistics(s *stats.Statistics) {
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("\n--- %s ", s.Hostname))
	if !s.DestIsIP {
		msg.WriteString(fmt.Sprintf("(%s) ", s.IP))
	}
	msg.WriteString("TCPing statistics ---\n")

	msg.WriteString(fmt.Sprintf(
		"%d %s probes transmitted on port %d | %d received, ",
		s.TotalProbes(),
		s.ProtocolStr(),
		s.Port,
		s.TotalSuccessfulProbes,
	))

	msg.WriteString(fmt.Sprintf("%.2f%% packet loss\n", s.PacketLoss()))

	msg.WriteString(fmt.Sprintf("successful probes:   %d\n", s.TotalSuccessfulProbes))
	msg.WriteString(fmt.Sprintf("unsuccessful probes: %d\n", s.TotalUnsuccessfulProbes))

	msg.WriteString("last successful probe:   ")
	if s.LastSuccessfulProbe.IsZero() {
		msg.WriteString("Never succeeded\n")
	} else {
		msg.WriteString(fmt.Sprintf("%s\n", s.LastSuccessfulProbeFormatted()))
	}

	msg.WriteString("last unsuccessful probe: ")
	if s.LastUnsuccessfulProbe.IsZero() {
		msg.WriteString("Never failed\n")
	} else {
		msg.WriteString(fmt.Sprintf("%s\n", s.LastUnsuccessfulProbeFormatted()))
	}

	msg.WriteString(fmt.Sprintf("total uptime:   %s\n", s.TotalUptimeDuration()))
	msg.WriteString(fmt.Sprintf("total downtime: %s\n", s.TotalDowntimeDuration()))

	if s.LongestUptime.Duration != 0 {
		msg.WriteString("longest consecutive uptime:   ")
		msg.WriteString(fmt.Sprintf("%s ", s.LongestUptimeDuration()))
		msg.WriteString(fmt.Sprintf("from %s ", s.LongestUptimeStartTime()))
		msg.WriteString(fmt.Sprintf("to %s\n", s.LongestUptimeEndTime()))
	}

	if s.LongestDowntime.Duration != 0 {
		msg.WriteString(fmt.Sprintf("longest consecutive downtime: %s ", s.LongestDowntimeDuration()))
		msg.WriteString(fmt.Sprintf("from %s ", s.LongestDowntimeStartTime()))
		msg.WriteString(fmt.Sprintf("to %s\n", s.LongestDowntimeEndTime()))
	}

	if !s.DestIsIP {
		timeNoun := "time"
		if s.RetriedHostnameLookups != 1 {
			timeNoun = "times"
		}

		msg.WriteString(fmt.Sprintf("retried to resolve hostname: %d %s\n",
			s.RetriedHostnameLookups,
			timeNoun,
		))

		if len(s.HostnameChanges) > 1 {
			msg.WriteString("IP address changes:\n")
			for i := 0; i < len(s.HostnameChanges)-1; i++ {
				msg.WriteString(fmt.Sprintf("  from %s ", s.HostnameChanges[i].Addr.String()))
				msg.WriteString(fmt.Sprintf("to %s ", s.HostnameChanges[i+1].Addr.String()))
				msg.WriteString(fmt.Sprintf("at %s ", s.HostnameChanges[i+1].WhenFormatted()))
				msg.WriteString(fmt.Sprintf("took %s ms\n", s.HostnameChanges[i+1].DurationStr()))
			}
		}
	}

	if s.TotalSuccessfulProbes > 0 {
		msg.WriteString("rtt min/avg/max/mdev: ")
		msg.WriteString(fmt.Sprintf("%.3f/%.3f/%.3f/%.3f ms\n",
			s.RTTResults.Min,
			s.RTTResults.Average,
			s.RTTResults.Max,
			s.RTTResults.Mdev,
		))
	}

	msg.WriteString(fmt.Sprint(strings.Repeat("-", 40) + "\n"))
	msg.WriteString(fmt.Sprintf("TCPing started at: %s\n", s.StartTimeFormatted()))

	// If the program was not terminated, no need to show the end time
	if !s.EndTime.IsZero() {
		msg.WriteString(fmt.Sprintf("TCPing ended at:   %s\n", s.EndTimeFormatted()))
	}

	msg.WriteString(fmt.Sprintf("duration (HH:MM:SS): %s\n\n", s.RuntimeDuration()))

	fmt.Print(msg.String())
}

// PrintRetryingToResolve prints a message indicating that the program is retrying to resolve a hostname.
func (p *PlainPrinter) PrintRetryingToResolve(hostname string) {
	fmt.Printf("Retrying to resolve %s\n", hostname)
}

// PrintError prints an error message. It takes a print verb and then the arguments.
func (p *PlainPrinter) PrintError(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

// Shutdown prints statistics. Statistics are already finalized by
// finalizeStatistics by the time this runs. It does not exit the program -
// that decision belongs to the caller, not the printer.
func (p *PlainPrinter) Shutdown(s *stats.Statistics) {
	if !p.cfg.OmitStatistics {
		p.PrintStatistics(s)
	}
}
