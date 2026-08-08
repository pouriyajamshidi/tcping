package printers

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
	"github.com/pouriyajamshidi/tcping/v3/internal/utils"
)

// PlainPrinter is a printer that prints the TCPing results in a simple, plain text format.
type PlainPrinter struct{}

// NewPlainPrinter creates a new PlainPrinter instance with an optional timestamp setting.
func NewPlainPrinter() *PlainPrinter {
	return &PlainPrinter{}
}

// Shutdown sets the end time, prints statistics, and exits the program.
func (p *PlainPrinter) Shutdown(s *stats.Statistics) {
	s.EndTime = time.Now()
	PrintStats(p, s)
	os.Exit(0)
}

// PrintStart prints the start message indicating the TCPing operation on the given hostname and port.
func (p *PlainPrinter) PrintStart(s *stats.Statistics) {
	fmt.Printf("TCPinging %s on port %d\n", s.Hostname, s.Port)
}

// PrintProbeSuccess prints a success message for a probe, including round-trip time and streak info.
func (p *PlainPrinter) PrintProbeSuccess(s *stats.Statistics) {
	msg := "Reply from "

	if s.WithTimestamp {
		timestamp := s.StartTimeFormatted()
		msg = fmt.Sprintf("%v %v", timestamp, msg)
	}

	hostnameAndIP := s.IPStr()
	if s.Hostname != hostnameAndIP {
		hostnameAndIP = fmt.Sprintf("%s (%s)", s.Hostname, s.IPStr())
	}

	msg += fmt.Sprintf("%s on port %d", hostnameAndIP, s.Port)

	if s.WithSourceAddress {
		msg += fmt.Sprintf(" using %s", s.SourceAddr())
	}

	msg += fmt.Sprintf(" TCP_conn=%d time=%s ms\n", s.OngoingSuccessfulProbes, s.RTTStr())

	fmt.Print(msg)
}

// PrintProbeFailure prints a failure message for a probe.
func (p *PlainPrinter) PrintProbeFailure(s *stats.Statistics) {
	msg := "No reply from "

	if s.WithTimestamp {
		timestamp := time.Now().Format(s.StartTimeFormatted())
		msg = fmt.Sprintf("%v %v", timestamp, msg)
	}

	hostnameAndIP := s.IPStr()
	if s.Hostname != hostnameAndIP {
		hostnameAndIP = fmt.Sprintf("%s (%s)", s.Hostname, s.IPStr())
	}

	msg += fmt.Sprintf("%s on port %s TCP_conn=%d\n", hostnameAndIP, s.PortStr(), s.OngoingUnsuccessfulProbes)

	fmt.Print(msg)
}

// PrintTotalDownTime prints the total downtime when no response is received.
func (p *PlainPrinter) PrintTotalDownTime(s *stats.Statistics) {
	fmt.Printf("No response received for %s\n", utils.DurationToString(s.DownTime))
}

// PrintRetryingToResolve prints a message indicating that the program is retrying to resolve the hostname.
func (p *PlainPrinter) PrintRetryingToResolve(hostname string) {
	fmt.Printf("Retrying to resolve %s\n", hostname)
}

// PrintError prints error messages.
func (p *PlainPrinter) PrintError(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

// PrintStatistics prints detailed statistics about the TCPing session.
func (p *PlainPrinter) PrintStatistics(s *stats.Statistics) {
	msg := fmt.Sprintf("\n--- %s TCPing statistics ---\n", s.Hostname)

	if !s.DestIsIP {
		msg = fmt.Sprintf("\n--- %s (%s) TCPing statistics ---\n",
			s.Hostname,
			s.IP)
	}

	totalPackets := s.TotalSuccessfulProbes + s.TotalUnsuccessfulProbes

	msg += fmt.Sprintf("%d probes transmitted on port %d | %d received",
		totalPackets,
		s.Port,
		s.TotalSuccessfulProbes)

	packetLoss := (float32(s.TotalUnsuccessfulProbes) / float32(totalPackets)) * 100

	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	msg += fmt.Sprintf("%.2f%% packet loss\n", packetLoss)
	msg += fmt.Sprintf("successful probes:   %d\n", s.TotalSuccessfulProbes)
	msg += fmt.Sprintf("unsuccessful probes: %d\n", s.TotalUnsuccessfulProbes)

	msg += fmt.Sprintf("last successful probe:   ")
	if s.LastSuccessfulProbe.IsZero() {
		msg += fmt.Sprintf("Never succeeded\n")
	} else {
		msg += fmt.Sprintf("%v\n", s.LastSuccessfulProbe.Format(time.DateTime))
	}

	msg += fmt.Sprintf("last unsuccessful probe: ")
	if s.LastUnsuccessfulProbe.IsZero() {
		msg += fmt.Sprintf("Never failed\n")
	} else {
		msg += fmt.Sprintf("%v\n", s.LastUnsuccessfulProbe.Format(time.DateTime))
	}

	msg += fmt.Sprintf("total uptime: %s\n", utils.DurationToString(s.TotalUptime))
	msg += fmt.Sprintf("total downtime: %s\n", utils.DurationToString(s.TotalDowntime))

	if s.LongestUp.Duration != 0 {
		uptime := utils.DurationToString(s.LongestUp.Duration)

		msg += fmt.Sprintf("longest consecutive uptime:   ")
		msg += fmt.Sprintf("%v ", uptime)
		msg += fmt.Sprintf("from %v ", s.LongestUp.Start.Format(time.DateTime))
		msg += fmt.Sprintf("to %v\n", s.LongestUp.End.Format(time.DateTime))
	}

	if s.LongestDown.Duration != 0 {
		downtime := utils.DurationToString(s.LongestDown.Duration)

		msg += fmt.Sprintf("longest consecutive downtime: %v ", downtime)
		msg += fmt.Sprintf("from %v ", s.LongestDown.Start.Format(time.DateTime))
		msg += fmt.Sprintf("to %v\n", s.LongestDown.End.Format(time.DateTime))
	}

	if !s.DestIsIP {
		timeNoun := "time"
		if s.RetriedHostnameLookups > 1 {
			timeNoun = "times"
		}

		msg += fmt.Sprintf("retried to resolve hostname %d %s\n",
			s.RetriedHostnameLookups,
			timeNoun)

		if len(s.HostnameChanges) >= 2 {
			msg += fmt.Sprintf("IP address changes:\n")
			for i := 0; i < len(s.HostnameChanges)-1; i++ {
				msg += fmt.Sprintf("  from %s", s.HostnameChanges[i].Addr.String())
				msg += fmt.Sprintf(" to %s", s.HostnameChanges[i+1].Addr.String())
				msg += fmt.Sprintf(" at %v\n", s.HostnameChanges[i+1].When.Format(time.DateTime))
			}
		}
	}

	if s.RTTResults.HasResults {
		msg += fmt.Sprintf("rtt min/avg/max: ")
		msg += fmt.Sprintf("%.3f/%.3f/%.3f ms\n",
			s.RTTResults.Min,
			s.RTTResults.Average,
			s.RTTResults.Max)
	}

	msg += fmt.Sprintf("--------------------------------------\n")
	msg += fmt.Sprintf("TCPing started at: %v\n", s.StartTime.Format(time.DateTime))

	/* If the program was not terminated, no need to show the end time */
	if !s.EndTime.IsZero() {
		msg += fmt.Sprintf("TCPing ended at:   %v\n", s.EndTime.Format(time.DateTime))
	}

	durationTime := time.Time{}.Add(s.TotalDowntime + s.TotalUptime)
	msg += fmt.Sprintf("duration (HH:MM:SS): %v\n\n", durationTime.Format(time.TimeOnly))

	fmt.Print(msg)
}
