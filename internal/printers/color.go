package printers

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/gookit/color"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
	"github.com/pouriyajamshidi/tcping/v3/internal/utils"
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

// PrintStart prints a message indicating the start of a TCP ping attempt.
func (p *ColorPrinter) PrintStart(s *stats.Statistics) {
	printLightCyan("TCPinging %s on port %d\n", s.Hostname, s.Port)
}

// PrintProbeSuccess prints a message indicating a successful probe response.
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

// PrintProbeFailure prints a message indicating a failed probe attempt.
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

// PrintTotalDownTime prints the total duration of downtime when no response was received.
//
// Parameters:
//   - downtime: The total duration of downtime.
func (p *ColorPrinter) PrintTotalDownTime(s *stats.Statistics) {
	printYellow("No response received for %s\n", utils.DurationToString(s.DownTime))
}

// PrintRetryingToResolve prints a message indicating that the program is retrying to resolve a hostname.
//
// Parameters:
//   - hostname: The hostname that is being resolved.
func (p *ColorPrinter) PrintRetryingToResolve(hostname string) {
	printLightYellow("Retrying to resolve %s\n", hostname)
}

// PrintError prints an error message in red.
//
// Parameters:
//   - format: A format string for the error message.
//   - args: Arguments to format the message.
func (p *ColorPrinter) PrintError(format string, args ...any) {
	printRed(format+"\n", args...)
}

// PrintStatistics prints a summary of TCP ping statistics.
// It includes transmitted and received packets, packet loss percentage,
// successful and unsuccessful probes, uptime/downtime durations,
// longest uptime/downtime, IP address changes, and RTT statistics.
func (p *ColorPrinter) PrintStatistics(s *stats.Statistics) {
	if !s.DestIsIP {
		printYellow("\n--- %s (%s) TCPing statistics ---\n",
			s.Hostname,
			s.IPStr())
	} else {
		printYellow("\n--- %s TCPing statistics ---\n", s.Hostname)
	}

	totalPackets := s.TotalSuccessfulProbes + s.TotalUnsuccessfulProbes

	printYellow("%d probes transmitted on port %d | ", totalPackets, s.Port)
	printYellow("%d received, ", s.TotalSuccessfulProbes)

	packetLoss := (float32(s.TotalUnsuccessfulProbes) / float32(totalPackets)) * 100

	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	if packetLoss == 0 {
		printGreen("%.2f%%", packetLoss)
	} else if packetLoss > 0 && packetLoss <= 30 {
		printLightYellow("%.2f%%", packetLoss)
	} else {
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
		printGreen("%v\n", s.LastSuccessfulProbe.Format(time.DateTime))
	}

	printYellow("last unsuccessful probe: ")
	if s.LastUnsuccessfulProbe.IsZero() {
		printGreen("Never failed\n")
	} else {
		printRed("%v\n", s.LastUnsuccessfulProbe.Format(time.DateTime))
	}

	printYellow("total uptime: ")
	printGreen("  %s\n", utils.DurationToString(s.TotalUptime))
	printYellow("total downtime: ")
	printRed("%s\n", utils.DurationToString(s.TotalDowntime))

	if s.LongestUp.Duration != 0 {
		uptime := utils.DurationToString(s.LongestUp.Duration)

		printYellow("longest consecutive uptime:   ")
		printGreen("%v ", uptime)
		printYellow("from ")
		printLightBlue("%v ", s.LongestUp.Start.Format(time.DateTime))
		printYellow("to ")
		printLightBlue("%v\n", s.LongestUp.End.Format(time.DateTime))
	}

	if s.LongestDown.Duration != 0 {
		downtime := utils.DurationToString(s.LongestDown.Duration)

		printYellow("longest consecutive downtime: ")
		printRed("%v ", downtime)
		printYellow("from ")
		printLightBlue("%v ", s.LongestDown.Start.Format(time.DateTime))
		printYellow("to ")
		printLightBlue("%v\n", s.LongestDown.End.Format(time.DateTime))
	}

	if !s.DestIsIP {
		timeNoun := "time"
		if s.RetriedHostnameLookups > 1 {
			timeNoun = "times"
		}

		printYellow("retried to resolve hostname ")
		printRed("%d ", s.RetriedHostnameLookups)
		printYellow("%s\n", timeNoun)

		if len(s.HostnameChanges) > 1 {
			printYellow("IP address changes:\n")
			for i := 0; i < len(s.HostnameChanges)-1; i++ {
				printYellow("  from ")
				printRed(s.HostnameChanges[i].Addr.String())
				printYellow(" to ")
				printGreen(s.HostnameChanges[i+1].Addr.String())
				printYellow(" at ")
				printLightBlue("%v\n", s.HostnameChanges[i+1].When.Format(time.DateTime))
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

	printYellow("--------------------------------------\n")
	printYellow("TCPing started at: %v\n", s.StartTimeFormatted())

	/* If the program was not terminated, no need to show the end time */
	if !s.EndTime.IsZero() {
		printYellow("TCPing ended at:   %v\n", s.EndTimeFormatted())
	}

	durationTime := time.Time{}.Add(s.TotalDowntime + s.TotalUptime)
	printYellow("duration (HH:MM:SS): %v\n\n", durationTime.Format(time.TimeOnly))
}

// Shutdown sets the end time, prints statistics, and exits the program.
func (p *ColorPrinter) Shutdown(s *stats.Statistics) {
	s.EndTime = time.Now()
	PrintStats(p, s)
	os.Exit(0)
}
