package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/gookit/color"
)

type Color string

const (
	timeFormat        = "2006-01-02 15:04:05"
	hourFormat        = "15:04:05"
	Yellow      Color = "yellow"
	Cyan        Color = "cyan"
	Green       Color = "green"
	Red         Color = "red"
	LightYellow Color = "light_yellow"
	LightGreen  Color = "light_green"
	LightCyan   Color = "light_cyan"
	LightBlue   Color = "light_blue"
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

type planePrinter struct{}

func printReply(color Color, message string) {
	t := time.Now().Format(timeFormat)
	messageWithTimestamp := fmt.Sprintf("%s %s", t, message)

	switch color {
	case Yellow:
		colorYellow(messageWithTimestamp)
	case Cyan:
		colorCyan(messageWithTimestamp)
	case Green:
		colorGreen(messageWithTimestamp)
	case Red:
		colorRed(messageWithTimestamp)
	case LightYellow:
		colorLightYellow(messageWithTimestamp)
	case LightGreen:
		colorLightGreen(messageWithTimestamp)
	case LightCyan:
		colorLightCyan(messageWithTimestamp)
	case LightBlue:
		colorLightBlue(messageWithTimestamp)
	default:
		colorYellow(messageWithTimestamp)
	}
}

func (p *planePrinter) printStart(hostname string, port uint16) {
	printReply(LightCyan, fmt.Sprintf("TCPinging %s on port %d\n", hostname, port))
}

func (p *planePrinter) printStatistics(t tcping) {
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

func (p *planePrinter) printProbeSuccess(hostname, ip string, port uint16, streak uint, rtt float32) {
	if hostname == "" {
		printReply(LightGreen, fmt.Sprintf("Reply from %s on port %d TCP_conn=%d time=%.3f ms\n",
			ip, port, streak, rtt))
		return
	}

	printReply(Green, fmt.Sprintf("Reply from %s (%s) on port %d TCP_conn=%d time=%.3f ms\n",
		hostname, ip, port, streak, rtt))
}

func (p *planePrinter) printProbeFail(hostname, ip string, port uint16, streak uint) {
	if hostname == "" {
		printReply(Red, fmt.Sprintf("No reply from %s on port %d TCP_conn=%d\n",
			ip, port, streak))
		return
	}
	printReply(Red, fmt.Sprintf("No reply from %s (%s) on port %d TCP_conn=%d\n",
		hostname, ip, port, streak))
}

func (p *planePrinter) printTotalDownTime(downtime time.Duration) {
	printReply(Yellow, fmt.Sprintf("No response received for %s\n", durationToString(downtime)))
}

func (p *planePrinter) printRetryingToResolve(hostname string) {
	printReply(LightYellow, fmt.Sprintf("retrying to resolve %s\n", hostname))
}

func (p *planePrinter) printInfo(format string, args ...any) {
	printReply(LightCyan, fmt.Sprintf(format+"\n", args...))
}

func (p *planePrinter) printError(format string, args ...any) {
	printReply(Red, fmt.Sprintf(format+"\n", args...))
}

func (p *planePrinter) printVersion() {
	printReply(LightYellow, fmt.Sprintf("TCPING version %s\n", version))
}

type jsonPrinter struct {
	e *json.Encoder
}

func newJSONPrinter(withIndent bool) *jsonPrinter {
	encoder := json.NewEncoder(os.Stdout)
	if withIndent {
		encoder.SetIndent("", "\t")
	}
	return &jsonPrinter{e: encoder}
}

// print is a little helper method for p.e.Encode.
// at also sets data.Timestamp to Now().
func (p *jsonPrinter) print(data JSONData) {
	data.Timestamp = time.Now()
	p.e.Encode(data)
}

// JSONEventType is a special type, each for each method
// in the printer interface so that automatic tools
// can understand what kind of an event they've received.
type JSONEventType string

const (
	// startEvent is an event type for [printStart] method.
	startEvent JSONEventType = "start"
	// probeEvent is a general event type for both
	// [printProbeSuccess] and [printProbeFail].
	probeEvent JSONEventType = "probe"
	// retryEvent is an event type for [printRetryingToResolve] method.
	retryEvent JSONEventType = "retry"
	// retrySuccessEvent is an event type for [printTotalDowntime] method.
	retrySuccessEvent JSONEventType = "retry-success"
	// statisticsEvent is a event type for [printStatistics] method.
	statisticsEvent JSONEventType = "statistics"
	// infoEvent is a event type for [printInfo] method.
	infoEvent JSONEventType = "info"
	// versionEvent is a event type for [printVersion] method.
	versionEvent JSONEventType = "version"
	// errorEvent is a event type for [printError] method.
	errorEvent JSONEventType = "error"
)

// JSONData contains all possible fields for JSON output.
// Because one event usually contains only a subset of fields,
// other fields will be omitted in the output.
type JSONData struct {
	// Type is a mandatory field that specifies type of a message/event.
	Type JSONEventType `json:"type"`
	// Message contains a human-readable message.
	Message string `json:"message"`
	// Timestamp contains data when a message was sent.
	Timestamp time.Time `json:"timestamp"`

	// Optional fields below

	Addr                 string           `json:"addr,omitempty"`
	Hostname             string           `json:"hostname,omitempty"`
	HostnameResolveTries uint             `json:"hostname_resolve_tries,omitempty"`
	HostnameChanges      []hostnameChange `json:"hostname_changes,omitempty"`
	DestIsIP             *bool            `json:"dst_is_ip,omitempty"`
	Port                 uint16           `json:"port,omitempty"`
	Rtt                  float32          `json:"time,omitempty"`

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
	TotalDuration string `json:"total_duration,omitempty"`
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

	// TotalPacketLoss in seconds.
	//
	// It's a string on purpose, as we'd like to have exactly
	// 3 decimal places without doing extra math.
	TotalPacketLoss         string `json:"total_packet_loss,omitempty"`
	TotalPackets            uint   `json:"total_packets,omitempty"`
	TotalSuccessfulProbes   uint   `json:"total_successful_probes,omitempty"`
	TotalUnsuccessfulProbes uint   `json:"total_unsuccessful_probes,omitempty"`
	// TotalUptime in seconds.
	TotalUptime float64 `json:"total_uptime,omitempty"`
	// TotalDowntime in seconds.
	TotalDowntime float64 `json:"total_downtime,omitempty"`
}

// printStart prints the initial message before doing probes.
func (p *jsonPrinter) printStart(hostname string, port uint16) {
	p.print(JSONData{
		Type:     startEvent,
		Message:  fmt.Sprintf("TCPinging %s on port %d", hostname, port),
		Hostname: hostname,
		Port:     port,
	})
}

// printReply prints TCP probe replies according to our policies in JSON format.
func (p *jsonPrinter) printProbeSuccess(
	hostname, ip string,
	port uint16,
	streak uint,
	rtt float32,
) {
	var (
		// for *bool fields
		f    = false
		t    = true
		data = JSONData{
			Type:                  probeEvent,
			Hostname:              hostname,
			Addr:                  ip,
			Port:                  port,
			Rtt:                   rtt,
			DestIsIP:              &t,
			Success:               &t,
			TotalSuccessfulProbes: streak,
		}
	)

	if hostname != "" {
		data.DestIsIP = &f
		data.Message = fmt.Sprintf("Reply from %s (%s) on port %d time=%.3f",
			hostname, ip, port, rtt)
	} else {
		data.Message = fmt.Sprintf("Reply from %s on port %d time=%.3f",
			ip, port, rtt)
	}

	p.print(data)
}

func (p *jsonPrinter) printProbeFail(hostname, ip string, port uint16, streak uint) {
	var (
		// for *bool fields
		f    = false
		t    = true
		data = JSONData{
			Type:                    probeEvent,
			Hostname:                hostname,
			Addr:                    ip,
			Port:                    port,
			DestIsIP:                &t,
			Success:                 &f,
			TotalUnsuccessfulProbes: streak,
		}
	)

	if hostname != "" {
		data.DestIsIP = &f
		data.Message = fmt.Sprintf("No reply from %s (%s) on port %d",
			hostname, ip, port)
	} else {
		data.Message = fmt.Sprintf("No reply from %s on port %d",
			ip, port)
	}

	p.print(data)
}

// printStatistics prints all gathered stats when program exits.
func (p *jsonPrinter) printStatistics(t tcping) {
	data := JSONData{
		Type:     statisticsEvent,
		Message:  fmt.Sprintf("stats for %s", t.userInput.hostname),
		Addr:     t.userInput.ip.String(),
		Hostname: t.userInput.hostname,

		StartTimestamp:          &t.startTime,
		TotalDowntime:           t.totalDowntime.Seconds(),
		TotalPackets:            t.totalSuccessfulProbes + t.totalUnsuccessfulProbes,
		TotalSuccessfulProbes:   t.totalSuccessfulProbes,
		TotalUnsuccessfulProbes: t.totalUnsuccessfulProbes,
		TotalUptime:             t.totalUptime.Seconds(),
	}

	if len(t.hostnameChanges) > 1 {
		data.HostnameChanges = t.hostnameChanges
	}

	loss := (float32(data.TotalUnsuccessfulProbes) / float32(data.TotalPackets)) * 100
	if math.IsNaN(float64(loss)) {
		loss = 0
	}
	data.TotalPacketLoss = fmt.Sprintf("%.2f", loss)

	if !t.lastSuccessfulProbe.IsZero() {
		data.LastSuccessfulProbe = &t.lastSuccessfulProbe
	}
	if !t.lastUnsuccessfulProbe.IsZero() {
		data.LastUnsuccessfulProbe = &t.lastUnsuccessfulProbe
	}

	if t.longestUptime.duration != 0 {
		data.LongestUptime = fmt.Sprintf("%.0f", t.longestUptime.duration.Seconds())
		data.LongestUptimeStart = &t.longestUptime.start
		data.LongestUptimeEnd = &t.longestUptime.end
	}

	if t.longestDowntime.duration != 0 {
		data.LongestDowntime = fmt.Sprintf("%.0f", t.longestDowntime.duration.Seconds())
		data.LongestDowntimeStart = &t.longestDowntime.start
		data.LongestDowntimeEnd = &t.longestDowntime.end
	}

	if !t.destIsIP {
		data.HostnameResolveTries = t.retriedHostnameLookups
	}

	if t.rttResults.hasResults {
		data.LatencyMin = fmt.Sprintf("%.3f", t.rttResults.min)
		data.LatencyAvg = fmt.Sprintf("%.3f", t.rttResults.average)
		data.LatencyMax = fmt.Sprintf("%.3f", t.rttResults.max)
	}

	if !t.endTime.IsZero() {
		data.EndTimestamp = &t.endTime
	}

	totalDuration := t.totalDowntime + t.totalUptime
	data.TotalDuration = fmt.Sprintf("%.0f", totalDuration.Seconds())

	p.print(data)
}

// printTotalDownTime prints the total downtime,
// if the next retry was successful.
func (p *jsonPrinter) printTotalDownTime(downtime time.Duration) {
	p.print(JSONData{
		Type:          retrySuccessEvent,
		Message:       fmt.Sprintf("no response received for %s", durationToString(downtime)),
		TotalDowntime: downtime.Seconds(),
	})
}

// printRetryingToResolve print the message retrying to resolve,
// after n failed probes.
func (p *jsonPrinter) printRetryingToResolve(hostname string) {
	p.print(JSONData{
		Type:     retryEvent,
		Message:  fmt.Sprintf("retrying to resolve %s", hostname),
		Hostname: hostname,
	})
}

func (p *jsonPrinter) printInfo(format string, args ...any) {
	p.print(JSONData{
		Type:    infoEvent,
		Message: fmt.Sprintf(format, args...),
	})
}

func (p *jsonPrinter) printError(format string, args ...any) {
	p.print(JSONData{
		Type:    errorEvent,
		Message: fmt.Sprintf(format, args...),
	})
}

func (p *jsonPrinter) printVersion() {
	p.print(JSONData{
		Type:    versionEvent,
		Message: fmt.Sprintf("TCPING version %s\n", version),
	})
}

// durationToString creates a human-readable string for a given duration
func durationToString(duration time.Duration) string {
	hours := math.Floor(duration.Hours())
	if hours > 0 {
		duration -= time.Duration(hours * float64(time.Hour))
	}

	minutes := math.Floor(duration.Minutes())
	if minutes > 0 {
		duration -= time.Duration(minutes * float64(time.Minute))
	}

	seconds := duration.Seconds()

	switch {
	// Hours
	case hours >= 2:
		return fmt.Sprintf("%.0f hours %.0f minutes %.0f seconds", hours, minutes, seconds)
	case hours == 1 && minutes == 0 && seconds == 0:
		return fmt.Sprintf("%.0f hour", hours)
	case hours == 1:
		return fmt.Sprintf("%.0f hour %.0f minutes %.0f seconds", hours, minutes, seconds)

	// Minutes
	case minutes >= 2:
		return fmt.Sprintf("%.0f minutes %.0f seconds", minutes, seconds)
	case minutes == 1 && seconds == 0:
		return fmt.Sprintf("%.0f minute", minutes)
	case minutes == 1:
		return fmt.Sprintf("%.0f minute %.0f seconds", minutes, seconds)

	// Seconds
	case seconds == 1:
		return fmt.Sprintf("%.0f second", seconds)
	case seconds < 1:
		return fmt.Sprintf("%.1f seconds", seconds)
	default:
		return fmt.Sprintf("%.0f seconds", seconds)

	}
}
