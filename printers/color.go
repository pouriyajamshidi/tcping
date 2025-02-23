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
	cfg PrinterConfig
}

// NewColorPrinter creates a new ColorPrinter instance.
// The showTimestamp parameter controls whether timestamps should be included in printed messages.
func NewColorPrinter(cfg PrinterConfig) *ColorPrinter {
	return &ColorPrinter{cfg: cfg}
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

// PrintProbeSuccess prints a message indicating a successful probe response.
// It includes the source IP (if shown), target IP/hostname, port, connection streak, and RTT.
//
// Parameters:
//   - sourceAddr: The local address used for the TCP connection.
//   - userInput: The user-provided input data (hostname, IP, port, etc.).
//   - streak: The number of consecutive successful probes.
//   - rtt: The round-trip time of the probe in milliseconds (3 decimal points).
func (p *ColorPrinter) PrintProbeSuccess(startTime time.Time, sourceAddr string, opts types.Options, streak uint, rtt string) {
	if p.cfg.ShowFailuresOnly {
		return
	}

	timestamp := ""
	if p.cfg.WithTimestamp {
		timestamp = startTime.Format(consts.TimeFormat)
	}

	if opts.Hostname == opts.IP.String() {
		if timestamp == "" {
			if p.cfg.WithSourceAddress {
				consts.ColorLightGreen("Reply from %s on port %d using %s TCP_conn=%d time=%s ms\n",
					opts.IP.String(),
					opts.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				consts.ColorLightGreen("Reply from %s on port %d TCP_conn=%d time=%s ms\n",
					opts.IP.String(),
					opts.Port,
					streak,
					rtt)
			}
		} else {
			if p.cfg.WithSourceAddress {
				consts.ColorLightGreen("%s Reply from %s on port %d using %s TCP_conn=%d time=%s ms\n",
					timestamp,
					opts.IP.String(),
					opts.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				consts.ColorLightGreen("%s Reply from %s on port %d TCP_conn=%d time=%s ms\n",
					timestamp,
					opts.IP.String(),
					opts.Port,
					streak,
					rtt)
			}
		}
	} else {
		if timestamp == "" {
			if p.cfg.WithSourceAddress {
				consts.ColorLightGreen("Reply from %s (%s) on port %d using %s TCP_conn=%d time=%s ms\n",
					opts.Hostname,
					opts.IP.String(),
					opts.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				consts.ColorLightGreen("Reply from %s (%s) on port %d TCP_conn=%d time=%s ms\n",
					opts.Hostname,
					opts.IP.String(),
					opts.Port,
					streak,
					rtt)
			}
		} else {
			if p.cfg.WithSourceAddress {
				consts.ColorLightGreen("%s Reply from %s (%s) on port %d using %s TCP_conn=%d time=%s ms\n",
					timestamp,
					opts.Hostname,
					opts.IP.String(),
					opts.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				consts.ColorLightGreen("%s Reply from %s (%s) on port %d TCP_conn=%d time=%s ms\n",
					timestamp,
					opts.Hostname,
					opts.IP.String(),
					opts.Port,
					streak,
					rtt)
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
func (p *ColorPrinter) PrintProbeFailure(startTime time.Time, opts types.Options, streak uint) {
	timestamp := ""
	if p.cfg.WithTimestamp {
		timestamp = startTime.Format(consts.TimeFormat)
	}

	if opts.Hostname == "" {
		if timestamp == "" {
			consts.ColorRed("No reply from %s on port %d TCP_conn=%d\n",
				opts.IP,
				opts.Port,
				streak)
		} else {
			consts.ColorRed("%s No reply from %s on port %d TCP_conn=%d\n",
				timestamp,
				opts.IP,
				opts.Port,
				streak)
		}
	} else {
		if timestamp == "" {
			consts.ColorRed("No reply from %s (%s) on port %d TCP_conn=%d\n",
				opts.Hostname,
				opts.IP,
				opts.Port,
				streak)
		} else {
			consts.ColorRed("%s No reply from %s (%s) on port %d TCP_conn=%d\n",
				timestamp,
				opts.Hostname,
				opts.IP,
				opts.Port,
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
	consts.ColorLightYellow("Retrying to resolve %s\n", hostname)
}

// PrintError prints an error message in red.
//
// Parameters:
//   - format: A format string for the error message.
//   - args: Arguments to format the message.
func (p *ColorPrinter) PrintError(format string, args ...any) {
	consts.ColorRed(format+"\n", args...)
}

// PrintStatistics prints a summary of TCP ping statistics.
// It includes transmitted and received packets, packet loss percentage,
// successful and unsuccessful probes, uptime/downtime durations,
// longest uptime/downtime, IP address changes, and RTT statistics.
func (p *ColorPrinter) PrintStatistics(t types.Tcping) {
	if !t.DestIsIP {
		consts.ColorYellow("\n--- %s (%s) TCPing statistics ---\n",
			t.Options.Hostname,
			t.Options.IP)
	} else {
		consts.ColorYellow("\n--- %s TCPing statistics ---\n", t.Options.Hostname)
	}

	totalPackets := t.TotalSuccessfulProbes + t.TotalUnsuccessfulProbes

	consts.ColorYellow("%d probes transmitted on port %d | ", totalPackets, t.Options.Port)
	consts.ColorYellow("%d received, ", t.TotalSuccessfulProbes)

	packetLoss := (float32(t.TotalUnsuccessfulProbes) / float32(totalPackets)) * 100

	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	if packetLoss == 0 {
		consts.ColorGreen("%.2f%%", packetLoss)
	} else if packetLoss > 0 && packetLoss <= 30 {
		consts.ColorLightYellow("%.2f%%", packetLoss)
	} else {
		consts.ColorRed("%.2f%%", packetLoss)
	}

	consts.ColorYellow(" packet loss\n")

	consts.ColorYellow("successful probes:   ")
	consts.ColorGreen("%d\n", t.TotalSuccessfulProbes)

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

	consts.ColorYellow("total uptime: ")
	consts.ColorGreen("  %s\n", utils.DurationToString(t.TotalUptime))
	consts.ColorYellow("total downtime: ")
	consts.ColorRed("%s\n", utils.DurationToString(t.TotalDowntime))

	if t.LongestUptime.Duration != 0 {
		uptime := utils.DurationToString(t.LongestUptime.Duration)

		consts.ColorYellow("longest consecutive uptime:   ")
		consts.ColorGreen("%v ", uptime)
		consts.ColorYellow("from ")
		consts.ColorLightBlue("%v ", t.LongestUptime.Start.Format(consts.TimeFormat))
		consts.ColorYellow("to ")
		consts.ColorLightBlue("%v\n", t.LongestUptime.End.Format(consts.TimeFormat))
	}

	if t.LongestDowntime.Duration != 0 {
		downtime := utils.DurationToString(t.LongestDowntime.Duration)

		consts.ColorYellow("longest consecutive downtime: ")
		consts.ColorRed("%v ", downtime)
		consts.ColorYellow("from ")
		consts.ColorLightBlue("%v ", t.LongestDowntime.Start.Format(consts.TimeFormat))
		consts.ColorYellow("to ")
		consts.ColorLightBlue("%v\n", t.LongestDowntime.End.Format(consts.TimeFormat))
	}

	if !t.DestIsIP {
		timeNoun := "time"
		if t.RetriedHostnameLookups > 1 {
			timeNoun = "times"
		}

		consts.ColorYellow("retried to resolve hostname ")
		consts.ColorRed("%d ", t.RetriedHostnameLookups)
		consts.ColorYellow("%s\n", timeNoun)

		if len(t.HostnameChanges) > 1 {
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
