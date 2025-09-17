package printers

import (
	"fmt"
	"math"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/utils"
	"github.com/pouriyajamshidi/tcping/v3/types"
)

// PlainPrinter is a printer that prints the TCPing results in a simple, plain text format.
type PlainPrinter struct {
	cfg PrinterConfig
}

// NewPlainPrinter creates a new PlainPrinter instance with an optional timestamp setting.
func NewPlainPrinter(cfg PrinterConfig) *PlainPrinter {
	return &PlainPrinter{cfg: cfg}
}

// PrintStart prints the start message indicating the TCPing operation on the given hostname and port.
func (p *PlainPrinter) PrintStart(hostname string, port uint16) {
	fmt.Printf("TCPinging %s on port %d\n", hostname, port)
}

// PrintProbeSuccess prints a success message for a probe, including round-trip time and streak info.
func (p *PlainPrinter) PrintProbeSuccess(startTime time.Time, sourceAddr string, opts types.Options, streak uint, rtt string) {
	timestamp := ""
	if p.cfg.WithTimestamp {
		timestamp = startTime.Format(time.DateTime)
	}

	if opts.Hostname == "" {
		if timestamp == "" {
			if p.cfg.WithSourceAddress {
				fmt.Printf("Reply from %s on port %d using %s TCP_conn=%d time=%s ms\n",
					opts.IP.String(),
					opts.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				fmt.Printf("Reply from %s on port %d TCP_conn=%d time=%s ms\n",
					opts.IP.String(),
					opts.Port,
					streak,
					rtt)
			}
		} else {
			if p.cfg.WithSourceAddress {
				fmt.Printf("%s Reply from %s on port %d using %s TCP_conn=%d time=%s ms\n",
					timestamp,
					opts.IP.String(),
					opts.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				fmt.Printf("%s Reply from %s on port %d TCP_conn=%d time=%s ms\n",
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
				fmt.Printf("Reply from %s (%s) on port %d using %s TCP_conn=%d time=%s ms\n",
					opts.Hostname,
					opts.IP.String(),
					opts.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				fmt.Printf("Reply from %s (%s) on port %d TCP_conn=%d time=%s ms\n",
					opts.Hostname,
					opts.IP.String(),
					opts.Port,
					streak,
					rtt)
			}
		} else {
			if p.cfg.WithSourceAddress {
				fmt.Printf("%s Reply from %s (%s) on port %d using %s TCP_conn=%d time=%s ms\n",
					timestamp,
					opts.Hostname,
					opts.IP.String(),
					opts.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				fmt.Printf("%s Reply from %s (%s) on port %d TCP_conn=%d time=%s ms\n",
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

// PrintProbeFailure prints a failure message for a probe.
func (p *PlainPrinter) PrintProbeFailure(startTime time.Time, opts types.Options, streak uint) {
	timestamp := ""
	if p.cfg.WithTimestamp {
		timestamp = startTime.Format(time.DateTime)
	}

	if opts.Hostname == "" {
		if timestamp == "" {
			fmt.Printf("No reply from %s on port %d TCP_conn=%d\n",
				opts.IP,
				opts.Port,
				streak)
		} else {
			fmt.Printf("%s No reply from %s on port %d TCP_conn=%d\n",
				timestamp,
				opts.IP,
				opts.Port,
				streak)
		}
	} else {
		if timestamp == "" {
			fmt.Printf("No reply from %s (%s) on port %d TCP_conn=%d\n",
				opts.Hostname,
				opts.IP,
				opts.Port,
				streak)
		} else {
			fmt.Printf("%s No reply from %s (%s) on port %d TCP_conn=%d\n",
				timestamp,
				opts.Hostname,
				opts.IP,
				opts.Port,
				streak)
		}
	}
}

// PrintTotalDownTime prints the total downtime when no response is received.
func (p *PlainPrinter) PrintTotalDownTime(downtime time.Duration) {
	fmt.Printf("No response received for %s\n", utils.DurationToString(downtime))
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
func (p *PlainPrinter) PrintStatistics(t types.Tcping) {
	totalPackets := t.TotalSuccessfulProbes + t.TotalUnsuccessfulProbes
	packetLoss := (float32(t.TotalUnsuccessfulProbes) / float32(totalPackets)) * 100

	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	if !t.DestIsIP {
		fmt.Printf("\n--- %s (%s) TCPing statistics ---\n",
			t.Options.Hostname,
			t.Options.IP)
	} else {
		fmt.Printf("\n--- %s TCPing statistics ---\n", t.Options.Hostname)
	}
	fmt.Printf("%d probes transmitted on port %d | %d received",
		totalPackets,
		t.Options.Port,
		t.TotalSuccessfulProbes)

	fmt.Printf("%.2f%% packet loss\n", packetLoss)
	fmt.Printf("successful probes:   %d\n", t.TotalSuccessfulProbes)
	fmt.Printf("unsuccessful probes: %d\n", t.TotalUnsuccessfulProbes)

	fmt.Printf("last successful probe:   ")
	if t.LastSuccessfulProbe.IsZero() {
		fmt.Printf("Never succeeded\n")
	} else {
		fmt.Printf("%v\n", t.LastSuccessfulProbe.Format(time.DateTime))
	}

	fmt.Printf("last unsuccessful probe: ")
	if t.LastUnsuccessfulProbe.IsZero() {
		fmt.Printf("Never failed\n")
	} else {
		fmt.Printf("%v\n", t.LastUnsuccessfulProbe.Format(time.DateTime))
	}

	fmt.Printf("total uptime: %s\n", utils.DurationToString(t.TotalUptime))
	fmt.Printf("total downtime: %s\n", utils.DurationToString(t.TotalDowntime))

	if t.LongestUptime.Duration != 0 {
		uptime := utils.DurationToString(t.LongestUptime.Duration)

		fmt.Printf("longest consecutive uptime:   ")
		fmt.Printf("%v ", uptime)
		fmt.Printf("from %v ", t.LongestUptime.Start.Format(time.DateTime))
		fmt.Printf("to %v\n", t.LongestUptime.End.Format(time.DateTime))
	}

	if t.LongestDowntime.Duration != 0 {
		downtime := utils.DurationToString(t.LongestDowntime.Duration)

		fmt.Printf("longest consecutive downtime: %v ", downtime)
		fmt.Printf("from %v ", t.LongestDowntime.Start.Format(time.DateTime))
		fmt.Printf("to %v\n", t.LongestDowntime.End.Format(time.DateTime))
	}

	if !t.DestIsIP {
		timeNoun := "time"
		if t.RetriedHostnameLookups > 1 {
			timeNoun = "times"
		}

		fmt.Printf("retried to resolve hostname %d %s\n",
			t.RetriedHostnameLookups,
			timeNoun)

		if len(t.HostnameChanges) >= 2 {
			fmt.Printf("IP address changes:\n")
			for i := 0; i < len(t.HostnameChanges)-1; i++ {
				fmt.Printf("  from %s", t.HostnameChanges[i].Addr.String())
				fmt.Printf(" to %s", t.HostnameChanges[i+1].Addr.String())
				fmt.Printf(" at %v\n", t.HostnameChanges[i+1].When.Format(time.DateTime))
			}
		}
	}

	if t.RttResults.HasResults {
		fmt.Printf("rtt min/avg/max: ")
		fmt.Printf("%.3f/%.3f/%.3f ms\n",
			t.RttResults.Min,
			t.RttResults.Average,
			t.RttResults.Max)
	}

	fmt.Printf("--------------------------------------\n")
	fmt.Printf("TCPing started at: %v\n", t.StartTime.Format(time.DateTime))

	/* If the program was not terminated, no need to show the end time */
	if !t.EndTime.IsZero() {
		fmt.Printf("TCPing ended at:   %v\n", t.EndTime.Format(time.DateTime))
	}

	durationTime := time.Time{}.Add(t.TotalDowntime + t.TotalUptime)
	fmt.Printf("duration (HH:MM:SS): %v\n\n", durationTime.Format(time.TimeOnly))
}
