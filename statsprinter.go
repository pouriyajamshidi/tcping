package main

import (
	"encoding/json"
	"fmt"
	"os"
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

type coloredPrinter struct{}

func (p *coloredPrinter) printStart(hostname string, port uint16) {
	colorLightCyan("TCPinging %s on port %d\n", hostname, port)
}

func (p *coloredPrinter) printStatistics(s stats) {
	totalPackets := s.totalSuccessfulProbes + s.totalUnsuccessfulProbes
	totalUptime := calcTime(s.totalUptime)
	totalDowntime := calcTime(s.totalDowntime)
	packetLoss := (float32(s.totalUnsuccessfulProbes) / float32(totalPackets)) * 100

	/* general stats */
	if !s.isIP {
		colorYellow("\n--- %s (%s) TCPing statistics ---\n", s.hostname, s.ip)
	} else {
		colorYellow("\n--- %s TCPing statistics ---\n", s.hostname)
	}
	colorYellow("%d probes transmitted on port %d | ", totalPackets, s.port)
	colorYellow("%d received, ", s.totalSuccessfulProbes)

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
	colorGreen("%d\n", s.totalSuccessfulProbes)

	/* unsuccessful packet stats */
	colorYellow("unsuccessful probes: ")
	colorRed("%d\n", s.totalUnsuccessfulProbes)

	colorYellow("last successful probe:   ")
	if s.lastSuccessfulProbe.IsZero() {
		colorRed("Never succeeded\n")
	} else {
		colorGreen("%v\n", s.lastSuccessfulProbe.Format(timeFormat))
	}

	colorYellow("last unsuccessful probe: ")
	if s.lastUnsuccessfulProbe.IsZero() {
		colorGreen("Never failed\n")
	} else {
		colorRed("%v\n", s.lastUnsuccessfulProbe.Format(timeFormat))
	}

	/* uptime and downtime stats */
	colorYellow("total uptime: ")
	colorGreen("  %s\n", totalUptime)
	colorYellow("total downtime: ")
	colorRed("%s\n", totalDowntime)

	/* calculate the last longest time */
	if !s.wasDown {
		calcLongestUptime(&s,
			time.Duration(s.ongoingSuccessfulProbes)*time.Second)
	} else {
		calcLongestDowntime(&s,
			time.Duration(s.ongoingUnsuccessfulProbes)*time.Second)
	}

	/* longest uptime stats */
	if s.longestUptime.duration != 0 {
		uptime := calcTime(s.longestUptime.duration)

		colorYellow("longest consecutive uptime:   ")
		colorGreen("%v ", uptime)
		colorYellow("from ")
		colorLightBlue("%v ", s.longestUptime.start.Format(timeFormat))
		colorYellow("to ")
		colorLightBlue("%v\n", s.longestUptime.end.Format(timeFormat))
	}

	/* longest downtime stats */
	if s.longestDowntime.duration != 0 {
		downtime := calcTime(s.longestDowntime.duration)

		colorYellow("longest consecutive downtime: ")
		colorRed("%v ", downtime)
		colorYellow("from ")
		colorLightBlue("%v ", s.longestDowntime.start.Format(timeFormat))
		colorYellow("to ")
		colorLightBlue("%v\n", s.longestDowntime.end.Format(timeFormat))
	}

	/* resolve retry stats */
	if !s.isIP {
		colorYellow("retried to resolve hostname ")
		colorRed("%d ", s.retriedHostnameResolves)
		colorYellow("times\n")
	}

	rtt := findMinAvgMaxRttTime(s.rtt)
	if rtt.hasResults {
		colorYellow("rtt ")
		colorGreen("min")
		colorYellow("/")
		colorCyan("avg")
		colorYellow("/")
		colorRed("max: ")
		colorGreen("%.3f", rtt.min)
		colorYellow("/")
		colorCyan("%.3f", rtt.average)
		colorYellow("/")
		colorRed("%.3f", rtt.max)
		colorYellow(" ms\n")
	}

	/* duration stats */
	var duration time.Time
	var durationDiff time.Duration

	colorYellow("--------------------------------------\n")
	colorYellow("TCPing started at: %v\n", s.startTime.Format(timeFormat))

	/* If the program was not terminated, no need to show the end time */
	if s.endTime.IsZero() {
		durationDiff = time.Since(s.startTime)
	} else {
		colorYellow("TCPing ended at:   %v\n", s.endTime.Format(timeFormat))
		durationDiff = s.endTime.Sub(s.startTime)
	}

	duration = time.Time{}.Add(durationDiff)
	colorYellow("duration (HH:MM:SS): %v\n\n", duration.Format(hourFormat))
}

func (p *coloredPrinter) printProbeSuccess(hostname, ip string, port uint16, streak uint, rtt float32) {
	if hostname == "" {
		colorLightGreen("Reply from %s on port %d TCP_conn=%d time=%.3f ms\n",
			ip, port, streak, rtt)
		return
	}

	colorLightGreen("Reply from %s (%s) on port %d TCP_conn=%d time=%.3f ms\n",
		hostname, ip, port, streak, rtt)
}

func (p *coloredPrinter) printProbeFail(hostname, ip string, port uint16, streak uint) {
	if hostname == "" {
		colorRed("No reply from %s on port %d TCP_conn=%d\n",
			ip, port, streak)
		return
	}
	colorRed("No reply from %s (%s) on port %d TCP_conn=%d\n",
		hostname, ip, port, streak)
}

func (p *coloredPrinter) printTotalDownTime(downtime time.Duration) {
	colorYellow("No response received for %s\n", calcTime(downtime))
}

func (p *coloredPrinter) printRetryingToResolve(hostname string) {
	colorLightYellow("retrying to resolve %s\n", hostname)
}

type jsonPrinter struct {
	*stats
}

type JSONEventType string

const (
	// probe is an event type that represents 1 probe / ping / request event.
	probe JSONEventType = "probe"
	// retry is an event type that's being sent,
	// when tcping retries to resolve a hostname.
	retry JSONEventType = "retry"
	// retrySuccess is a sub-type for of Retry event, when previous retry
	// was unsuccessful, but the next one became successful.
	retrySuccess JSONEventType = "retry-success"
	// start is an event type that's sent only once, at the very beginning
	// before doing any actual work.
	start JSONEventType = "start"
	// statsEvent is a final event that's being sent before tcping exits.
	statsEvent JSONEventType = "stats"
)

// jsonEncoder stores the encoder for json output.
// It could be used to tweak default options, like indentation
// or change output to another writer interface.
var jsonEncoder = json.NewEncoder(os.Stdout)

// printJson is a shortcut for Encode() on jsonEncoder.
var printJson = jsonEncoder.Encode

// JSONData contains all possible fields for JSON output.
// Because one event usually contains only a subset of fields,
// other fields will be omitted in the output.
type JSONData struct {
	// Type is a mandatory field that specifies type of a message/event.
	// Possible types are:
	//	- start
	// 	- probe
	// 	- retry
	// 	- stats
	//	- retry-success
	Type JSONEventType `json:"type"`
	// Message contains a human-readable message.
	Message string `json:"message"`
	// Timestamp contains data when a message was sent.
	Timestamp time.Time `json:"timestamp"`

	// Optional fields below

	Addr                 string `json:"addr,omitempty"`
	Hostname             string `json:"hostname,omitempty"`
	HostnameResolveTries uint   `json:"hostname_resolve_tries,omitempty"`
	IsIP                 *bool  `json:"is_ip,omitempty"`
	Port                 uint16 `json:"port,omitempty"`

	// Success is a special field from probe messages, containing information
	// whether request was successful or not.
	// It's a pointer on purpose, otherwise success=false will be omitted,
	// but we still need to omit it for non-probe messages.
	Success *bool `json:"success,omitempty"`

	// Latency in ms for a successful probe messages.
	Latency float32 `json:"latency,omitempty"`

	// LatencyMin is a latency stat for the stats event.
	//
	// It's a string on purpose, as we'd like to have exactly
	// 3 decimal places without doing extra math.
	LatencyMin string `json:"latency_min,omitempty"`
	// LatencyAvg is a latency stat for the stats event.
	//
	// It's a string on purpose, as we'd like to have exactly
	// 3 decimal places without doing extra math.
	LatencyAvg string `json:"latency_avg,omitempty"`
	// LatencyMax is a latency stat for the stats event.
	//
	// It's a string on purpose, as we'd like to have exactly
	// 3 decimal places without doing extra math.
	LatencyMax string `json:"latency_max,omitempty"`

	// TotalDuration is a total amount of seconds that program was running.
	//
	// It's a string on purpose, as we'd like to have exactly
	// 3 decimal places without doing extra math.
	TotalDuration string `json:"duration,omitempty"`
	// StartTimestamp is used as a start time of TotalDuration for stats messages.
	StartTimestamp *time.Time `json:"start_timestamp,omitempty"`
	// EndTimestamp is used as an end of TotalDuration for stats messages.
	EndTimestamp *time.Time `json:"end_timestamp,omitempty"`

	LastSuccessfulProbe   *time.Time `json:"last_successful_probe,omitempty"`
	LastUnsuccessfulProbe *time.Time `json:"last_unsuccessful_probe,omitempty"`

	// LongestUptime in seconds.
	//
	// It's a string on purpose, as we'd like to have exactly
	// 3 decimal places without doing extra math.
	LongestUptime      string     `json:"longest_uptime,omitempty"`
	LongestUptimeEnd   *time.Time `json:"longest_uptime_end,omitempty"`
	LongestUptimeStart *time.Time `json:"longest_uptime_start,omitempty"`

	// LongestDowntime in seconds.
	//
	// It's a string on purpose, as we'd like to have exactly
	// 3 decimal places without doing extra math.
	LongestDowntime      string     `json:"longest_downtime,omitempty"`
	LongestDowntimeEnd   *time.Time `json:"longest_downtime_end,omitempty"`
	LongestDowntimeStart *time.Time `json:"longest_downtime_start,omitempty"`

	TotalPacketLoss         float32 `json:"total_packet_loss,omitempty"`
	TotalPackets            uint    `json:"total_packets,omitempty"`
	TotalSuccessfulProbes   uint    `json:"total_successful_probes,omitempty"`
	TotalUnsuccessfulProbes uint    `json:"total_unsuccessful_probes,omitempty"`
	// TotalUptime in seconds.
	TotalUptime float64 `json:"total_uptime,omitempty"`
	// TotalDowntime in seconds.
	TotalDowntime float64 `json:"total_downtime,omitempty"`
}

// printStart prints the initial message before doing probes.
func (j *jsonPrinter) printStart() {
	_ = printJson(JSONData{
		Type:      start,
		Message:   fmt.Sprintf("TCPinging %s on port %d", j.hostname, j.port),
		Hostname:  j.hostname,
		Port:      j.port,
		Timestamp: time.Now(),
	})
}

// printReply prints TCP probe replies according to our policies in JSON format.
func (j *jsonPrinter) printReply(replyMsg replyMsg) {
	// for *bool fields
	f := false
	t := true

	data := JSONData{
		Type:      probe,
		Addr:      j.ip.String(),
		Port:      j.port,
		IsIP:      &t,
		Success:   &f,
		Timestamp: time.Now(),
	}

	ipStr := data.Addr
	if !j.isIP {
		data.Hostname = j.hostname
		data.IsIP = &f
		ipStr = fmt.Sprintf("%s (%s)", data.Hostname, ipStr)
	}

	data.Message = fmt.Sprintf("%s from %s on port %d", replyMsg.msg, ipStr, j.port)

	// TODO: no reply
	if replyMsg.msg != "No Reply" {
		data.Latency = replyMsg.rtt
		data.TotalSuccessfulProbes = j.totalSuccessfulProbes
		data.Success = &t
	} else {
		data.TotalUnsuccessfulProbes = j.totalUnsuccessfulProbes
	}

	_ = printJson(data)
}

// printStatistics prints all gathered stats when program exits.
func (j *jsonPrinter) printStatistics() {
	data := JSONData{
		Type:      statsEvent,
		Message:   fmt.Sprintf("stats for %s", j.hostname),
		Hostname:  j.hostname,
		Timestamp: time.Now(),

		StartTimestamp:          &j.startTime,
		TotalDowntime:           j.totalDowntime.Seconds(),
		TotalPackets:            j.totalSuccessfulProbes + j.totalUnsuccessfulProbes,
		TotalSuccessfulProbes:   j.totalSuccessfulProbes,
		TotalUnsuccessfulProbes: j.totalUnsuccessfulProbes,
		TotalUptime:             j.totalUptime.Seconds(),
	}

	data.TotalPacketLoss = (float32(data.TotalUnsuccessfulProbes) / float32(data.TotalPackets)) * 100

	if !j.lastSuccessfulProbe.IsZero() {
		data.LastSuccessfulProbe = &j.lastSuccessfulProbe
	}
	if !j.lastUnsuccessfulProbe.IsZero() {
		data.LastUnsuccessfulProbe = &j.lastUnsuccessfulProbe
	}

	/* calculate the last longest time */
	if !j.wasDown {
		calcLongestUptime(j.stats,
			time.Duration(j.ongoingSuccessfulProbes)*time.Second)
	} else {
		calcLongestDowntime(j.stats,
			time.Duration(j.ongoingUnsuccessfulProbes)*time.Second)
	}

	if j.longestUptime.duration != 0 {
		data.LongestUptime = fmt.Sprintf("%.3f", j.longestUptime.duration.Seconds())
		data.LongestUptimeStart = &j.longestUptime.start
		data.LongestUptimeEnd = &j.longestUptime.end
	}

	if j.longestDowntime.duration != 0 {
		data.LongestDowntime = fmt.Sprintf("%.3f", j.longestDowntime.duration.Seconds())
		data.LongestDowntimeStart = &j.longestDowntime.start
		data.LongestDowntimeEnd = &j.longestDowntime.end
	}

	if !j.isIP {
		data.HostnameResolveTries = j.retriedHostnameResolves
	}

	latencyStats := findMinAvgMaxRttTime(j.rtt)
	if latencyStats.hasResults {
		data.LatencyMin = fmt.Sprintf("%.3f", latencyStats.min)
		data.LatencyAvg = fmt.Sprintf("%.3f", latencyStats.average)
		data.LatencyMax = fmt.Sprintf("%.3f", latencyStats.max)
	}

	if j.endTime.IsZero() {
		t := time.Now()
		data.EndTimestamp = &t
	} else {
		data.EndTimestamp = &j.endTime
	}

	data.TotalDuration = fmt.Sprintf("%.3f",
		data.EndTimestamp.Sub(*data.StartTimestamp).Seconds())

	_ = printJson(data)
}

// printTotalDownTime prints the total downtime,
// if the next retry was successfull.
func (j *jsonPrinter) printTotalDownTime(now time.Time) {
	downtime := time.Since(j.startOfDowntime)

	_ = printJson(&JSONData{
		Type:          retrySuccess,
		Message:       fmt.Sprintf("no response received for %s", calcTime(downtime)),
		TotalDowntime: downtime.Seconds(),
		Timestamp:     time.Now(),
	})
}

// printRetryingToResolve print the message retrying to resolve,
// after n failed probes.
func (j *jsonPrinter) printRetryingToResolve() {
	_ = printJson(JSONData{
		Type:      retry,
		Message:   fmt.Sprintf("retrying to resolve %s", j.hostname),
		Hostname:  j.hostname,
		Timestamp: time.Now(),
	})
}
