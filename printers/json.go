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
	startEvent        JSONEventType = "start"        // Event type for `PrintStart` method.
	probeEvent        JSONEventType = "probe"        // Event type for both `PrintProbeSuccess` and `PrintProbeFail`.
	retryEvent        JSONEventType = "retry"        // Event type for `PrintRetryingToResolve` method.
	retrySuccessEvent JSONEventType = "retrySuccess" // Event type for `PrintTotalDowntime` method.
	statisticsEvent   JSONEventType = "statistics"   // Event type for `PrintStatistics` method.
	infoEvent         JSONEventType = "info"         // Event type for `PrintInfo` method.
	errorEvent        JSONEventType = "error"        // Event type for `PrintError` method.
)

// JSONData contains all possible fields for JSON output.
// Because one event usually contains only a subset of fields,
// other fields will be omitted in the output.
type JSONData struct {
	Type JSONEventType `json:"type"` // Specifies type of a message/event.
	// Success is a special field from probe messages, containing information
	// whether request was successful or not.
	// It's a pointer on purpose, otherwise success=false will be omitted,
	// but we still need to omit it for non-probe messages.
	Success                   *bool                  `json:"success,omitempty"`
	Timestamp                 string                 `json:"timestamp,omitempty"`
	Message                   string                 `json:"message"` // Message contains a message similar to other plain and colored printers.
	IPAddr                    string                 `json:"ipAddress,omitempty"`
	Hostname                  string                 `json:"hostname,omitempty"`
	Port                      uint16                 `json:"port,omitempty"`
	SourceAddr                string                 `json:"sourceAddress,omitempty"`
	HostnameResolveTries      uint                   `json:"hostnameResolveTries,omitempty"`
	HostnameChanges           []types.HostnameChange `json:"hostnameChanges,omitempty"`
	DestIsIP                  *bool                  `json:"destinationIsIP,omitempty"`
	Rtt                       string                 `json:"time,omitempty"`
	StartTimestamp            string                 `json:"startTimestamp,omitempty"` // StartTimestamp is used as a start time of TotalDuration for stats messages.
	EndTimestamp              string                 `json:"endTimestamp,omitempty"`   // EndTimestamp is used as an end of TotalDuration for stats messages.
	LastSuccessfulProbe       string                 `json:"lastSuccessfulProbe,omitempty"`
	LastUnsuccessfulProbe     string                 `json:"lastUnsuccessfulProbe,omitempty"`
	LongestUptimeStart        string                 `json:"longestUptimeStart,omitempty"`
	LongestUptimeEnd          string                 `json:"longestUptimeEnd,omitempty"`
	LongestDowntimeEnd        string                 `json:"longestDowntimeEnd,omitempty"`
	LongestDowntimeStart      string                 `json:"longestDowntimeStart,omitempty"`
	Latency                   float32                `json:"latency,omitempty"`         // Latency in ms for a successful probe messages.
	LatencyMin                string                 `json:"latencyMin,omitempty"`      // LatencyMin is a stringified 3 decimal places min latency for the stats event.
	LatencyAvg                string                 `json:"latencyAvg,omitempty"`      // LatencyAvg is a stringified 3 decimal places avg latency for the stats event.
	LatencyMax                string                 `json:"latencyMax,omitempty"`      // LatencyMax is a stringified 3 decimal places max latency for the stats event.
	TotalDuration             string                 `json:"totalDuration,omitempty"`   // TotalDuration is a total amount of seconds that program was running
	LongestUptime             string                 `json:"longestUptime,omitempty"`   // LongestUptime is the longest uptime in seconds.
	LongestDowntime           string                 `json:"longestDowntime,omitempty"` // LongestUptime is the longest uptime in seconds.
	TotalPacketLoss           string                 `json:"totalPacketLoss,omitempty"` // TotalPacketLoss in seconds.
	TotalPackets              uint                   `json:"totalPackets,omitempty"`
	OngoingSuccessfulProbes   uint                   `json:"ongoingSuccessfulProbes,omitempty"`
	OngoingUnsuccessfulProbes uint                   `json:"ongoingUnsuccessfulProbes,omitempty"`
	TotalUptime               float64                `json:"totalUptime,omitempty"`   // TotalUptime in seconds.
	TotalDowntime             float64                `json:"totalDowntime,omitempty"` // TotalDowntime in seconds.
}

// JSONPrinter is a struct that holds a JSON encoder to print structured JSON output.
type JSONPrinter struct {
	encoder *json.Encoder
	cfg     PrinterConfig
}

// NewJSONPrinter creates a new JSONPrinter instance.
// If prettify is true, the JSON output will be formatted with indentation.
func NewJSONPrinter(cfg PrinterConfig) *JSONPrinter {
	encoder := json.NewEncoder(os.Stdout)

	if cfg.PrettyJSON {
		encoder.SetIndent("", "\t")
	}

	return &JSONPrinter{encoder: encoder, cfg: cfg}
}

// PrintStart prints the initial message before doing probes.
func (p *JSONPrinter) PrintStart(hostname string, port uint16) {
	p.encoder.Encode(JSONData{
		Type:     startEvent,
		Message:  fmt.Sprintf("TCPinging %s on port %d", hostname, port),
		Hostname: hostname,
		Port:     port,
	})
}

// PrintProbeSuccess prints successful TCP probe replies in JSON format.
func (p *JSONPrinter) PrintProbeSuccess(startTime time.Time, sourceAddr string, opts types.Options, streak uint, rtt string) {
	if p.cfg.ShowFailuresOnly {
		return
	}

	// so that *bool fields not get omitted
	f := false
	t := true

	data := JSONData{
		Type:                    probeEvent,
		Hostname:                opts.Hostname,
		IPAddr:                  opts.IP.String(),
		Port:                    opts.Port,
		Rtt:                     rtt,
		DestIsIP:                &t,
		Success:                 &t,
		OngoingSuccessfulProbes: streak,
	}

	timestamp := ""
	if p.cfg.WithTimestamp {
		timestamp = startTime.Format(consts.TimeFormat)
	}

	if opts.Hostname == opts.IP.String() {
		data.Hostname = "" // to omit it from the output

		if timestamp == "" {
			if p.cfg.WithSourceAddress {
				data.Message = fmt.Sprintf("Reply from %s on port %d using %s TCP_conn=%d time=%s ms",
					opts.IP.String(),
					opts.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				data.Message = fmt.Sprintf("Reply from %s on port %d TCP_conn=%d time=%s ms",
					opts.IP.String(),
					opts.Port,
					streak,
					rtt)
			}
		} else {
			data.Timestamp = time.Now().Format(consts.TimeFormat)

			if p.cfg.WithSourceAddress {
				data.Message = fmt.Sprintf("%s Reply from %s on port %d using %s TCP_conn=%d time=%s ms",
					timestamp,
					opts.IP.String(),
					opts.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				data.Message = fmt.Sprintf("%s Reply from %s on port %d TCP_conn=%d time=%s ms",
					timestamp,
					opts.IP.String(),
					opts.Port,
					streak,
					rtt)
			}
		}
	} else {
		data.DestIsIP = &f

		if timestamp == "" {
			if p.cfg.WithSourceAddress {
				data.Message = fmt.Sprintf("Reply from %s (%s) on port %d using %s TCP_conn=%d time=%s ms",
					opts.Hostname,
					opts.IP.String(),
					opts.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				data.Message = fmt.Sprintf("Reply from %s (%s) on port %d TCP_conn=%d time=%s ms",
					opts.Hostname,
					opts.IP.String(),
					opts.Port,
					streak,
					rtt)
			}
		} else {
			data.Timestamp = time.Now().Format(consts.TimeFormat)

			if p.cfg.WithSourceAddress {
				data.Message = fmt.Sprintf("%s Reply from %s (%s) on port %d using %s TCP_conn=%d time=%s ms",
					timestamp,
					opts.Hostname,
					opts.IP.String(),
					opts.Port,
					sourceAddr,
					streak,
					rtt)
			} else {
				data.Message = fmt.Sprintf("%s Reply from %s (%s) on port %d TCP_conn=%d time=%s ms",
					timestamp,
					opts.Hostname,
					opts.IP.String(),
					opts.Port,
					streak,
					rtt)
			}
		}
	}

	p.encoder.Encode(data)
}

// PrintProbeFail prints a JSON message when a TCP probe fails.
func (p *JSONPrinter) PrintProbeFail(startTime time.Time, opts types.Options, streak uint) {
	// so that *bool fields not get omitted
	f := false
	t := true

	data := JSONData{
		Type:                      probeEvent,
		Hostname:                  opts.Hostname,
		IPAddr:                    opts.IP.String(),
		Port:                      opts.Port,
		DestIsIP:                  &t,
		Success:                   &f,
		OngoingUnsuccessfulProbes: streak,
	}

	timestamp := ""
	if p.cfg.WithTimestamp {
		timestamp = startTime.Format(consts.TimeFormat)
	}

	if opts.Hostname == opts.IP.String() {
		data.Hostname = "" // to omit it from the output

		if timestamp == "" {
			data.Message = fmt.Sprintf("No reply from %s on port %d",
				opts.IP.String(),
				opts.Port)
		} else {
			data.Message = fmt.Sprintf("%s No reply from %s on port %d",
				timestamp,
				opts.IP.String(),
				opts.Port)
		}
	} else {
		data.DestIsIP = &f

		if timestamp == "" {
			data.Message = fmt.Sprintf("No reply from %s (%s) on port %d",
				opts.Hostname,
				opts.IP.String(),
				opts.Port)
		} else {
			data.Message = fmt.Sprintf("%s No reply from %s (%s) on port %d",
				timestamp,
				opts.Hostname,
				opts.IP.String(),
				opts.Port)
		}
	}

	p.encoder.Encode(data)
}

// PrintTotalDownTime prints the total downtime,
// if the next retry was successful.
func (p *JSONPrinter) PrintTotalDownTime(downtime time.Duration) {
	p.encoder.Encode(JSONData{
		Type:          retrySuccessEvent,
		Message:       fmt.Sprintf("No response received for %s", utils.DurationToString(downtime)),
		TotalDowntime: downtime.Seconds(),
	})
}

// PrintRetryingToResolve print the message retrying to resolve,
// after n failed probes.
func (p *JSONPrinter) PrintRetryingToResolve(hostname string) {
	p.encoder.Encode(JSONData{
		Type:     retryEvent,
		Message:  fmt.Sprintf("Retrying to resolve %s", hostname),
		Hostname: hostname,
	})
}

// PrintInfo formats and prints an informational message in JSON format.
func (p *JSONPrinter) PrintInfo(format string, args ...any) {
	p.encoder.Encode(JSONData{
		Type:    infoEvent,
		Message: fmt.Sprintf(format, args...),
	})
}

// PrintError formats and prints an error message in JSON format.
func (p *JSONPrinter) PrintError(format string, args ...any) {
	p.encoder.Encode(JSONData{
		Type:    errorEvent,
		Message: fmt.Sprintf(format, args...),
	})
}

// PrintStatistics prints all gathered stats when program exits.
func (p *JSONPrinter) PrintStatistics(t types.Tcping) {
	data := JSONData{
		Type:                      statisticsEvent,
		IPAddr:                    t.Options.IP.String(),
		Hostname:                  t.Options.Hostname,
		Timestamp:                 time.Now().Format(consts.TimeFormat),
		StartTimestamp:            t.StartTime.Format(consts.TimeFormat),
		TotalDowntime:             t.TotalDowntime.Seconds(),
		TotalPackets:              t.TotalSuccessfulProbes + t.TotalUnsuccessfulProbes,
		OngoingSuccessfulProbes:   t.TotalSuccessfulProbes,
		OngoingUnsuccessfulProbes: t.TotalUnsuccessfulProbes,
		TotalUptime:               t.TotalUptime.Seconds(),
	}

	if !t.DestIsIP {
		data.Message = fmt.Sprintf("%s (%s) TCPing statistics - ",
			t.Options.Hostname,
			t.Options.IP)
	} else {
		data.Message = fmt.Sprintf("%s TCPing statistics - ", t.Options.IP)
	}

	data.Message += fmt.Sprintf("%d probes transmitted on port %d | %d received",
		data.TotalPackets,
		t.Options.Port,
		t.TotalSuccessfulProbes)

	if len(t.HostnameChanges) > 1 {
		data.HostnameChanges = t.HostnameChanges
	}

	loss := (float32(data.OngoingUnsuccessfulProbes) / float32(data.TotalPackets)) * 100
	if math.IsNaN(float64(loss)) {
		loss = 0
	}

	data.TotalPacketLoss = fmt.Sprintf("%.2f", loss)

	if !t.LastSuccessfulProbe.IsZero() {
		data.LastSuccessfulProbe = t.LastSuccessfulProbe.Format(consts.TimeFormat)
	}

	if !t.LastUnsuccessfulProbe.IsZero() {
		data.LastUnsuccessfulProbe = t.LastUnsuccessfulProbe.Format(consts.TimeFormat)
	}

	if t.LongestUptime.Duration != 0 {
		data.LongestUptime = fmt.Sprintf("%.0f", t.LongestUptime.Duration.Seconds())
		data.LongestUptimeStart = t.LongestUptime.Start.Format(consts.TimeFormat)
		data.LongestUptimeEnd = t.LongestUptime.End.Format(consts.TimeFormat)
	}

	if t.LongestDowntime.Duration != 0 {
		data.LongestDowntime = fmt.Sprintf("%.0f", t.LongestDowntime.Duration.Seconds())
		data.LongestDowntimeStart = t.LongestDowntime.Start.Format(consts.TimeFormat)
		data.LongestDowntimeEnd = t.LongestDowntime.End.Format(consts.TimeFormat)
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
		data.EndTimestamp = t.EndTime.Format(consts.TimeFormat)
	}

	totalDuration := t.TotalDowntime + t.TotalUptime
	data.TotalDuration = fmt.Sprintf("%.0f", totalDuration.Seconds())

	p.encoder.Encode(data)
}
