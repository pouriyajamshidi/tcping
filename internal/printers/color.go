package printers

import (
	"fmt"
	"math"
	"time"

	"github.com/gookit/color"
)

var (
	colorYellow      = color.Yellow.Printf
	colorGreen       = color.Green.Printf
	colorRed         = color.Red.Printf
	colorCyan        = color.Cyan.Printf
	colorLightYellow = color.LightYellow.Printf
	colorLightBlue   = color.FgLightBlue.Printf
	colorLightGreen  = color.LightGreen.Printf
	colorLightCyan   = color.LightCyan.Printf
)

type ColorPrinter struct {
	showTimestamp *bool
}

func NewColorPrinter(showTimestamp *bool) *ColorPrinter {
	return &ColorPrinter{showTimestamp: showTimestamp}
}

func (p *ColorPrinter) printStart(hostname string, port uint16) {
	colorLightCyan("TCPinging %s on port %d\n", hostname, port)
}

func (p *ColorPrinter) printStatistics(t tcping) {
	totalPackets := t.totalSuccessfulProbes + t.totalUnsuccessfulProbes
	packetLoss := (float32(t.totalUnsuccessfulProbes) / float32(totalPackets)) * 100

	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	/* general stats */
	if !t.destIsIP {
		colorYellow("\n--- %s (%s) TCPing statistics ---\n", t.userInput.hostname, t.userInput.ip)
	} else {
		colorYellow("\n--- %s TCPing statistics ---\n", t.userInput.hostname)
	}
	colorYellow("%d probes transmitted on port %d | ", totalPackets, t.userInput.port)
	colorYellow("%d received, ", t.totalSuccessfulProbes)

	/* packet loss stats */
	if packetLoss == 0 {
		colorGreen("%.2f%%", packetLoss)
	} else if packetLoss > 0 && packetLoss <= 30 {
		colorLightYellow("%.2f%%", packetLoss)
	} else {
		colorRed("%.2f%%", packetLoss)
	}

	colorYellow(" packet loss\n")

	/* successful packet stats */
	colorYellow("successful probes:   ")
	colorGreen("%d\n", t.totalSuccessfulProbes)

	/* unsuccessful packet stats */
	colorYellow("unsuccessful probes: ")
	colorRed("%d\n", t.totalUnsuccessfulProbes)

	colorYellow("last successful probe:   ")
	if t.lastSuccessfulProbe.IsZero() {
		colorRed("Never succeeded\n")
	} else {
		colorGreen("%v\n", t.lastSuccessfulProbe.Format(timeFormat))
	}

	colorYellow("last unsuccessful probe: ")
	if t.lastUnsuccessfulProbe.IsZero() {
		colorGreen("Never failed\n")
	} else {
		colorRed("%v\n", t.lastUnsuccessfulProbe.Format(timeFormat))
	}

	/* uptime and downtime stats */
	colorYellow("total uptime: ")
	colorGreen("  %s\n", durationToString(t.totalUptime))
	colorYellow("total downtime: ")
	colorRed("%s\n", durationToString(t.totalDowntime))

	/* longest uptime stats */
	if t.longestUptime.duration != 0 {
		uptime := durationToString(t.longestUptime.duration)

		colorYellow("longest consecutive uptime:   ")
		colorGreen("%v ", uptime)
		colorYellow("from ")
		colorLightBlue("%v ", t.longestUptime.start.Format(timeFormat))
		colorYellow("to ")
		colorLightBlue("%v\n", t.longestUptime.end.Format(timeFormat))
	}

	/* longest downtime stats */
	if t.longestDowntime.duration != 0 {
		downtime := durationToString(t.longestDowntime.duration)

		colorYellow("longest consecutive downtime: ")
		colorRed("%v ", downtime)
		colorYellow("from ")
		colorLightBlue("%v ", t.longestDowntime.start.Format(timeFormat))
		colorYellow("to ")
		colorLightBlue("%v\n", t.longestDowntime.end.Format(timeFormat))
	}

	/* resolve retry stats */
	if !t.destIsIP {
		colorYellow("retried to resolve hostname ")
		colorRed("%d ", t.retriedHostnameLookups)
		colorYellow("times\n")

		if len(t.hostnameChanges) >= 2 {
			colorYellow("IP address changes:\n")
			for i := 0; i < len(t.hostnameChanges)-1; i++ {
				colorYellow("  from ")
				colorRed(t.hostnameChanges[i].Addr.String())
				colorYellow(" to ")
				colorGreen(t.hostnameChanges[i+1].Addr.String())
				colorYellow(" at ")
				colorLightBlue("%v\n", t.hostnameChanges[i+1].When.Format(timeFormat))
			}
		}
	}

	if t.rttResults.hasResults {
		colorYellow("rtt ")
		colorGreen("min")
		colorYellow("/")
		colorCyan("avg")
		colorYellow("/")
		colorRed("max: ")
		colorGreen("%.3f", t.rttResults.min)
		colorYellow("/")
		colorCyan("%.3f", t.rttResults.average)
		colorYellow("/")
		colorRed("%.3f", t.rttResults.max)
		colorYellow(" ms\n")
	}

	colorYellow("--------------------------------------\n")
	colorYellow("TCPing started at: %v\n", t.startTime.Format(timeFormat))

	/* If the program was not terminated, no need to show the end time */
	if !t.endTime.IsZero() {
		colorYellow("TCPing ended at:   %v\n", t.endTime.Format(timeFormat))
	}

	durationTime := time.Time{}.Add(t.totalDowntime + t.totalUptime)
	colorYellow("duration (HH:MM:SS): %v\n\n", durationTime.Format(hourFormat))
}

func (p *ColorPrinter) printProbeSuccess(sourceAddr string, userInput userInput, streak uint, rtt float32) {
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

	colorLightGreen(msg)
}

func (p *ColorPrinter) printProbeFail(userInput userInput, streak uint) {
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

	colorRed(msg)
}

func (p *ColorPrinter) printTotalDownTime(downtime time.Duration) {
	colorYellow("No response received for %s\n", durationToString(downtime))
}

func (p *ColorPrinter) printRetryingToResolve(hostname string) {
	colorLightYellow("retrying to resolve %s\n", hostname)
}

func (p *ColorPrinter) printInfo(format string, args ...any) {
	colorLightBlue(format+"\n", args...)
}

func (p *ColorPrinter) printError(format string, args ...any) {
	colorRed(format+"\n", args...)
}

func (p *ColorPrinter) printVersion() {
	colorGreen("TCPING version %s\n", version)
}
