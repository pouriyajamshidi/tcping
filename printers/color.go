// Package printers contains the logic for printing information
package printers

import (
	"math"
	"os"
	"time"

	"github.com/gookit/color"
	"github.com/pouriyajamshidi/tcping/v3/statistics"
)

// Color functions used when printing information
var (
	ColorCyan        = color.Cyan.Printf
	ColorLightCyan   = color.LightCyan.Printf
	ColorGreen       = color.Green.Printf
	ColorLightGreen  = color.LightGreen.Printf
	ColorYellow      = color.Yellow.Printf
	ColorLightYellow = color.LightYellow.Printf
	ColorRed         = color.Red.Printf
	ColorLightBlue   = color.FgLightBlue.Printf
)

// ColorPrinter provides functionality for printing messages with color support.
// It optionally includes a timestamp in the output if ShowTimestamp is enabled.
type ColorPrinter struct{}

// NewColorPrinter creates a new ColorPrinter instance.
// The showTimestamp parameter controls whether timestamps should be included in printed messages.
func NewColorPrinter() *ColorPrinter {
	return &ColorPrinter{}
}

// Shutdown sets the end time, prints statistics, and exits the program.
func (p *ColorPrinter) Shutdown(s *statistics.Statistics) {
	s.EndTime = time.Now()
	if s.DestWasDown {
		statistics.SetLongestDuration(s.StartOfDowntime, time.Since(s.StartOfDowntime), &s.LongestDowntime)
	} else {
		statistics.SetLongestDuration(s.StartOfUptime, time.Since(s.StartOfUptime), &s.LongestUptime)
	}

	s.RTTResults = statistics.CalcMinAvgMaxRttTime(s.RTT)
	p.PrintStatistics(s)
	os.Exit(0)
}

// PrintStart prints a message indicating the start of a TCP ping attempt.
// The message is printed in light cyan and includes the target hostname and port.
//
// Parameters:
//   - hostname: The target host for the TCP ping.
//   - port: The target port number.
func (p *ColorPrinter) PrintStart(s *statistics.Statistics) {
	ColorLightCyan("TCPinging %s on port %d\n", s.Hostname, s.Port)
}

// PrintProbeSuccess prints a message indicating a successful probe response.
// It includes the source IP (if shown), target IP/hostname, port, connection streak, and RTT.
//
// Parameters:
//   - sourceAddr: The local address used for the TCP connection.
//   - userInput: The user-provided input data (hostname, IP, port, etc.).
//   - streak: The number of consecutive successful probes.
//   - rtt: The round-trip time of the probe in milliseconds (3 decimal points).
func (p *ColorPrinter) PrintProbeSuccess(s *statistics.Statistics) {
	timestamp := ""
	if s.WithTimestamp {
		timestamp = s.StartTimeFormatted()
	}

	if s.Hostname == s.IPStr() {
		if timestamp == "" {
			if s.WithSourceAddress {
				ColorLightGreen("Reply from %s on port %d using %s TCP_conn=%d time=%s ms\n",
					s.IP.String(),
					s.Port,
					s.SourceAddr(),
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			} else {
				ColorLightGreen("Reply from %s on port %d TCP_conn=%d time=%s ms\n",
					s.IP.String(),
					s.Port,
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			}
		} else {
			if s.WithSourceAddress {
				ColorLightGreen("%s Reply from %s on port %d using %s TCP_conn=%d time=%s ms\n",
					timestamp,
					s.IP.String(),
					s.Port,
					s.SourceAddr(),
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			} else {
				ColorLightGreen("%s Reply from %s on port %d TCP_conn=%d time=%s ms\n",
					timestamp,
					s.IP.String(),
					s.Port,
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			}
		}
	} else {
		if timestamp == "" {
			if s.WithSourceAddress {
				ColorLightGreen("Reply from %s (%s) on port %d using %s TCP_conn=%d time=%s ms\n",
					s.Hostname,
					s.IP.String(),
					s.Port,
					s.SourceAddr(),
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			} else {
				ColorLightGreen("Reply from %s (%s) on port %d TCP_conn=%d time=%s ms\n",
					s.Hostname,
					s.IP.String(),
					s.Port,
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			}
		} else {
			if s.WithSourceAddress {
				ColorLightGreen("%s Reply from %s (%s) on port %d using %s TCP_conn=%d time=%s ms\n",
					timestamp,
					s.Hostname,
					s.IP.String(),
					s.Port,
					s.SourceAddr(),
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			} else {
				ColorLightGreen("%s Reply from %s (%s) on port %d TCP_conn=%d time=%s ms\n",
					timestamp,
					s.Hostname,
					s.IP.String(),
					s.Port,
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			}
		}
	}
}

// PrintProbeFailure prints a message indicating a failed probe attempt.
// It includes the target hostname/IP, port, and failed connection streak.
//
// Parameters:
//   - userInput: The user-provided input data (hostname, IP, port, etc.).
//   - streak: The number of consecutive failed probes.
func (p *ColorPrinter) PrintProbeFailure(s *statistics.Statistics) {
	timestamp := ""
	if s.WithTimestamp {
		timestamp = s.StartTimeFormatted()
	}

	if s.Hostname == "" {
		if timestamp == "" {
			ColorRed("No reply from %s on port %d TCP_conn=%d\n",
				s.IP,
				s.Port,
				s.OngoingUnsuccessfulProbes)
		} else {
			ColorRed("%s No reply from %s on port %d TCP_conn=%d\n",
				timestamp,
				s.IP,
				s.Port,
				s.OngoingUnsuccessfulProbes)
		}
	} else {
		if timestamp == "" {
			ColorRed("No reply from %s (%s) on port %d TCP_conn=%d\n",
				s.Hostname,
				s.IP,
				s.Port,
				s.OngoingUnsuccessfulProbes)
		} else {
			ColorRed("%s No reply from %s (%s) on port %d TCP_conn=%d\n",
				timestamp,
				s.Hostname,
				s.IP,
				s.Port,
				s.OngoingUnsuccessfulProbes)
		}
	}
}

// PrintTotalDownTime prints the total duration of downtime when no response was received.
//
// Parameters:
//   - downtime: The total duration of downtime.
func (p *ColorPrinter) PrintTotalDownTime(s *statistics.Statistics) {
	ColorYellow("No response received for %s\n", statistics.DurationToString(s.DownTime))
}

// PrintRetryingToResolve prints a message indicating that the program is retrying to resolve a hostname.
//
// Parameters:
//   - hostname: The hostname that is being resolved.
func (p *ColorPrinter) PrintRetryingToResolve(s *statistics.Statistics) {
	ColorLightYellow("Retrying to resolve %s\n", s.Hostname)
}

// PrintError prints an error message in red.
//
// Parameters:
//   - format: A format string for the error message.
//   - args: Arguments to format the message.
func (p *ColorPrinter) PrintError(format string, args ...any) {
	ColorRed(format+"\n", args...)
}

// PrintStatistics prints a summary of TCP ping statistics.
// It includes transmitted and received packets, packet loss percentage,
// successful and unsuccessful probes, uptime/downtime durations,
// longest uptime/downtime, IP address changes, and RTT statistics.
func (p *ColorPrinter) PrintStatistics(s *statistics.Statistics) {
	if !s.DestIsIP {
		ColorYellow("\n--- %s (%s) TCPing statistics ---\n",
			s.Hostname,
			s.IPStr())
	} else {
		ColorYellow("\n--- %s TCPing statistics ---\n", s.Hostname)
	}

	totalPackets := s.TotalSuccessfulProbes + s.TotalUnsuccessfulProbes

	ColorYellow("%d probes transmitted on port %d | ", totalPackets, s.Port)
	ColorYellow("%d received, ", s.TotalSuccessfulProbes)

	packetLoss := (float32(s.TotalUnsuccessfulProbes) / float32(totalPackets)) * 100

	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	switch {
	case packetLoss == 0:
		ColorGreen("%.2f%%", packetLoss)
	case packetLoss > 0 && packetLoss <= 30:
		ColorLightYellow("%.2f%%", packetLoss)
	default:
		ColorRed("%.2f%%", packetLoss)
	}

	ColorYellow(" packet loss\n")

	ColorYellow("successful probes:   ")
	ColorGreen("%d\n", s.TotalSuccessfulProbes)

	ColorYellow("unsuccessful probes: ")
	ColorRed("%d\n", s.TotalUnsuccessfulProbes)

	ColorYellow("last successful probe:   ")
	if s.LastSuccessfulProbe.IsZero() {
		ColorRed("Never succeeded\n")
	} else {
		ColorGreen("%v\n", s.LastSuccessfulProbe.Format(time.DateTime))
	}

	ColorYellow("last unsuccessful probe: ")
	if s.LastUnsuccessfulProbe.IsZero() {
		ColorGreen("Never failed\n")
	} else {
		ColorRed("%v\n", s.LastUnsuccessfulProbe.Format(time.DateTime))
	}

	ColorYellow("total uptime: ")
	ColorGreen("  %s\n", statistics.DurationToString(s.TotalUptime))
	ColorYellow("total downtime: ")
	ColorRed("%s\n", statistics.DurationToString(s.TotalDowntime))

	if s.LongestUp.Duration != 0 {
		uptime := statistics.DurationToString(s.LongestUp.Duration)

		ColorYellow("longest consecutive uptime:   ")
		ColorGreen("%v ", uptime)
		ColorYellow("from ")
		ColorLightBlue("%v ", s.LongestUp.Start.Format(time.DateTime))
		ColorYellow("to ")
		ColorLightBlue("%v\n", s.LongestUp.End.Format(time.DateTime))
	}

	if s.LongestDown.Duration != 0 {
		downtime := statistics.DurationToString(s.LongestDown.Duration)

		ColorYellow("longest consecutive downtime: ")
		ColorRed("%v ", downtime)
		ColorYellow("from ")
		ColorLightBlue("%v ", s.LongestDown.Start.Format(time.DateTime))
		ColorYellow("to ")
		ColorLightBlue("%v\n", s.LongestDown.End.Format(time.DateTime))
	}

	if !s.DestIsIP {
		timeNoun := "time"
		if s.RetriedHostnameLookups > 1 {
			timeNoun = "times"
		}

		ColorYellow("retried to resolve hostname ")
		ColorRed("%d ", s.RetriedHostnameLookups)
		ColorYellow("%s\n", timeNoun)

		if len(s.HostnameChanges) > 1 {
			ColorYellow("IP address changes:\n")
			for i := 0; i < len(s.HostnameChanges)-1; i++ {
				ColorYellow("  from ")
				ColorRed(s.HostnameChanges[i].Addr.String())
				ColorYellow(" to ")
				ColorGreen(s.HostnameChanges[i+1].Addr.String())
				ColorYellow(" at ")
				ColorLightBlue("%v\n", s.HostnameChanges[i+1].When.Format(time.DateTime))
			}
		}
	}

	if s.RTTResults.HasResults {
		ColorYellow("rtt ")
		ColorGreen("min")
		ColorYellow("/")
		ColorCyan("avg")
		ColorYellow("/")
		ColorRed("max: ")
		ColorGreen("%.3f", s.RTTResults.Min)
		ColorYellow("/")
		ColorCyan("%.3f", s.RTTResults.Average)
		ColorYellow("/")
		ColorRed("%.3f", s.RTTResults.Max)
		ColorYellow(" ms\n")
	}

	ColorYellow("--------------------------------------\n")
	ColorYellow("TCPing started at: %v\n", s.StartTimeFormatted())

	/* If the program was not terminated, no need to show the end time */
	if !s.EndTime.IsZero() {
		ColorYellow("TCPing ended at:   %v\n", s.EndTimeFormatted())
	}

	durationTime := time.Time{}.Add(s.TotalDowntime + s.TotalUptime)
	ColorYellow("duration (HH:MM:SS): %v\n\n", durationTime.Format(time.TimeOnly))
}
