package printers

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/consts"
	"github.com/pouriyajamshidi/tcping/v2/internal/utils"
	"github.com/pouriyajamshidi/tcping/v2/types"
)

// JSONEventType is a special type for each method
// in the printer interface so that automatic tools
// can understand what kind of an event they've received.
// For instance, start vs probe vs statistics...
type JSONEventType string

const (
	startEvent        JSONEventType = "start"         // Event type for `PrintStart` method.
	probeEvent        JSONEventType = "probe"         // Event type for both `PrintProbeSuccess` and `PrintProbeFail`.
	retryEvent        JSONEventType = "retry"         // Event type for `PrintRetryingToResolve` method.
	retrySuccessEvent JSONEventType = "retry-success" // Event type for `PrintTotalDowntime` method.
	statisticsEvent   JSONEventType = "statistics"    // Event type for `PrintStatistics` method.
	infoEvent         JSONEventType = "info"          // Event type for `PrintInfo` method.
	versionEvent      JSONEventType = "version"       // Event type for `PrintVersion` method. // FIXME: This is unused
	errorEvent        JSONEventType = "error"         // Event type for `PrintError` method.
)

// JSONData contains all possible fields for JSON output.
// Because one event usually contains only a subset of fields,
// other fields will be omitted in the output.
type JSONData struct {
	Type                 JSONEventType          `json:"type"`    // Specifies type of a message/event.
	Message              string                 `json:"message"` // Message contains a human-readable message.
	Timestamp            string                 `json:"timestamp,omitempty"`
	IPAddr               string                 `json:"ipAddress,omitempty"`
	SourceAddr           string                 `json:"sourceAddress,omitempty"`
	Hostname             string                 `json:"hostname,omitempty"`
	HostnameResolveTries uint                   `json:"hostnameResolveTries,omitempty"`
	HostnameChanges      []types.HostnameChange `json:"hostnameChanges,omitempty"`
	DestIsIP             *bool                  `json:"destinationIsIP,omitempty"`
	Port                 uint16                 `json:"port,omitempty"`
	Rtt                  float32                `json:"time,omitempty"`

	// Success is a special field from probe messages, containing information
	// whether request was successful or not.
	// It's a pointer on purpose, otherwise success=false will be omitted,
	// but we still need to omit it for non-probe messages.
	Success                 *bool      `json:"success,omitempty"`
	StartTimestamp          *time.Time `json:"startTimestamp,omitempty"` // StartTimestamp is used as a start time of TotalDuration for stats messages.
	EndTimestamp            *time.Time `json:"endTimestamp,omitempty"`   // EndTimestamp is used as an end of TotalDuration for stats messages.
	LastSuccessfulProbe     *time.Time `json:"lastSuccessfulProbe,omitempty"`
	LastUnsuccessfulProbe   *time.Time `json:"lastUnsuccessfulProbe,omitempty"`
	LongestUptimeStart      *time.Time `json:"longestUptimeStart,omitempty"`
	LongestUptimeEnd        *time.Time `json:"longestUptimeEnd,omitempty"`
	LongestDowntimeEnd      *time.Time `json:"longestDowntimeEnd,omitempty"`
	LongestDowntimeStart    *time.Time `json:"longestDowntimeStart,omitempty"`
	Latency                 float32    `json:"latency,omitempty"`         // Latency in ms for a successful probe messages.
	LatencyMin              string     `json:"latencyMin,omitempty"`      // LatencyMin is a stringified 3 decimal places min latency for the stats event.
	LatencyAvg              string     `json:"latencyAvg,omitempty"`      // LatencyAvg is a stringified 3 decimal places avg latency for the stats event.
	LatencyMax              string     `json:"latencyMax,omitempty"`      // LatencyMax is a stringified 3 decimal places max latency for the stats event.
	TotalDuration           string     `json:"totalDuration,omitempty"`   // TotalDuration is a total amount of seconds that program was running
	LongestUptime           string     `json:"longestUptime,omitempty"`   // LongestUptime is the longest uptime in seconds.
	LongestDowntime         string     `json:"longestDowntime,omitempty"` // LongestUptime is the longest uptime in seconds.
	TotalPacketLoss         string     `json:"totalPacketLoss,omitempty"` // TotalPacketLoss in seconds. //TODO: is this a good comment?
	TotalPackets            uint       `json:"totalPackets,omitempty"`
	TotalSuccessfulProbes   uint       `json:"totalSuccessfulProbes,omitempty"`
	TotalUnsuccessfulProbes uint       `json:"totalUnsuccessfulProbes,omitempty"`
	TotalUptime             float64    `json:"totalUptime,omitempty"`   // TotalUptime in seconds.
	TotalDowntime           float64    `json:"totalDowntime,omitempty"` // TotalDowntime in seconds.
}

// JSONPrinter is a struct that holds a JSON encoder to print structured JSON output.
type JSONPrinter struct {
	encoder *json.Encoder
	cfg     PrinterConfig
}

// NewJSONPrinter creates a new JSONPrinter instance.
// If prettify is true, the JSON output will be formatted with indentation.
func NewJSONPrinter(opts PrinterConfig) *JSONPrinter {
	encoder := json.NewEncoder(os.Stdout)

	if opts.PrettyJSON {
		encoder.SetIndent("", "\t")
	}

	return &JSONPrinter{encoder: encoder, cfg: opts}
}

// Print is a little helper method for p.encoder.Encode.
// it also sets data.Timestamp to Now().
func (p *JSONPrinter) Print(data JSONData) {
	data.Timestamp = time.Now().Format(consts.TimeFormat)
	p.encoder.Encode(data)
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

// PrintProbeSuccess prints successful TCP probe replies in JSON format.
func (p *JSONPrinter) PrintProbeSuccess(startTime time.Time, sourceAddr string, opts types.Options, streak uint, rtt float32) {
	// so that *bool fields not get omitted
	f := false
	t := true

	data := JSONData{
		Type:                  probeEvent,
		Hostname:              opts.Hostname,
		IPAddr:                opts.IP.String(),
		Port:                  opts.Port,
		Rtt:                   rtt,
		DestIsIP:              &t,
		Success:               &t,
		TotalSuccessfulProbes: streak,
	}

	timestamp := ""
	if opts.ShowTimestamp {
		timestamp = startTime.Format(consts.TimeFormat)
	}

	// TODO: why timestamp is false when -D is provided?

	if opts.Hostname == opts.IP.String() {
		data.DestIsIP = &f
		if timestamp == "" {
			if opts.ShowSourceAddress {
				data.Message = fmt.Sprintf("Reply from %s on port %d using %s TCP_conn=%d time=%.3f ms",
					opts.IP.String(),
					opts.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				data.Message = fmt.Sprintf("Reply from %s on port %d TCP_conn=%d time=%.3f ms",
					opts.IP.String(),
					opts.Port,
					streak,
					rtt)
			}
		} else {
			if opts.ShowSourceAddress {
				data.Message = fmt.Sprintf("%s Reply from %s on port %d using %s TCP_conn=%d time=%.3f ms",
					timestamp,
					opts.IP.String(),
					opts.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				data.Message = fmt.Sprintf("%s Reply from %s on port %d TCP_conn=%d time=%.3f ms",
					timestamp,
					opts.IP.String(),
					opts.Port,
					streak,
					rtt)
			}
		}
	} else {
		if timestamp == "" {
			if opts.ShowSourceAddress {
				data.Message = fmt.Sprintf("Reply from %s (%s) on port %d using %s TCP_conn=%d time=%.3f ms",
					opts.Hostname,
					opts.IP.String(),
					opts.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				data.Message = fmt.Sprintf("Reply from %s (%s) on port %d TCP_conn=%d time=%.3f ms",
					opts.Hostname,
					opts.IP.String(),
					opts.Port,
					streak,
					rtt)
			}
		} else {
			if opts.ShowSourceAddress {
				data.Message = fmt.Sprintf("%s Reply from %s (%s) on port %d using %s TCP_conn=%d time=%.3f ms",
					timestamp,
					opts.Hostname,
					opts.IP.String(),
					opts.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				data.Message = fmt.Sprintf("%s Reply from %s (%s) on port %d TCP_conn=%d time=%.3f ms",
					timestamp,
					opts.Hostname,
					opts.IP.String(),
					opts.Port,
					streak,
					rtt)
			}
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
			IPAddr:                  opts.IP.String(),
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
		IPAddr:   t.Options.IP.String(),
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
		data.LongestUptimeStart = &t.LongestUptime.Start
		data.LongestUptimeEnd = &t.LongestUptime.End
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
		Message:       fmt.Sprintf("No response received for %s", utils.DurationToString(downtime)),
		TotalDowntime: downtime.Seconds(),
	})
}

// PrintRetryingToResolve print the message retrying to resolve,
// after n failed probes.
func (p *JSONPrinter) PrintRetryingToResolve(hostname string) {
	p.Print(JSONData{
		Type:     retryEvent,
		Message:  fmt.Sprintf("Retrying to resolve %s", hostname),
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
