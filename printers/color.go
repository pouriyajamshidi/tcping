// Package printers contains the logic for printing information
package printers

import (
	"math"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/consts"
	"github.com/pouriyajamshidi/tcping/v2/internal/utils"
	"github.com/pouriyajamshidi/tcping/v2/types"
)

// ColorPrinter provides functionality for printing messages with color support.
// It optionally includes a timestamp in the output if ShowTimestamp is enabled.
type ColorPrinter struct {
	ShowTimestamp bool
}

// NewColorPrinter creates a new ColorPrinter instance.
// The showTimestamp parameter controls whether timestamps should be included in printed messages.
func NewColorPrinter(showTimestamp bool) *ColorPrinter {
	return &ColorPrinter{ShowTimestamp: showTimestamp}
}

// PrintStart prints a message indicating the start of a TCP ping attempt.
// The message is printed in light cyan and includes the target hostname and port.
//
// Parameters:
//   - hostname: The target host for the TCP ping.
//   - port: The target port number.
func (p *ColorPrinter) PrintStart(hostname string, port uint16) {
	consts.ColorLightCyan("TCPinging %s on port %d\n", hostname, port)
}

// PrintStatistics prints a summary of TCP ping statistics.
// It includes transmitted and received packets, packet loss percentage,
// successful and unsuccessful probes, uptime/downtime durations,
// longest uptime/downtime, IP address changes, and RTT statistics.
func (p *ColorPrinter) PrintStatistics(t types.Tcping) {
	totalPackets := t.TotalSuccessfulProbes + t.TotalUnsuccessfulProbes
	packetLoss := (float32(t.TotalUnsuccessfulProbes) / float32(totalPackets)) * 100

	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	/* general stats */
	if !t.DestIsIP {
		consts.ColorYellow("\n--- %s (%s) TCPing statistics ---\n", t.Options.Hostname, t.Options.IP)
	} else {
		consts.ColorYellow("\n--- %s TCPing statistics ---\n", t.Options.Hostname)
	}
	consts.ColorYellow("%d probes transmitted on port %d | ", totalPackets, t.Options.Port)
	consts.ColorYellow("%d received, ", t.TotalSuccessfulProbes)

	/* packet loss stats */
	if packetLoss == 0 {
		consts.ColorGreen("%.2f%%", packetLoss)
	} else if packetLoss > 0 && packetLoss <= 30 {
		consts.ColorLightYellow("%.2f%%", packetLoss)
	} else {
		consts.ColorRed("%.2f%%", packetLoss)
	}

	consts.ColorYellow(" packet loss\n")

	/* successful packet stats */
	consts.ColorYellow("successful probes:   ")
	consts.ColorGreen("%d\n", t.TotalSuccessfulProbes)

	/* unsuccessful packet stats */
	consts.ColorYellow("unsuccessful probes: ")
	consts.ColorRed("%d\n", t.TotalUnsuccessfulProbes)

	consts.ColorYellow("last successful probe:   ")
	if t.LastSuccessfulProbe.IsZero() {
		consts.ColorRed("Never succeeded\n")
	} else {
		consts.ColorGreen("%v\n", t.LastSuccessfulProbe.Format(consts.TimeFormat))
	}

	consts.ColorYellow("last unsuccessful probe: ")
	if t.LastUnsuccessfulProbe.IsZero() {
		consts.ColorGreen("Never failed\n")
	} else {
		consts.ColorRed("%v\n", t.LastUnsuccessfulProbe.Format(consts.TimeFormat))
	}

	/* uptime and downtime stats */
	consts.ColorYellow("total uptime: ")
	consts.ColorGreen("  %s\n", utils.DurationToString(t.TotalUptime))
	consts.ColorYellow("total downtime: ")
	consts.ColorRed("%s\n", utils.DurationToString(t.TotalDowntime))

	/* longest uptime stats */
	if t.LongestUptime.Duration != 0 {
		uptime := utils.DurationToString(t.LongestUptime.Duration)

		consts.ColorYellow("longest consecutive uptime:   ")
		consts.ColorGreen("%v ", uptime)
		consts.ColorYellow("from ")
		consts.ColorLightBlue("%v ", t.LongestUptime.Start.Format(consts.TimeFormat))
		consts.ColorYellow("to ")
		consts.ColorLightBlue("%v\n", t.LongestUptime.End.Format(consts.TimeFormat))
	}

	/* longest downtime stats */
	if t.LongestDowntime.Duration != 0 {
		downtime := utils.DurationToString(t.LongestDowntime.Duration)

		consts.ColorYellow("longest consecutive downtime: ")
		consts.ColorRed("%v ", downtime)
		consts.ColorYellow("from ")
		consts.ColorLightBlue("%v ", t.LongestDowntime.Start.Format(consts.TimeFormat))
		consts.ColorYellow("to ")
		consts.ColorLightBlue("%v\n", t.LongestDowntime.End.Format(consts.TimeFormat))
	}

	/* resolve retry stats */
	if !t.DestIsIP {
		consts.ColorYellow("retried to resolve hostname ")
		consts.ColorRed("%d ", t.RetriedHostnameLookups)
		consts.ColorYellow("times\n")

		if len(t.HostnameChanges) >= 2 {
			consts.ColorYellow("IP address changes:\n")
			for i := 0; i < len(t.HostnameChanges)-1; i++ {
				consts.ColorYellow("  from ")
				consts.ColorRed(t.HostnameChanges[i].Addr.String())
				consts.ColorYellow(" to ")
				consts.ColorGreen(t.HostnameChanges[i+1].Addr.String())
				consts.ColorYellow(" at ")
				consts.ColorLightBlue("%v\n", t.HostnameChanges[i+1].When.Format(consts.TimeFormat))
			}
		}
	}

	if t.RttResults.HasResults {
		consts.ColorYellow("rtt ")
		consts.ColorGreen("min")
		consts.ColorYellow("/")
		consts.ColorCyan("avg")
		consts.ColorYellow("/")
		consts.ColorRed("max: ")
		consts.ColorGreen("%.3f", t.RttResults.Min)
		consts.ColorYellow("/")
		consts.ColorCyan("%.3f", t.RttResults.Average)
		consts.ColorYellow("/")
		consts.ColorRed("%.3f", t.RttResults.Max)
		consts.ColorYellow(" ms\n")
	}

	consts.ColorYellow("--------------------------------------\n")
	consts.ColorYellow("TCPing started at: %v\n", t.StartTime.Format(consts.TimeFormat))

	/* If the program was not terminated, no need to show the end time */
	if !t.EndTime.IsZero() {
		consts.ColorYellow("TCPing ended at:   %v\n", t.EndTime.Format(consts.TimeFormat))
	}

	durationTime := time.Time{}.Add(t.TotalDowntime + t.TotalUptime)
	consts.ColorYellow("duration (HH:MM:SS): %v\n\n", durationTime.Format(consts.HourFormat))
}

// PrintProbeSuccess prints a message indicating a successful probe response.
// It includes the source IP (if shown), target IP/hostname, port, connection streak, and RTT.
//
// Parameters:
//   - sourceAddr: The local address used for the TCP connection.
//   - userInput: The user-provided input data (hostname, IP, port, etc.).
//   - streak: The number of consecutive successful probes.
//   - rtt: The round-trip time of the probe in milliseconds.
func (p *ColorPrinter) PrintProbeSuccess(sourceAddr string, userInput types.Options, streak uint, rtt float32) {
	timestamp := ""
	if p.ShowTimestamp {
		timestamp = time.Now().Format(consts.TimeFormat)
	}
	if userInput.Hostname == "" {
		if timestamp == "" {
			if userInput.ShowSourceAddress {
				consts.ColorLightGreen("Reply from %s on port %d using %s TCP_conn=%d time=%.3f ms\n",
					userInput.IP.String(),
					userInput.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				consts.ColorLightGreen("Reply from %s on port %d TCP_conn=%d time=%.3f ms\n",
					userInput.IP.String(),
					userInput.Port,
					streak,
					rtt)
			}
		} else {
			if userInput.ShowSourceAddress {
				consts.ColorLightGreen("%s Reply from %s on port %d using %s TCP_conn=%d time=%.3f ms\n",
					timestamp,
					userInput.IP.String(),
					userInput.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				consts.ColorLightGreen("%s Reply from %s on port %d TCP_conn=%d time=%.3f ms\n",
					timestamp,
					userInput.IP.String(),
					userInput.Port,
					streak,
					rtt)
			}
		}
	} else {
		if timestamp == "" {
			if userInput.ShowSourceAddress {
				consts.ColorLightGreen("Reply from %s (%s) on port %d using %s TCP_conn=%d time=%.3f ms\n",
					userInput.Hostname,
					userInput.IP.String(),
					userInput.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				consts.ColorLightGreen("Reply from %s (%s) on port %d TCP_conn=%d time=%.3f ms\n",
					userInput.Hostname,
					userInput.IP.String(),
					userInput.Port,
					streak,
					rtt)
			}
		} else {
			if userInput.ShowSourceAddress {
				consts.ColorLightGreen("%s Reply from %s (%s) on port %d using %s TCP_conn=%d time=%.3f ms\n",
					timestamp,
					userInput.Hostname,
					userInput.IP.String(),
					userInput.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				consts.ColorLightGreen("%s Reply from %s (%s) on port %d TCP_conn=%d time=%.3f ms\n",
					timestamp,
					userInput.Hostname,
					userInput.IP.String(),
					userInput.Port,
					streak,
					rtt)
			}
		}
	}
}

// PrintProbeFail prints a message indicating a failed probe attempt.
// It includes the target hostname/IP, port, and failed connection streak.
//
// Parameters:
//   - userInput: The user-provided input data (hostname, IP, port, etc.).
//   - streak: The number of consecutive failed probes.
func (p *ColorPrinter) PrintProbeFail(userInput types.Options, streak uint) {
	timestamp := ""
	if p.ShowTimestamp {
		timestamp = time.Now().Format(consts.TimeFormat)
	}
	if userInput.Hostname == "" {
		if timestamp == "" {
			consts.ColorRed("No reply from %s on port %d TCP_conn=%d\n",
				userInput.IP,
				userInput.Port,
				streak)
		} else {
			consts.ColorRed("%s No reply from %s on port %d TCP_conn=%d\n",
				timestamp,
				userInput.IP,
				userInput.Port,
				streak)
		}
	} else {
		if timestamp == "" {
			consts.ColorRed("No reply from %s (%s) on port %d TCP_conn=%d\n",
				userInput.Hostname,
				userInput.IP,
				userInput.Port,
				streak)
		} else {
			consts.ColorRed("%s No reply from %s (%s) on port %d TCP_conn=%d\n",
				timestamp,
				userInput.Hostname,
				userInput.IP,
				userInput.Port,
				streak)
		}
	}
}

// PrintTotalDownTime prints the total duration of downtime when no response was received.
//
// Parameters:
//   - downtime: The total duration of downtime.
func (p *ColorPrinter) PrintTotalDownTime(downtime time.Duration) {
	consts.ColorYellow("No response received for %s\n", utils.DurationToString(downtime))
}

// PrintRetryingToResolve prints a message indicating that the program is retrying to resolve a hostname.
//
// Parameters:
//   - hostname: The hostname that is being resolved.
func (p *ColorPrinter) PrintRetryingToResolve(hostname string) {
	consts.ColorLightYellow("retrying to resolve %s\n", hostname)
}

// PrintInfo prints an informational message in light blue.
//
// Parameters:
//   - format: A format string for the message.
//   - args: Arguments to format the message.
func (p *ColorPrinter) PrintInfo(format string, args ...any) {
	consts.ColorLightBlue(format+"\n", args...)
}

// PrintError prints an error message in red.
//
// Parameters:
//   - format: A format string for the error message.
//   - args: Arguments to format the message.
func (p *ColorPrinter) PrintError(format string, args ...any) {
	consts.ColorRed(format+"\n", args...)
}
