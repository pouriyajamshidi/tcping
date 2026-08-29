package printers

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
)

// PlainPrinter provides functionality for printing messages in plain text (colorless).
type PlainPrinter struct{}

// NewPlainPrinter creates a new PlainPrinter instance.
func NewPlainPrinter() *PlainPrinter {
	return &PlainPrinter{}
}

// PrintStart prints the first message to indicate the target's address and port.
func (p *PlainPrinter) PrintStart(s *stats.Statistics) {
	fmt.Printf("TCPinging %s on port %d\n", s.Hostname, s.Port)
}

// PrintProbeSuccess prints a message when there is a successful probe response.
func (p *PlainPrinter) PrintProbeSuccess(s *stats.Statistics) {
	msg := "Reply from "

	if s.WithTimestamp {
		msg = fmt.Sprintf("%s %s", s.CurrentTimestamp(), msg)
	}

	target := s.IPStr()
	if s.Hostname != target {
		target = fmt.Sprintf("%s (%s)", s.Hostname, s.IPStr())
	}

	msg += fmt.Sprintf("%s on port %d ", target, s.Port)

	if s.WithSourceAddress {
		msg += fmt.Sprintf("using %s ", s.SourceAddr())
	}

	msg += fmt.Sprintf("TCP_conn=%d ", s.OngoingSuccessfulProbes)
	msg += fmt.Sprintf("time=%s ms\n", s.RTTStr())

	fmt.Print(msg)
}

// PrintProbeFailure prints a message the probe has failed.
func (p *PlainPrinter) PrintProbeFailure(s *stats.Statistics) {
	msg := "No reply from "

	if s.WithTimestamp {
		msg = fmt.Sprintf("%s %s", s.CurrentTimestamp(), msg)
	}

	target := s.IPStr()
	if s.Hostname != target {
		target = fmt.Sprintf("%s (%s)", s.Hostname, s.IPStr())
	}

	msg += fmt.Sprintf("%s on port %d ", target, s.Port)

	if s.WithSourceAddress && s.SourceAddr() != "" {
		msg += fmt.Sprintf("using %s ", s.SourceAddr())
	}

	msg += fmt.Sprintf("TCP_conn=%d\n", s.OngoingUnsuccessfulProbes)

	fmt.Print(msg)
}

// PrintStatistics prints the summary of all probe statistics.
func (p *PlainPrinter) PrintStatistics(s *stats.Statistics) {
	msg := fmt.Sprintf("\n--- %s ", s.Hostname)
	if !s.DestIsIP {
		msg = fmt.Sprintf("(%s) ", s.IP)
	}
	msg += "TCPing statistics ---\n"

	msg += fmt.Sprintf(
		"%d probes transmitted on port %d | %d received, ",
		s.TotalProbes(),
		s.Port,
		s.TotalSuccessfulProbes,
	)

	msg += fmt.Sprintf("%.2f%% packet loss\n", s.PacketLoss())

	msg += fmt.Sprintf("successful probes:   %d\n", s.TotalSuccessfulProbes)
	msg += fmt.Sprintf("unsuccessful probes: %d\n", s.TotalUnsuccessfulProbes)

	msg += "last successful probe:   "
	if s.LastSuccessfulProbe.IsZero() {
		msg += "Never succeeded\n"
	} else {
		msg += fmt.Sprintf("%s\n", s.LastSuccessfulProbeFormatted())
	}

	msg += "last unsuccessful probe: "
	if s.LastUnsuccessfulProbe.IsZero() {
		msg += "Never failed\n"
	} else {
		msg += fmt.Sprintf("%s\n", s.LastUnsuccessfulProbeFormatted())
	}

	msg += fmt.Sprintf("total uptime: %s\n", s.TotalUptimeDuration())
	msg += fmt.Sprintf("total downtime: %s\n", s.TotalDowntimeDuration())

	if s.LongestUp.Duration != 0 {
		msg += "longest consecutive uptime:   "
		msg += fmt.Sprintf("%s ", s.LongestUptimeDuration())
		msg += fmt.Sprintf("from %s ", s.LongestUptimeStartTime())
		msg += fmt.Sprintf("to %s\n", s.LongestUptimeEndTime())
	}

	if s.LongestDown.Duration != 0 {
		msg += fmt.Sprintf("longest consecutive downtime: %s ", s.LongestDowntimeDuration())
		msg += fmt.Sprintf("from %s ", s.LongestDowntimeStartTime())
		msg += fmt.Sprintf("to %s\n", s.LongestDowntimeEndTime())
	}

	if !s.DestIsIP {
		timeNoun := "time"
		if s.RetriedHostnameLookups != 1 {
			timeNoun = "times"
		}

		msg += fmt.Sprintf("retried to resolve hostname %d %s\n",
			s.RetriedHostnameLookups,
			timeNoun,
		)

		if len(s.HostnameChanges) > 1 {
			msg += "IP address changes:\n"
			for i := 0; i < len(s.HostnameChanges)-1; i++ {
				msg += fmt.Sprintf("  from %s ", s.HostnameChanges[i].Addr.String())
				msg += fmt.Sprintf("to %s ", s.HostnameChanges[i+1].Addr.String())
				msg += fmt.Sprintf("at %s\n", s.HostnameChanges[i+1].WhenFormatted())
			}
		}
	}

	if s.RTTResults.HasResults {
		msg += "rtt min/avg/max: "
		msg += fmt.Sprintf("%.3f/%.3f/%.3f ms\n",
			s.RTTResults.Min,
			s.RTTResults.Average,
			s.RTTResults.Max,
		)
	}

	msg += fmt.Sprint(strings.Repeat("-", 40) + "\n")
	msg += fmt.Sprintf("TCPing started at: %s\n", s.StartTimeFormatted())

	// If the program was not terminated, no need to show the end time
	if !s.EndTime.IsZero() {
		msg += fmt.Sprintf("TCPing ended at:   %s\n", s.EndTimeFormatted())
	}

	msg += fmt.Sprintf("duration (HH:MM:SS): %s\n\n", s.RuntimeDuration())

	fmt.Print(msg)
}

// PrintRetryingToResolve prints a message indicating that the program is retrying to resolve a hostname.
func (p *PlainPrinter) PrintRetryingToResolve(hostname string) {
	fmt.Printf("Retrying to resolve %s\n", hostname)
}

// PrintDownTimeDuration prints the total duration of downtime when no response was received.
func (p *PlainPrinter) PrintDownTimeDuration(s *stats.Statistics) {
	fmt.Printf("No response received for %s\n", s.DowntimeDuration())
}

// PrintError prints an error message. It takes a print verb and then the arguments.
func (p *PlainPrinter) PrintError(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

// Shutdown sets the end time, prints statistics, and exits the program.
func (p *PlainPrinter) Shutdown(s *stats.Statistics) {
	s.EndTime = time.Now()
	PrintStats(p, s)
	os.Exit(0)
}
