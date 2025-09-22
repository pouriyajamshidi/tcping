package printers

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/utils"
	"github.com/pouriyajamshidi/tcping/v3/probes/statistics"
)

// PlainPrinter is a printer that prints the TCPing results in a simple, plain text format.
type PlainPrinter struct{}

// NewPlainPrinter creates a new PlainPrinter instance with an optional timestamp setting.
func NewPlainPrinter() *PlainPrinter {
	return &PlainPrinter{}
}

// Shutdown sets the end time, prints statistics, and exits the program.
func (p *PlainPrinter) Shutdown(s *statistics.Statistics) {
	s.EndTime = time.Now()
	PrintStats(p, s)
	os.Exit(0)
}

// PrintStart prints the start message indicating the TCPing operation on the given hostname and port.
func (p *PlainPrinter) PrintStart(s *statistics.Statistics) {
	fmt.Printf("TCPinging %s on port %d\n", s.Hostname, s.Port)
}

// PrintProbeSuccess prints a success message for a probe, including round-trip time and streak info.
func (p *PlainPrinter) PrintProbeSuccess(s *statistics.Statistics) {
	timestamp := ""
	if s.WithTimestamp {
		timestamp = s.StartTimeFormatted()
	}

	if s.Hostname == "" {
		if timestamp == "" {
			if s.WithSourceAddress {
				fmt.Printf("Reply from %s on port %d using %s TCP_conn=%d time=%s ms\n",
					s.IP.String(),
					s.Port,
					s.SourceAddr(),
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			} else {
				fmt.Printf("Reply from %s on port %d TCP_conn=%d time=%s ms\n",
					s.IP.String(),
					s.Port,
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			}
		} else {
			if s.WithSourceAddress {
				fmt.Printf("%s Reply from %s on port %d using %s TCP_conn=%d time=%s ms\n",
					timestamp,
					s.IP.String(),
					s.Port,
					s.SourceAddr(),
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			} else {
				fmt.Printf("%s Reply from %s on port %d TCP_conn=%d time=%s ms\n",
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
				fmt.Printf("Reply from %s (%s) on port %d using %s TCP_conn=%d time=%s ms\n",
					s.Hostname,
					s.IP.String(),
					s.Port,
					s.SourceAddr(),
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			} else {
				fmt.Printf("Reply from %s (%s) on port %d TCP_conn=%d time=%s ms\n",
					s.Hostname,
					s.IP.String(),
					s.Port,
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			}
		} else {
			if s.WithSourceAddress {
				fmt.Printf("%s Reply from %s (%s) on port %d using %s TCP_conn=%d time=%s ms\n",
					timestamp,
					s.Hostname,
					s.IP.String(),
					s.Port,
					s.SourceAddr(),
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			} else {
				fmt.Printf("%s Reply from %s (%s) on port %d TCP_conn=%d time=%s ms\n",
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

// PrintProbeFailure prints a failure message for a probe.
func (p *PlainPrinter) PrintProbeFailure(s *statistics.Statistics) {
	timestamp := ""
	if s.WithTimestamp {
		timestamp = s.StartTimeFormatted()
	}

	if s.Hostname == "" {
		if timestamp == "" {
			fmt.Printf("No reply from %s on port %d TCP_conn=%d\n",
				s.IP,
				s.Port,
				s.OngoingUnsuccessfulProbes)
		} else {
			fmt.Printf("%s No reply from %s on port %d TCP_conn=%d\n",
				timestamp,
				s.IP,
				s.Port,
				s.OngoingUnsuccessfulProbes)
		}
	} else {
		if timestamp == "" {
			fmt.Printf("No reply from %s (%s) on port %d TCP_conn=%d\n",
				s.Hostname,
				s.IP,
				s.Port,
				s.OngoingUnsuccessfulProbes)
		} else {
			fmt.Printf("%s No reply from %s (%s) on port %d TCP_conn=%d\n",
				timestamp,
				s.Hostname,
				s.IP,
				s.Port,
				s.OngoingUnsuccessfulProbes)
		}
	}
}

// PrintTotalDownTime prints the total downtime when no response is received.
func (p *PlainPrinter) PrintTotalDownTime(s *statistics.Statistics) {
	fmt.Printf("No response received for %s\n", utils.DurationToString(s.DownTime))
}

// PrintRetryingToResolve prints a message indicating that the program is retrying to resolve the hostname.
func (p *PlainPrinter) PrintRetryingToResolve(s *statistics.Statistics) {
	fmt.Printf("Retrying to resolve %s\n", s.Hostname)
}

// PrintError prints error messages.
func (p *PlainPrinter) PrintError(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

// PrintStatistics prints detailed statistics about the TCPing session.
func (p *PlainPrinter) PrintStatistics(s *statistics.Statistics) {
	if !s.DestIsIP {
		fmt.Printf("\n--- %s (%s) TCPing statistics ---\n",
			s.Hostname,
			s.IP)
	} else {
		fmt.Printf("\n--- %s TCPing statistics ---\n", s.Hostname)
	}

	totalPackets := s.TotalSuccessfulProbes + s.TotalUnsuccessfulProbes

	fmt.Printf("%d probes transmitted on port %d | %d received",
		totalPackets,
		s.Port,
		s.TotalSuccessfulProbes)

	packetLoss := (float32(s.TotalUnsuccessfulProbes) / float32(totalPackets)) * 100

	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	fmt.Printf("%.2f%% packet loss\n", packetLoss)
	fmt.Printf("successful probes:   %d\n", s.TotalSuccessfulProbes)
	fmt.Printf("unsuccessful probes: %d\n", s.TotalUnsuccessfulProbes)

	fmt.Printf("last successful probe:   ")
	if s.LastSuccessfulProbe.IsZero() {
		fmt.Printf("Never succeeded\n")
	} else {
		fmt.Printf("%v\n", s.LastSuccessfulProbe.Format(time.DateTime))
	}

	fmt.Printf("last unsuccessful probe: ")
	if s.LastUnsuccessfulProbe.IsZero() {
		fmt.Printf("Never failed\n")
	} else {
		fmt.Printf("%v\n", s.LastUnsuccessfulProbe.Format(time.DateTime))
	}

	fmt.Printf("total uptime: %s\n", utils.DurationToString(s.TotalUptime))
	fmt.Printf("total downtime: %s\n", utils.DurationToString(s.TotalDowntime))

	if s.LongestUp.Duration != 0 {
		uptime := utils.DurationToString(s.LongestUp.Duration)

		fmt.Printf("longest consecutive uptime:   ")
		fmt.Printf("%v ", uptime)
		fmt.Printf("from %v ", s.LongestUp.Start.Format(time.DateTime))
		fmt.Printf("to %v\n", s.LongestUp.End.Format(time.DateTime))
	}

	if s.LongestDown.Duration != 0 {
		downtime := utils.DurationToString(s.LongestDown.Duration)

		fmt.Printf("longest consecutive downtime: %v ", downtime)
		fmt.Printf("from %v ", s.LongestDown.Start.Format(time.DateTime))
		fmt.Printf("to %v\n", s.LongestDown.End.Format(time.DateTime))
	}

	if !s.DestIsIP {
		timeNoun := "time"
		if s.RetriedHostnameLookups > 1 {
			timeNoun = "times"
		}

		fmt.Printf("retried to resolve hostname %d %s\n",
			s.RetriedHostnameLookups,
			timeNoun)

		if len(s.HostnameChanges) >= 2 {
			fmt.Printf("IP address changes:\n")
			for i := 0; i < len(s.HostnameChanges)-1; i++ {
				fmt.Printf("  from %s", s.HostnameChanges[i].Addr.String())
				fmt.Printf(" to %s", s.HostnameChanges[i+1].Addr.String())
				fmt.Printf(" at %v\n", s.HostnameChanges[i+1].When.Format(time.DateTime))
			}
		}
	}

	if s.RTTResults.HasResults {
		fmt.Printf("rtt min/avg/max: ")
		fmt.Printf("%.3f/%.3f/%.3f ms\n",
			s.RTTResults.Min,
			s.RTTResults.Average,
			s.RTTResults.Max)
	}

	fmt.Printf("--------------------------------------\n")
	fmt.Printf("TCPing started at: %v\n", s.StartTime.Format(time.DateTime))

	/* If the program was not terminated, no need to show the end time */
	if !s.EndTime.IsZero() {
		fmt.Printf("TCPing ended at:   %v\n", s.EndTime.Format(time.DateTime))
	}

	durationTime := time.Time{}.Add(s.TotalDowntime + s.TotalUptime)
	fmt.Printf("duration (HH:MM:SS): %v\n\n", durationTime.Format(time.TimeOnly))
}
