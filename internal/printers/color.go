package printers

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
)

// Color function aliases to use when printing information
var (
	printCyan        = color.Cyan.Printf
	printLightCyan   = color.LightCyan.Printf
	printGreen       = color.Green.Printf
	printLightGreen  = color.LightGreen.Printf
	printYellow      = color.Yellow.Printf
	printLightYellow = color.LightYellow.Printf
	printRed         = color.Red.Printf
	printLightBlue   = color.FgLightBlue.Printf
)

// ColorPrinter provides functionality for printing colored messages.
type ColorPrinter struct{}

// NewColorPrinter creates a new ColorPrinter instance.
func NewColorPrinter() *ColorPrinter {
	return &ColorPrinter{}
}

// PrintStart prints the first message to indicate the target's address and port.
func (p *ColorPrinter) PrintStart(s *stats.Statistics) {
	printLightCyan("TCPinging %s on port %d\n", s.Hostname, s.Port)
}

// PrintProbeSuccess prints a message when there is a successful probe response.
func (p *ColorPrinter) PrintProbeSuccess(s *stats.Statistics) {
	msg := "Reply from "

	if s.WithTimestamp {
		timestamp := s.StartTimeFormatted()
		msg = fmt.Sprintf("%s %s", timestamp, msg)
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

	printLightGreen(msg)
}

// PrintProbeFailure prints a message the probe has failed.
func (p *ColorPrinter) PrintProbeFailure(s *stats.Statistics) {
	msg := "No reply from "

	if s.WithTimestamp {
		timestamp := time.Now().Format(s.StartTimeFormatted())
		msg = fmt.Sprintf("%v %v", timestamp, msg)
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

	printRed(msg)
}

// PrintStatistics prints the summary of all probe statistics.
func (p *ColorPrinter) PrintStatistics(s *stats.Statistics) {
	printYellow("\n--- %s ", s.Hostname)
	if !s.DestIsIP {
		printYellow("(%s) ", s.IPStr())
	}
	printYellow("TCPing statistics ---\n")

	printYellow(
		"%d probes transmitted on port %d | %d received, ",
		s.TotalProbes(),
		s.Port,
		s.TotalSuccessfulProbes,
	)

	packetLoss := s.PacketLoss()

	switch {
	case packetLoss == 0:
		printGreen("%.2f%%", packetLoss)
	case packetLoss <= 30:
		printLightYellow("%.2f%%", packetLoss)
	default:
		printRed("%.2f%%", packetLoss)
	}

	printYellow(" packet loss\n")

	printYellow("successful probes:   ")
	printGreen("%d\n", s.TotalSuccessfulProbes)

	printYellow("unsuccessful probes: ")
	printRed("%d\n", s.TotalUnsuccessfulProbes)

	printYellow("last successful probe:   ")
	if s.LastSuccessfulProbe.IsZero() {
		printRed("Never succeeded\n")
	} else {
		printGreen("%s\n", s.LastSuccessfulProbeFormatted())
	}

	printYellow("last unsuccessful probe: ")
	if s.LastUnsuccessfulProbe.IsZero() {
		printGreen("Never failed\n")
	} else {
		printRed("%s\n", s.LastUnsuccessfulProbeFormatted())
	}

	printYellow("total uptime:   ")
	printGreen("%s\n", s.TotalUptimeDuration())
	printYellow("total downtime: ")
	printRed("%s\n", s.TotalDowntimeDuration())

	if s.LongestUp.Duration != 0 {
		printYellow("longest consecutive uptime:   ")
		printGreen("%s ", s.LongestUptimeDuration())
		printYellow("from ")
		printLightBlue("%s ", s.LongestUptimeStartTime())
		printYellow("to ")
		printLightBlue("%s\n", s.LongestUptimeEndTime())
	}

	if s.LongestDown.Duration != 0 {
		printYellow("longest consecutive downtime: ")
		printRed("%s ", s.LongestDowntimeDuration())
		printYellow("from ")
		printLightBlue("%s ", s.LongestDowntimeStartTime())
		printYellow("to ")
		printLightBlue("%s\n", s.LongestDowntimeEndTime())
	}

	if !s.DestIsIP {
		timeNoun := "time"
		if s.RetriedHostnameLookups != 1 {
			timeNoun = "times"
		}

		printYellow("retried to resolve hostname: %d %s\n",
			s.RetriedHostnameLookups,
			timeNoun,
		)

		if len(s.HostnameChanges) > 1 {
			printYellow("IP address changes:\n")
			for i := 0; i < len(s.HostnameChanges)-1; i++ {
				printYellow("  from ")
				printRed(s.HostnameChanges[i].Addr.String())
				printYellow(" to ")
				printGreen(s.HostnameChanges[i+1].Addr.String())
				printYellow(" at ")
				printLightBlue("%s\n", s.HostnameChanges[i+1].WhenFormatted())
			}
		}
	}

	if s.RTTResults.HasResults {
		printYellow("rtt ")
		printGreen("min")
		printYellow("/")
		printCyan("avg")
		printYellow("/")
		printRed("max: ")
		printGreen("%.3f", s.RTTResults.Min)
		printYellow("/")
		printCyan("%.3f", s.RTTResults.Average)
		printYellow("/")
		printRed("%.3f", s.RTTResults.Max)
		printYellow(" ms\n")
	}

	printYellow(strings.Repeat("-", 40) + "\n")
	printYellow("TCPing started at: %s\n", s.StartTimeFormatted())

	// If the program was not terminated, no need to show the end time
	if !s.EndTime.IsZero() {
		printYellow("TCPing ended at:   %s\n", s.EndTimeFormatted())
	}

	printYellow("duration (HH:MM:SS): %s\n\n", s.RuntimeDuration())
}

// PrintRetryingToResolve prints a message indicating that the program is retrying to resolve a hostname.
func (p *ColorPrinter) PrintRetryingToResolve(hostname string) {
	printLightYellow("Retrying to resolve %s\n", hostname)
}

// PrintDownTimeDuration prints the total duration of downtime when no response was received.
func (p *ColorPrinter) PrintDownTimeDuration(s *stats.Statistics) {
	printYellow("No response received for %s\n", s.DowntimeDuration())
}

// PrintError prints an error message in red. It takes a print verb and then the arguments.
func (p *ColorPrinter) PrintError(format string, args ...any) {
	printRed(format+"\n", args...)
}

// Shutdown sets the end time, prints statistics, and exits the program.
func (p *ColorPrinter) Shutdown(s *stats.Statistics) {
	s.EndTime = time.Now()
	PrintStats(p, s)
	os.Exit(0)
}
