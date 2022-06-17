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
)

type StatsPrinter interface {
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

func NewStatsPlanePrinter(stats *stats) StatsPrinter {
	return &statsPlanePrinter{stats: stats}
}

func (p *statsPlanePrinter) printStart() {
	color.LightCyan.Printf("TCPinging %s on port %s\n", p.hostname, p.port)
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
	colorGreen("%d", rttResults.fastest)
	colorYellow("/")
	colorCyan("%.2f", rttResults.average)
	colorYellow("/")
	colorRed("%d", rttResults.slowest)
	colorYellow(" ms\n")

	/* duration stats */
	p.printDurationStats()
}

/* Print TCP probe replies according to our policies */
func (p *statsPlanePrinter) printReply(replyMsg replyMsg) {
	if p.isIP {
		if replyMsg.msg == "No reply" {
			colorRed("%s from %s on port %s TCP_conn=%d\n",
				replyMsg.msg, p.ip, p.port, p.totalUnsuccessfulPkts)
		} else {
			colorLightGreen("%s from %s on port %s TCP_conn=%d time=%d ms\n",
				replyMsg.msg, p.ip, p.port, p.totalSuccessfulPkts, replyMsg.rtt)
		}
	} else {
		if replyMsg.msg == "No reply" {
			colorRed("%s from %s (%s) on port %s TCP_conn=%d\n",
				replyMsg.msg, p.hostname, p.ip, p.port, p.totalUnsuccessfulPkts)
		} else {
			colorLightGreen("%s from %s (%s) on port %s TCP_conn=%d time=%d ms\n",
				replyMsg.msg, p.hostname, p.ip, p.port, p.totalSuccessfulPkts, replyMsg.rtt)
		}
	}
}

func (p *statsPlanePrinter) printTotalDownTime(now time.Time) {
	latestDowntimeDuration := p.startOfDowntime.Sub(now).Seconds()
	calculatedDowntime := calcTime(uint(math.Ceil(latestDowntimeDuration)))
	color.Yellow.Printf("No response received for %s\n", calculatedDowntime)
}

/* Print the longest uptime */
func (p *statsPlanePrinter) printLongestUptime() {
	if p.longestUptime.duration == 0 {
		return
	}

	uptime := calcTime(uint(math.Ceil(p.longestUptime.duration)))

	colorYellow("longest uptime:   ")
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

	colorYellow("longest downtime: ")
	colorRed("%v ", downtime)
	colorYellow("from ")
	colorLightBlue("%v ", p.longestDowntime.start.Format(timeFormat))
	colorYellow("to ")
	colorLightBlue("%v\n", p.longestDowntime.end.Format(timeFormat))
}

/* Print the number of times that we tried resolving a hostname after a failure */
func (p *statsPlanePrinter) printRetryResolveStats() {
	colorYellow("Retried to resolve hostname ")
	colorRed("%d ", p.retriedHostnameResolves)
	colorYellow("times\n")
}

func (p *statsPlanePrinter) printRetryingToResolve() {
	colorLightYellow("Retrying to resolve %s\n", p.hostname)
}

func jsonPrintf(message string) {
	data := struct {
		Message string `json:"message"`
	}{
		Message: message,
	}
	outputJson, _ := json.Marshal(&data)
	fmt.Println(string(outputJson))
}

type statsJsonPrinter struct {
	*stats
}

func NewStatsJsonPrinter(stats *stats) StatsPrinter {
	return &statsJsonPrinter{stats: stats}
}

func (s *statsJsonPrinter) printStart() {
	jsonPrintf("printStart")
}

/* Print the last successful and unsuccessful probes */
func (s *statsJsonPrinter) printLastSucUnsucProbes() {
	jsonPrintf("printLastSucUnsucProbes")
}

/* Print the start and end time of the program */
func (s *statsJsonPrinter) printDurationStats() {
	jsonPrintf("printDurationStats")
}

/* Print statistics when program exits */
func (s *statsJsonPrinter) printStatistics() {
	jsonPrintf("printStatistics")
}

/* Print TCP probe replies according to our policies */
func (s *statsJsonPrinter) printReply(replyMsg replyMsg) {
	jsonPrintf("printReply")
}

func (s *statsJsonPrinter) printTotalDownTime(t time.Time) {
	jsonPrintf("printTotalDownTime")
}

/* Print the longest uptime */
func (s *statsJsonPrinter) printLongestUptime() {
	jsonPrintf("printLongestUptime")
}

/* Print the longest downtime */
func (s *statsJsonPrinter) printLongestDowntime() {
	jsonPrintf("printLongestDowntime")
}

/* Print the number of times that we tried resolving a hostname after a failure */
func (s *statsJsonPrinter) printRetryResolveStats() {
	jsonPrintf("printRetryResolveStats")
}

func (s *statsJsonPrinter) printRetryingToResolve() {
	jsonPrintf("printRetryingToResolve")
}
