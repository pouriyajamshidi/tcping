package printers

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/internal/utils"
	"github.com/pouriyajamshidi/tcping/v2/types"
)

// JSONPrinter is a struct that holds a JSON encoder to print structured JSON output.
type JSONPrinter struct {
	e *json.Encoder
}

// NewJSONPrinter creates a new JSONPrinter instance.
// If withIndent is true, the JSON output will be formatted with indentation.
func NewJSONPrinter(withIndent bool) *JSONPrinter {
	encoder := json.NewEncoder(os.Stdout)
	if withIndent {
		encoder.SetIndent("", "\t")
	}
	return &JSONPrinter{e: encoder}
}

// Print is a little helper method for p.e.Encode.
// at also sets data.Timestamp to Now().
func (p *JSONPrinter) Print(data JSONData) {
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

	Addr                 string                 `json:"addr,omitempty"`
	LocalAddr            string                 `json:"local_address,omitempty"`
	Hostname             string                 `json:"hostname,omitempty"`
	HostnameResolveTries uint                   `json:"hostname_resolve_tries,omitempty"`
	HostnameChanges      []types.HostnameChange `json:"hostname_changes,omitempty"`
	DestIsIP             *bool                  `json:"dst_is_ip,omitempty"`
	Port                 uint16                 `json:"port,omitempty"`
	Rtt                  float32                `json:"time,omitempty"`

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

	// LongestUptime In seconds.
	//
	// It's a string on purpose, as we'd like to have exactly
	// 3 decimal places without doing extra math.
	LongestUptime      string     `json:"longest_uptime,omitempty"`
	LongestUptimeENd   *time.Time `json:"longest_uptime_end,omitempty"`
	LongestUptimeSTart *time.Time `json:"longest_uptime_start,omitempty"`

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

// PrintStart prints the initial message before doing probes.
func (p *JSONPrinter) PrintStart(hostname string, port uint16) {
	p.Print(JSONData{
		Type:     startEvent,
		Message:  fmt.Sprintf("TCPinging %s on port %d", hostname, port),
		Hostname: hostname,
		Port:     port,
	})
}

// PrintProbeSuccess prints TCP probe replies according to our policies in JSON format.
func (p *JSONPrinter) PrintProbeSuccess(startTime time.Time, sourceAddr string, opts types.Options, streak uint, rtt float32) {
	var (
		// for *bool fields
		f    = false
		t    = true
		data = JSONData{
			Type:                  probeEvent,
			Hostname:              opts.Hostname,
			Addr:                  opts.IP.String(),
			Port:                  opts.Port,
			Rtt:                   rtt,
			DestIsIP:              &t,
			Success:               &t,
			TotalSuccessfulProbes: streak,
		}
	)
	if opts.ShowSourceAddress {
		data.LocalAddr = sourceAddr
	}

	if opts.Hostname != "" {
		data.DestIsIP = &f
		if opts.ShowSourceAddress {
			data.Message = fmt.Sprintf("Reply from %s (%s) on port %d using %s time=%.3f ms",
				opts.Hostname, opts.IP.String(), opts.Port, sourceAddr, rtt)
		} else {
			data.Message = fmt.Sprintf("Reply from %s (%s) on port %d time=%.3f ms",
				opts.Hostname, opts.IP.String(), opts.Port, rtt)
		}
	} else {
		if opts.ShowSourceAddress {
			data.Message = fmt.Sprintf("Reply from %s on port %d using %s time=%.3f ms",
				opts.IP.String(), opts.Port, sourceAddr, rtt)
		} else {
			data.Message = fmt.Sprintf("Reply from %s on port %d time=%.3f ms",
				opts.IP.String(), opts.Port, rtt)
		}
	}

	p.Print(data)
}

// PrintProbeFail prints a JSON message when a TCP probe fails.
func (p *JSONPrinter) PrintProbeFail(startTime time.Time, opts types.Options, streak uint) {
	var (
		// for *bool fields
		f    = false
		t    = true
		data = JSONData{
			Type:                    probeEvent,
			Hostname:                opts.Hostname,
			Addr:                    opts.IP.String(),
			Port:                    opts.Port,
			DestIsIP:                &t,
			Success:                 &f,
			TotalUnsuccessfulProbes: streak,
		}
	)

	if opts.Hostname != "" {
		data.DestIsIP = &f
		data.Message = fmt.Sprintf("No reply from %s (%s) on port %d",
			opts.Hostname, opts.IP.String(), opts.Port)
	} else {
		data.Message = fmt.Sprintf("No reply from %s on port %d",
			opts.IP.String(), opts.Port)
	}

	p.Print(data)
}

// PrintStatistics prints all gathered stats when program exits.
func (p *JSONPrinter) PrintStatistics(t types.Tcping) {
	data := JSONData{
		Type:     statisticsEvent,
		Message:  fmt.Sprintf("stats for %s", t.Options.Hostname),
		Addr:     t.Options.IP.String(),
		Hostname: t.Options.Hostname,

		StartTimestamp:          &t.StartTime,
		TotalDowntime:           t.TotalDowntime.Seconds(),
		TotalPackets:            t.TotalSuccessfulProbes + t.TotalUnsuccessfulProbes,
		TotalSuccessfulProbes:   t.TotalSuccessfulProbes,
		TotalUnsuccessfulProbes: t.TotalUnsuccessfulProbes,
		TotalUptime:             t.TotalUptime.Seconds(),
	}

	if len(t.HostnameChanges) > 1 {
		data.HostnameChanges = t.HostnameChanges
	}

	loss := (float32(data.TotalUnsuccessfulProbes) / float32(data.TotalPackets)) * 100
	if math.IsNaN(float64(loss)) {
		loss = 0
	}
	data.TotalPacketLoss = fmt.Sprintf("%.2f", loss)

	if !t.LastSuccessfulProbe.IsZero() {
		data.LastSuccessfulProbe = &t.LastSuccessfulProbe
	}
	if !t.LastUnsuccessfulProbe.IsZero() {
		data.LastUnsuccessfulProbe = &t.LastUnsuccessfulProbe
	}

	if t.LongestUptime.Duration != 0 {
		data.LongestUptime = fmt.Sprintf("%.0f", t.LongestUptime.Duration.Seconds())
		data.LongestUptimeSTart = &t.LongestUptime.Start
		data.LongestUptimeENd = &t.LongestUptime.End
	}

	if t.LongestDowntime.Duration != 0 {
		data.LongestDowntime = fmt.Sprintf("%.0f", t.LongestDowntime.Duration.Seconds())
		data.LongestDowntimeStart = &t.LongestDowntime.Start
		data.LongestDowntimeEnd = &t.LongestDowntime.End
	}

	if !t.DestIsIP {
		data.HostnameResolveTries = t.RetriedHostnameLookups
	}

	if t.RttResults.HasResults {
		data.LatencyMin = fmt.Sprintf("%.3f", t.RttResults.Min)
		data.LatencyAvg = fmt.Sprintf("%.3f", t.RttResults.Average)
		data.LatencyMax = fmt.Sprintf("%.3f", t.RttResults.Max)
	}

	if !t.EndTime.IsZero() {
		data.EndTimestamp = &t.EndTime
	}

	totalDuration := t.TotalDowntime + t.TotalUptime
	data.TotalDuration = fmt.Sprintf("%.0f", totalDuration.Seconds())

	p.Print(data)
}

// PrintTotalDownTime prints the total downtime,
// if the next retry was successful.
func (p *JSONPrinter) PrintTotalDownTime(downtime time.Duration) {
	p.Print(JSONData{
		Type:          retrySuccessEvent,
		Message:       fmt.Sprintf("no response received for %s", utils.DurationToString(downtime)),
		TotalDowntime: downtime.Seconds(),
	})
}

// PrintRetryingToResolve print the message retrying to resolve,
// after n failed probes.
func (p *JSONPrinter) PrintRetryingToResolve(hostname string) {
	p.Print(JSONData{
		Type:     retryEvent,
		Message:  fmt.Sprintf("retrying to resolve %s", hostname),
		Hostname: hostname,
	})
}

// PrintInfo formats and prints an informational message in JSON format.
func (p *JSONPrinter) PrintInfo(format string, args ...any) {
	p.Print(JSONData{
		Type:    infoEvent,
		Message: fmt.Sprintf(format, args...),
	})
}

// PrintError formats and prints an error message in JSON format.
func (p *JSONPrinter) PrintError(format string, args ...any) {
	p.Print(JSONData{
		Type:    errorEvent,
		Message: fmt.Sprintf(format, args...),
	})
}
