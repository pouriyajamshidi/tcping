package printers

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"
)

type JSONPrinter struct {
	e *json.Encoder
}

func NewJSONPrinter(withIndent bool) *JSONPrinter {
	encoder := json.NewEncoder(os.Stdout)
	if withIndent {
		encoder.SetIndent("", "\t")
	}
	return &JSONPrinter{e: encoder}
}

// print is a little helper method for p.e.Encode.
// at also sets data.Timestamp to Now().
func (p *JSONPrinter) print(data JSONData) {
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
	LocalAddr            string           `json:"local_address,omitempty"`
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
func (p *JSONPrinter) printStart(hostname string, port uint16) {
	p.print(JSONData{
		Type:     startEvent,
		Message:  fmt.Sprintf("TCPinging %s on port %d", hostname, port),
		Hostname: hostname,
		Port:     port,
	})
}

// printReply prints TCP probe replies according to our policies in JSON format.
func (p *JSONPrinter) printProbeSuccess(sourceAddr string, userInput userInput, streak uint, rtt float32) {
	var (
		// for *bool fields
		f    = false
		t    = true
		data = JSONData{
			Type:                  probeEvent,
			Hostname:              userInput.hostname,
			Addr:                  userInput.ip.String(),
			Port:                  userInput.port,
			Rtt:                   rtt,
			DestIsIP:              &t,
			Success:               &t,
			TotalSuccessfulProbes: streak,
		}
	)
	if userInput.showSourceAddress {
		data.LocalAddr = sourceAddr
	}

	if userInput.hostname != "" {
		data.DestIsIP = &f
		if userInput.showSourceAddress {
			data.Message = fmt.Sprintf("Reply from %s (%s) on port %d using %s time=%.3f ms",
				userInput.hostname, userInput.ip.String(), userInput.port, sourceAddr, rtt)
		} else {
			data.Message = fmt.Sprintf("Reply from %s (%s) on port %d time=%.3f ms",
				userInput.hostname, userInput.ip.String(), userInput.port, rtt)
		}
	} else {
		if userInput.showSourceAddress {
			data.Message = fmt.Sprintf("Reply from %s on port %d using %s time=%.3f ms",
				userInput.ip.String(), userInput.port, sourceAddr, rtt)
		} else {
			data.Message = fmt.Sprintf("Reply from %s on port %d time=%.3f ms",
				userInput.ip.String(), userInput.port, rtt)
		}
	}

	p.print(data)
}

func (p *JSONPrinter) printProbeFail(userInput userInput, streak uint) {
	var (
		// for *bool fields
		f    = false
		t    = true
		data = JSONData{
			Type:                    probeEvent,
			Hostname:                userInput.hostname,
			Addr:                    userInput.ip.String(),
			Port:                    userInput.port,
			DestIsIP:                &t,
			Success:                 &f,
			TotalUnsuccessfulProbes: streak,
		}
	)

	if userInput.hostname != "" {
		data.DestIsIP = &f
		data.Message = fmt.Sprintf("No reply from %s (%s) on port %d",
			userInput.hostname, userInput.ip.String(), userInput.port)
	} else {
		data.Message = fmt.Sprintf("No reply from %s on port %d",
			userInput.ip.String(), userInput.port)
	}

	p.print(data)
}

// printStatistics prints all gathered stats when program exits.
func (p *JSONPrinter) printStatistics(t tcping) {
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
func (p *JSONPrinter) printTotalDownTime(downtime time.Duration) {
	p.print(JSONData{
		Type:          retrySuccessEvent,
		Message:       fmt.Sprintf("no response received for %s", durationToString(downtime)),
		TotalDowntime: downtime.Seconds(),
	})
}

// printRetryingToResolve print the message retrying to resolve,
// after n failed probes.
func (p *JSONPrinter) printRetryingToResolve(hostname string) {
	p.print(JSONData{
		Type:     retryEvent,
		Message:  fmt.Sprintf("retrying to resolve %s", hostname),
		Hostname: hostname,
	})
}

func (p *JSONPrinter) printInfo(format string, args ...any) {
	p.print(JSONData{
		Type:    infoEvent,
		Message: fmt.Sprintf(format, args...),
	})
}

func (p *JSONPrinter) printError(format string, args ...any) {
	p.print(JSONData{
		Type:    errorEvent,
		Message: fmt.Sprintf(format, args...),
	})
}

func (p *JSONPrinter) printVersion() {
	p.print(JSONData{
		Type:    versionEvent,
		Message: fmt.Sprintf("TCPING version %s\n", version),
	})
}
