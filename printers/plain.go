package printers

import (
	"fmt"
	"math"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/consts"
	"github.com/pouriyajamshidi/tcping/v2/internal/utils"
	"github.com/pouriyajamshidi/tcping/v2/types"
)

// PlainPrinter is a printer that prints the TCPing results in a simple, plain text format.
type PlainPrinter struct {
	showTimestamp bool
}

// NewPlainPrinter creates a new PlainPrinter instance with an optional timestamp setting.
func NewPlainPrinter(showTimestamp bool) *PlainPrinter {
	return &PlainPrinter{showTimestamp: showTimestamp}
}

// PrintStart prints the start message indicating the TCPing operation on the given hostname and port.
func (p *PlainPrinter) PrintStart(hostname string, port uint16) {
	fmt.Printf("TCPinging %s on port %d\n", hostname, port)
}

// PrintStatistics prints detailed statistics about the TCPing session.
func (p *PlainPrinter) PrintStatistics(t types.Tcping) {
	totalPackets := t.TotalSuccessfulProbes + t.TotalUnsuccessfulProbes
	packetLoss := (float32(t.TotalUnsuccessfulProbes) / float32(totalPackets)) * 100

	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	/* general stats */
	if !t.DestIsIP {
		fmt.Printf("\n--- %s (%s) TCPing statistics ---\n", t.Options.Hostname, t.Options.IP)
	} else {
		fmt.Printf("\n--- %s TCPing statistics ---\n", t.Options.Hostname)
	}
	fmt.Printf("%d probes transmitted on port %d | %d received", totalPackets, t.Options.Port, t.TotalSuccessfulProbes)

	/* packet loss stats */
	fmt.Printf("%.2f%% packet loss\n", packetLoss)

	/* successful packet stats */
	fmt.Printf("successful probes:   %d\n", t.TotalSuccessfulProbes)

	/* unsuccessful packet stats */
	fmt.Printf("unsuccessful probes: %d\n", t.TotalUnsuccessfulProbes)

	fmt.Printf("last successful probe:   ")
	if t.LastSuccessfulProbe.IsZero() {
		fmt.Printf("Never succeeded\n")
	} else {
		fmt.Printf("%v\n", t.LastSuccessfulProbe.Format(consts.TimeFormat))
	}

	fmt.Printf("last unsuccessful probe: ")
	if t.LastUnsuccessfulProbe.IsZero() {
		fmt.Printf("Never failed\n")
	} else {
		fmt.Printf("%v\n", t.LastUnsuccessfulProbe.Format(consts.TimeFormat))
	}

	/* uptime and downtime stats */
	fmt.Printf("total uptime: %s\n", utils.DurationToString(t.TotalUptime))
	fmt.Printf("total downtime: %s\n", utils.DurationToString(t.TotalDowntime))

	/* longest uptime stats */
	if t.LongestUptime.Duration != 0 {
		uptime := utils.DurationToString(t.LongestUptime.Duration)

		fmt.Printf("longest consecutive uptime:   ")
		fmt.Printf("%v ", uptime)
		fmt.Printf("from %v ", t.LongestUptime.Start.Format(consts.TimeFormat))
		fmt.Printf("to %v\n", t.LongestUptime.End.Format(consts.TimeFormat))
	}

	/* longest downtime stats */
	if t.LongestDowntime.Duration != 0 {
		downtime := utils.DurationToString(t.LongestDowntime.Duration)

		fmt.Printf("longest consecutive downtime: %v ", downtime)
		fmt.Printf("from %v ", t.LongestDowntime.Start.Format(consts.TimeFormat))
		fmt.Printf("to %v\n", t.LongestDowntime.End.Format(consts.TimeFormat))
	}

	/* resolve retry stats */
	if !t.DestIsIP {
		fmt.Printf("retried to resolve hostname %d times\n", t.RetriedHostnameLookups)

		if len(t.HostnameChanges) >= 2 {
			fmt.Printf("IP address changes:\n")
			for i := 0; i < len(t.HostnameChanges)-1; i++ {
				fmt.Printf("  from %s", t.HostnameChanges[i].Addr.String())
				fmt.Printf(" to %s", t.HostnameChanges[i+1].Addr.String())
				fmt.Printf(" at %v\n", t.HostnameChanges[i+1].When.Format(consts.TimeFormat))
			}
		}
	}

	if t.RttResults.HasResults {
		fmt.Printf("rtt min/avg/max: ")
		fmt.Printf("%.3f/%.3f/%.3f ms\n", t.RttResults.Min, t.RttResults.Average, t.RttResults.Max)
	}

	fmt.Printf("--------------------------------------\n")
	fmt.Printf("TCPing started at: %v\n", t.StartTime.Format(consts.TimeFormat))

	/* If the program was not terminated, no need to show the end time */
	if !t.EndTime.IsZero() {
		fmt.Printf("TCPing ended at:   %v\n", t.EndTime.Format(consts.TimeFormat))
	}

	durationTime := time.Time{}.Add(t.TotalDowntime + t.TotalUptime)
	fmt.Printf("duration (HH:MM:SS): %v\n\n", durationTime.Format(consts.HourFormat))
}

// PrintProbeSuccess prints a success message for a probe, including round-trip time and streak info.
func (p *PlainPrinter) PrintProbeSuccess(sourceAddr string, userInput types.Options, streak uint, rtt float32) {
	timestamp := ""
	if p.showTimestamp {
		timestamp = time.Now().Format(consts.TimeFormat)
	}
	if userInput.Hostname == "" {
		if timestamp == "" {
			if userInput.ShowSourceAddress {
				fmt.Printf("Reply from %s on port %d using %s TCP_conn=%d time=%.3f ms\n", userInput.IP.String(), userInput.Port, sourceAddr, streak, rtt)
			} else {
				fmt.Printf("Reply from %s on port %d TCP_conn=%d time=%.3f ms\n", userInput.IP.String(), userInput.Port, streak, rtt)
			}
		} else {
			if userInput.ShowSourceAddress {
				fmt.Printf("%s Reply from %s on port %d using %s TCP_conn=%d time=%.3f ms\n", timestamp, userInput.IP.String(), userInput.Port, sourceAddr, streak, rtt)
			} else {
				fmt.Printf("%s Reply from %s on port %d TCP_conn=%d time=%.3f ms\n", timestamp, userInput.IP.String(), userInput.Port, streak, rtt)
			}
		}
	} else {
		if timestamp == "" {
			if userInput.ShowSourceAddress {
				fmt.Printf("Reply from %s (%s) on port %d using %s TCP_conn=%d time=%.3f ms\n", userInput.Hostname, userInput.IP.String(), userInput.Port, sourceAddr, streak, rtt)
			} else {
				fmt.Printf("Reply from %s (%s) on port %d TCP_conn=%d time=%.3f ms\n", userInput.Hostname, userInput.IP.String(), userInput.Port, streak, rtt)
			}
		} else {
			if userInput.ShowSourceAddress {
				fmt.Printf("%s Reply from %s (%s) on port %d using %s TCP_conn=%d time=%.3f ms\n", timestamp, userInput.Hostname, userInput.IP.String(), userInput.Port, sourceAddr, streak, rtt)
			} else {
				fmt.Printf("%s Reply from %s (%s) on port %d TCP_conn=%d time=%.3f ms\n", timestamp, userInput.Hostname, userInput.IP.String(), userInput.Port, streak, rtt)
			}
		}
	}
}

// PrintProbeFail prints a failure message for a probe.
func (p *PlainPrinter) PrintProbeFail(userInput types.Options, streak uint) {
	timestamp := ""
	if p.showTimestamp {
		timestamp = time.Now().Format(consts.TimeFormat)
	}
	if userInput.Hostname == "" {
		if timestamp == "" {
			fmt.Printf("No reply from %s on port %d TCP_conn=%d\n", userInput.IP, userInput.Port, streak)
		} else {
			fmt.Printf("%s No reply from %s on port %d TCP_conn=%d\n", timestamp, userInput.IP, userInput.Port, streak)
		}
	} else {
		if timestamp == "" {
			fmt.Printf("No reply from %s (%s) on port %d TCP_conn=%d\n", userInput.Hostname, userInput.IP, userInput.Port, streak)
		} else {
			fmt.Printf("%s No reply from %s (%s) on port %d TCP_conn=%d\n", timestamp, userInput.Hostname, userInput.IP, userInput.Port, streak)
		}
	}
}

// PrintTotalDownTime prints the total downtime when no response is received.
func (p *PlainPrinter) PrintTotalDownTime(downtime time.Duration) {
	fmt.Printf("No response received for %s\n", utils.DurationToString(downtime))
}

// PrintRetryingToResolve prints a message indicating that the program is retrying to resolve the hostname.
func (p *PlainPrinter) PrintRetryingToResolve(hostname string) {
	fmt.Printf("retrying to resolve %s\n", hostname)
}

// PrintInfo prints general informational messages.
func (p *PlainPrinter) PrintInfo(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

// PrintError prints error messages.
func (p *PlainPrinter) PrintError(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}
