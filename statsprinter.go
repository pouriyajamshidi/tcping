package main

import (
	"encoding/json"
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

const (
	noReply = "No reply"
)

type statsPrinter interface {
	printStart()
	printLastSucUnsucProbes()
	printDurationStats()
	printStatistics()
	printReply(replyMsg replyMsg)
	printTotalDownTime(time.Time)
	printLongestUptime()
	printLongestDowntime()
	printRetryResolveStats()
	printRetryingToResolve()
}

type statsPlanePrinter struct {
	*stats
}

type statsJsonPrinter struct {
	*stats
}

/* Print host name and port to use on tcping */
func (p *statsPlanePrinter) printStart() {
	colorLightCyan("TCPinging %s on port %d\n", p.hostname, p.port)
}

/* Print the last successful and unsuccessful probes */
func (p *statsPlanePrinter) printLastSucUnsucProbes() {
	formattedLastSuccessfulProbe := p.lastSuccessfulProbe.Format(timeFormat)
	formattedLastUnsuccessfulProbe := p.lastUnsuccessfulProbe.Format(timeFormat)

	colorYellow("last successful probe:   ")
	if formattedLastSuccessfulProbe == nullTimeFormat {
		colorRed("Never succeeded\n")
	} else {
		colorGreen("%v\n", formattedLastSuccessfulProbe)
	}

	colorYellow("last unsuccessful probe: ")
	if formattedLastUnsuccessfulProbe == nullTimeFormat {
		colorGreen("Never failed\n")
	} else {
		colorRed("%v\n", formattedLastUnsuccessfulProbe)
	}
}

/* Print the start and end time of the program */
func (p *statsPlanePrinter) printDurationStats() {
	var duration time.Time
	var durationDiff time.Duration

	colorYellow("--------------------------------------\n")
	colorYellow("TCPing started at: %v\n", p.startTime.Format(timeFormat))

	/* If the program was not terminated, no need to show the end time */
	if p.endTime.Format(timeFormat) == nullTimeFormat {
		durationDiff = time.Since(p.startTime)
	} else {
		colorYellow("TCPing ended at:   %v\n", p.endTime.Format(timeFormat))
		durationDiff = p.endTime.Sub(p.startTime)
	}

	duration = time.Time{}.Add(durationDiff)
	colorYellow("duration (HH:MM:SS): %v\n\n", duration.Format(hourFormat))
}

/* Print statistics when program exits */
func (p *statsPlanePrinter) printStatistics() {
	rttResults := findMinAvgMaxRttTime(p.rtt)

	if !rttResults.hasResults {
		return
	}

	totalPackets := p.totalSuccessfulPkts + p.totalUnsuccessfulPkts
	totalUptime := calcTime(uint(p.totalUptime.Seconds()))
	totalDowntime := calcTime(uint(p.totalDowntime.Seconds()))
	packetLoss := (float32(p.totalUnsuccessfulPkts) / float32(totalPackets)) * 100

	/* general stats */
	colorYellow("\n--- %s TCPing statistics ---\n", p.hostname)
	colorYellow("%d probes transmitted, ", totalPackets)
	colorYellow("%d received, ", p.totalSuccessfulPkts)

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
	colorGreen("%d\n", p.totalSuccessfulPkts)

	/* unsuccessful packet stats */
	colorYellow("unsuccessful probes: ")
	colorRed("%d\n", p.totalUnsuccessfulPkts)

	p.printLastSucUnsucProbes()

	/* uptime and downtime stats */
	colorYellow("total uptime: ")
	colorGreen("  %s\n", totalUptime)
	colorYellow("total downtime: ")
	colorRed("%s\n", totalDowntime)

	/* calculate the last longest time */
	if !p.wasDown {
		calcLongestUptime(p.stats, p.lastSuccessfulProbe)
	} else {
		calcLongestDowntime(p.stats, p.lastUnsuccessfulProbe)
	}

	/* longest uptime stats */
	p.printLongestUptime()

	/* longest downtime stats */
	p.printLongestDowntime()

	/* resolve retry stats */
	if !p.isIP {
		p.printRetryResolveStats()
	}

	/*TODO: see if formatted string would suit better */
	/* latency stats.*/
	colorYellow("rtt ")
	colorGreen("min")
	colorYellow("/")
	colorCyan("avg")
	colorYellow("/")
	colorRed("max: ")
	colorGreen("%.3f", rttResults.min)
	colorYellow("/")
	colorCyan("%.3f", rttResults.average)
	colorYellow("/")
	colorRed("%.3f", rttResults.max)
	colorYellow(" ms\n")

	/* duration stats */
	p.printDurationStats()
}

/* Print TCP probe replies according to our policies */
func (p *statsPlanePrinter) printReply(replyMsg replyMsg) {
	if p.isIP {
		if replyMsg.msg == noReply {
			colorRed("%s from %s on port %d TCP_conn=%d\n",
				replyMsg.msg, p.ip, p.port, p.totalUnsuccessfulPkts)
		} else {
			colorLightGreen("%s from %s on port %d TCP_conn=%d time=%.3f ms\n",
				replyMsg.msg, p.ip, p.port, p.totalSuccessfulPkts, replyMsg.rtt)
		}
	} else {
		if replyMsg.msg == noReply {
			colorRed("%s from %s (%s) on port %d TCP_conn=%d\n",
				replyMsg.msg, p.hostname, p.ip, p.port, p.totalUnsuccessfulPkts)
		} else {
			colorLightGreen("%s from %s (%s) on port %d TCP_conn=%d time=%.3f ms\n",
				replyMsg.msg, p.hostname, p.ip, p.port, p.totalSuccessfulPkts, replyMsg.rtt)
		}
	}
}

/* Print the total downtime */
func (p *statsPlanePrinter) printTotalDownTime(now time.Time) {
	latestDowntimeDuration := time.Since(p.startOfDowntime).Seconds()
	calculatedDowntime := calcTime(uint(math.Ceil(latestDowntimeDuration)))
	colorYellow("No response received for %s\n", calculatedDowntime)
}

/* Print the longest uptime */
func (p *statsPlanePrinter) printLongestUptime() {
	if p.longestUptime.duration == 0 {
		return
	}

	uptime := calcTime(uint(math.Ceil(p.longestUptime.duration)))

	colorYellow("longest consecutive uptime:   ")
	colorGreen("%v ", uptime)
	colorYellow("from ")
	colorLightBlue("%v ", p.longestUptime.start.Format(timeFormat))
	colorYellow("to ")
	colorLightBlue("%v\n", p.longestUptime.end.Format(timeFormat))
}

/* Print the longest downtime */
func (p *statsPlanePrinter) printLongestDowntime() {
	if p.longestDowntime.duration == 0 {
		return
	}

	downtime := calcTime(uint(math.Ceil(p.longestDowntime.duration)))

	colorYellow("longest consecutive downtime: ")
	colorRed("%v ", downtime)
	colorYellow("from ")
	colorLightBlue("%v ", p.longestDowntime.start.Format(timeFormat))
	colorYellow("to ")
	colorLightBlue("%v\n", p.longestDowntime.end.Format(timeFormat))
}

/* Print the number of times that we tried resolving a hostname after a failure */
func (p *statsPlanePrinter) printRetryResolveStats() {
	colorYellow("retried to resolve hostname ")
	colorRed("%d ", p.retriedHostnameResolves)
	colorYellow("times\n")
}

/* Print the message retrying to resolve */
func (p *statsPlanePrinter) printRetryingToResolve() {
	colorLightYellow("retrying to resolve %s\n", p.hostname)
}

/*

JSON output section

*/

/* Print message  in JSON format */
func jsonPrintf(format string, a ...interface{}) {
	data := struct {
		Message string `json:"message"`
	}{
		Message: fmt.Sprintf(format, a...),
	}
	outputJson, _ := json.Marshal(&data)
	fmt.Println(string(outputJson))
}

/* Print host name and port to use on tcping in JSON format */
func (j *statsJsonPrinter) printStart() {
	jsonPrintf("TCPinging %s on port %d", j.hostname, j.port)
}

/* Print the last successful and unsuccessful probes in JSON format */
func (j *statsJsonPrinter) printLastSucUnsucProbes() {
	formattedLastSuccessfulProbe := j.lastSuccessfulProbe.Format(timeFormat)
	formattedLastUnsuccessfulProbe := j.lastUnsuccessfulProbe.Format(timeFormat)

	successMsg := "last successful probe:   "
	if formattedLastSuccessfulProbe == nullTimeFormat {
		jsonPrintf("%s%s", successMsg, "Never succeeded")
	} else {
		jsonPrintf("%s%s", successMsg, formattedLastSuccessfulProbe)
	}

	failureMsg := "last unsuccessful probe: "
	if formattedLastUnsuccessfulProbe == nullTimeFormat {
		jsonPrintf("%s%s", failureMsg, "Never failed")
	} else {
		jsonPrintf("%s%s", failureMsg, formattedLastUnsuccessfulProbe)
	}
}

/* Print the start and end time of the program in JSON format */
func (j *statsJsonPrinter) printDurationStats() {
	var duration time.Time
	var durationDiff time.Duration
	endMsg := "still running"

	startMSg := fmt.Sprintf("started at: %v ", j.startTime.Format(timeFormat))

	/* If the program was not terminated, no need to show the end time */
	if j.endTime.Format(timeFormat) == nullTimeFormat {
		durationDiff = time.Since(j.startTime)
	} else {
		endMsg = fmt.Sprintf("ended at: %v ", j.endTime.Format(timeFormat))
		durationDiff = j.endTime.Sub(j.startTime)
	}

	duration = time.Time{}.Add(durationDiff)
	durationFormatted := fmt.Sprintf("duration (HH:MM:SS): %v", duration.Format(hourFormat))

	jsonPrintf(startMSg + endMsg + durationFormatted)
}

/* Print statistics when program exits in JSON format */
func (j *statsJsonPrinter) printStatistics() {

	rttResults := findMinAvgMaxRttTime(j.rtt)

	if !rttResults.hasResults {
		return
	}

	totalPackets := j.totalSuccessfulPkts + j.totalUnsuccessfulPkts
	totalUptime := calcTime(uint(j.totalUptime.Seconds()))
	totalDowntime := calcTime(uint(j.totalDowntime.Seconds()))
	packetLoss := (float32(j.totalUnsuccessfulPkts) / float32(totalPackets)) * 100

	/* general stats */
	jsonPrintf("%s TCPing statistics: %d probes transmitted, %d received", j.hostname, totalPackets, j.totalSuccessfulPkts)

	/* packet loss stats */
	jsonPrintf("%.2f%% packet loss", packetLoss)

	/* successful packet stats */
	jsonPrintf("successful probes: %d", j.totalSuccessfulPkts)

	/* unsuccessful packet stats */
	jsonPrintf("unsuccessful probes: %d", j.totalUnsuccessfulPkts)

	j.printLastSucUnsucProbes()

	/* uptime and downtime stats */
	jsonPrintf("total uptime: %s", totalUptime)
	jsonPrintf("total downtime: %s", totalDowntime)

	/* calculate the last longest time */
	if !j.wasDown {
		calcLongestUptime(j.stats, j.lastSuccessfulProbe)
	} else {
		calcLongestDowntime(j.stats, j.lastUnsuccessfulProbe)
	}

	/* longest uptime stats */
	j.printLongestUptime()

	/* longest downtime stats */
	j.printLongestDowntime()

	/* resolve retry stats */
	if !j.isIP {
		j.printRetryResolveStats()
	}

	/* latency stats.*/
	jsonPrintf("rtt min/avg/max: %.3f/%.3f/%.3f", rttResults.min, rttResults.average, rttResults.max)

	/* duration stats */
	j.printDurationStats()
}

/* Print TCP probe replies according to our policies in JSON format */
func (j *statsJsonPrinter) printReply(replyMsg replyMsg) {
	if j.isIP {
		if replyMsg.msg == noReply {
			jsonPrintf("%s from %s on port %d TCP_conn=%d",
				replyMsg.msg, j.ip, j.port, j.totalUnsuccessfulPkts)
		} else {
			jsonPrintf("%s from %s on port %d TCP_conn=%d time=%.3f ms",
				replyMsg.msg, j.ip, j.port, j.totalSuccessfulPkts, replyMsg.rtt)
		}
	} else {
		if replyMsg.msg == noReply {
			jsonPrintf("%s from %s (%s) on port %d TCP_conn=%d",
				replyMsg.msg, j.hostname, j.ip, j.port, j.totalUnsuccessfulPkts)
		} else {
			jsonPrintf("%s from %s (%s) on port %d TCP_conn=%d time=%.3f ms",
				replyMsg.msg, j.hostname, j.ip, j.port, j.totalSuccessfulPkts, replyMsg.rtt)
		}
	}
}

/* Print the total downtime in JSON format */
func (j *statsJsonPrinter) printTotalDownTime(now time.Time) {
	latestDowntimeDuration := time.Since(j.startOfDowntime).Seconds()
	calculatedDowntime := calcTime(uint(math.Ceil(latestDowntimeDuration)))

	jsonPrintf("No response received for %s", calculatedDowntime)
}

/* Print the longest uptime in JSON format */
func (j *statsJsonPrinter) printLongestUptime() {
	if j.longestUptime.duration == 0 {
		return
	}

	uptime := calcTime(uint(math.Ceil(j.longestUptime.duration)))
	longestUptimeStart := j.longestUptime.start.Format(timeFormat)
	longestUptimeEnd := j.longestUptime.end.Format(timeFormat)

	jsonPrintf("longest consecutive uptime: %v from %v to %v", uptime, longestUptimeStart, longestUptimeEnd)
}

/* Print the longest downtime in JSON format */
func (j *statsJsonPrinter) printLongestDowntime() {
	if j.longestDowntime.duration == 0 {
		return
	}

	downtime := calcTime(uint(math.Ceil(j.longestDowntime.duration)))

	longestDowntimeStart := j.longestDowntime.start.Format(timeFormat)
	longestDowntimeEnd := j.longestDowntime.end.Format(timeFormat)

	jsonPrintf("longest consecutive downtime: %v from %v to %v", downtime, longestDowntimeStart, longestDowntimeEnd)
}

/* Print the number of times that we tried resolving a hostname after a failure in JSON format */
func (j *statsJsonPrinter) printRetryResolveStats() {
	jsonPrintf("retried to resolve hostname %d times", j.retriedHostnameResolves)
}

/* Print the message retrying to resolve in JSON format */
func (j *statsJsonPrinter) printRetryingToResolve() {
	jsonPrintf("retrying to resolve %s", j.hostname)
}
