package printers

import (
	"fmt"
	"math"
	"time"
)

type PlainPrinter struct {
	showTimestamp *bool
}

func NewPlainPrinter(showTimestamp *bool) *PlainPrinter {
	return &PlainPrinter{showTimestamp: showTimestamp}
}

func (p *PlainPrinter) printStart(hostname string, port uint16) {
	fmt.Printf("TCPinging %s on port %d\n", hostname, port)
}

func (p *PlainPrinter) printStatistics(t tcping) {
	totalPackets := t.totalSuccessfulProbes + t.totalUnsuccessfulProbes
	packetLoss := (float32(t.totalUnsuccessfulProbes) / float32(totalPackets)) * 100

	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	/* general stats */
	if !t.destIsIP {
		fmt.Printf("\n--- %s (%s) TCPing statistics ---\n", t.userInput.hostname, t.userInput.ip)
	} else {
		fmt.Printf("\n--- %s TCPing statistics ---\n", t.userInput.hostname)
	}
	fmt.Printf("%d probes transmitted on port %d | %d received, ", totalPackets, t.userInput.port, t.totalSuccessfulProbes)

	/* packet loss stats */
	fmt.Printf("%.2f%% packet loss\n", packetLoss)

	/* successful packet stats */
	fmt.Printf("successful probes:   %d\n", t.totalSuccessfulProbes)

	/* unsuccessful packet stats */
	fmt.Printf("unsuccessful probes: %d\n", t.totalUnsuccessfulProbes)

	fmt.Printf("last successful probe:   ")
	if t.lastSuccessfulProbe.IsZero() {
		fmt.Printf("Never succeeded\n")
	} else {
		fmt.Printf("%v\n", t.lastSuccessfulProbe.Format(timeFormat))
	}

	fmt.Printf("last unsuccessful probe: ")
	if t.lastUnsuccessfulProbe.IsZero() {
		fmt.Printf("Never failed\n")
	} else {
		fmt.Printf("%v\n", t.lastUnsuccessfulProbe.Format(timeFormat))
	}

	/* uptime and downtime stats */
	fmt.Printf("total uptime: %s\n", durationToString(t.totalUptime))
	fmt.Printf("total downtime: %s\n", durationToString(t.totalDowntime))

	/* longest uptime stats */
	if t.longestUptime.duration != 0 {
		uptime := durationToString(t.longestUptime.duration)

		fmt.Printf("longest consecutive uptime:   ")
		fmt.Printf("%v ", uptime)
		fmt.Printf("from %v ", t.longestUptime.start.Format(timeFormat))
		fmt.Printf("to %v\n", t.longestUptime.end.Format(timeFormat))
	}

	/* longest downtime stats */
	if t.longestDowntime.duration != 0 {
		downtime := durationToString(t.longestDowntime.duration)

		fmt.Printf("longest consecutive downtime: %v ", downtime)
		fmt.Printf("from %v ", t.longestDowntime.start.Format(timeFormat))
		fmt.Printf("to %v\n", t.longestDowntime.end.Format(timeFormat))
	}

	/* resolve retry stats */
	if !t.destIsIP {
		fmt.Printf("retried to resolve hostname %d times\n", t.retriedHostnameLookups)

		if len(t.hostnameChanges) >= 2 {
			fmt.Printf("IP address changes:\n")
			for i := 0; i < len(t.hostnameChanges)-1; i++ {
				fmt.Printf("  from %s", t.hostnameChanges[i].Addr.String())
				fmt.Printf(" to %s", t.hostnameChanges[i+1].Addr.String())
				fmt.Printf(" at %v\n", t.hostnameChanges[i+1].When.Format(timeFormat))
			}
		}
	}

	if t.rttResults.hasResults {
		fmt.Printf("rtt min/avg/max: ")
		fmt.Printf("%.3f/%.3f/%.3f ms\n", t.rttResults.min, t.rttResults.average, t.rttResults.max)
	}

	fmt.Printf("--------------------------------------\n")
	fmt.Printf("TCPing started at: %v\n", t.startTime.Format(timeFormat))

	/* If the program was not terminated, no need to show the end time */
	if !t.endTime.IsZero() {
		fmt.Printf("TCPing ended at:   %v\n", t.endTime.Format(timeFormat))
	}

	durationTime := time.Time{}.Add(t.totalDowntime + t.totalUptime)
	fmt.Printf("duration (HH:MM:SS): %v\n\n", durationTime.Format(hourFormat))
}

func (p *PlainPrinter) printProbeSuccess(sourceAddr string, userInput userInput, streak uint, rtt float32) {
	msg := "Reply from "

	if *p.showTimestamp {
		timestamp := time.Now().Format(timeFormat)
		msg = fmt.Sprintf("%v %v", timestamp, msg)
	}

	hostnameAndIP := userInput.ip.String()
	if userInput.hostname != hostnameAndIP {
		hostnameAndIP = fmt.Sprintf("%s (%s)", userInput.hostname, userInput.ip)
	}

	msg += fmt.Sprintf("%s on port %d", hostnameAndIP, userInput.port)

	if userInput.showSourceAddress {
		msg += fmt.Sprintf(" using %s", sourceAddr)
	}

	msg += fmt.Sprintf(" TCP_conn=%d time=%.3f ms\n", streak, rtt)

	fmt.Print(msg)
}

func (p *PlainPrinter) printProbeFail(userInput userInput, streak uint) {
	msg := "No reply from "

	if *p.showTimestamp {
		timestamp := time.Now().Format(timeFormat)
		msg = fmt.Sprintf("%v %v", timestamp, msg)
	}

	hostnameAndIP := userInput.ip.String()
	if userInput.hostname != hostnameAndIP {
		hostnameAndIP = fmt.Sprintf("%s (%s)", userInput.hostname, userInput.ip)
	}

	msg += fmt.Sprintf("%s on port %d TCP_conn=%d\n", hostnameAndIP, userInput.port, streak)

	fmt.Print(msg)
}

func (p *PlainPrinter) printTotalDownTime(downtime time.Duration) {
	fmt.Printf("No response received for %s\n", durationToString(downtime))
}

func (p *PlainPrinter) printRetryingToResolve(hostname string) {
	fmt.Printf("retrying to resolve %s\n", hostname)
}

func (p *PlainPrinter) printInfo(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

func (p *PlainPrinter) printError(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

func (p *PlainPrinter) printVersion() {
	fmt.Printf("TCPING version %s\n", version)
}
