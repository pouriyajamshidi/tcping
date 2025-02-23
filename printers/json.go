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
	startEvent        EventType = "start"        // Event type for `PrintStart` method.
	probeEvent        EventType = "probe"        // Event type for both `PrintProbeSuccess` and `PrintProbeFail`.
	retryEvent        EventType = "retry"        // Event type for `PrintRetryingToResolve` method.
	retrySuccessEvent EventType = "retrySuccess" // Event type for `PrintTotalDowntime` method.
	statisticsEvent   EventType = "statistics"   // Event type for `PrintStatistics` method.
	infoEvent         EventType = "info"         // Event type for `PrintInfo` method.
	errorEvent        EventType = "error"        // Event type for `PrintError` method.
)

// JSONData contains all possible fields for JSON output.
// Because one event usually contains only a subset of fields,
// other fields will be omitted in the output.
type JSONData struct {
	Type EventType `json:"type"` // Specifies type of a message/event.
	// Success is a special field from probe messages, containing information
	// whether request was successful or not.
	// It's a pointer on purpose, otherwise success=false will be omitted,
	// but we still need to omit it for non-probe messages.
	Success                         *bool                  `json:"success,omitempty"`
	Timestamp                       string                 `json:"timestamp,omitempty"`
	Message                         string                 `json:"message"` // Message contains a message similar to other plain and colored printers.
	IPAddr                          string                 `json:"ipAddress,omitempty"`
	Hostname                        string                 `json:"hostname,omitempty"`
	Port                            uint16                 `json:"port,omitempty"`
	TotalDuration                   string                 `json:"totalDuration,omitempty"` // TotalDuration is a total amount of seconds that program was running
	TotalUptime                     string                 `json:"totalUptime,omitempty"`   // TotalUptime in seconds.
	TotalDowntime                   string                 `json:"totalDowntime,omitempty"` // TotalDowntime in seconds.
	TotalPackets                    uint                   `json:"totalPackets,omitempty"`
	TotalSuccessfulPackets          uint                   `json:"totalSuccessfulPackets,omitempty"`
	TotalUnsuccessfulPackets        uint                   `json:"totalUnsuccessfulPackets,omitempty"`
	TotalPacketLossPercent          string                 `json:"totalPacketLossPercent,omitempty"` // TotalPacketLoss in seconds.
	LongestUptime                   string                 `json:"longestUptime,omitempty"`          // LongestUptime is the longest uptime in seconds.
	LongestDowntime                 string                 `json:"longestDowntime,omitempty"`        // LongestDowntime is the longest downtime in seconds.
	SourceAddr                      string                 `json:"sourceAddress,omitempty"`
	HostnameResolveRetries          uint                   `json:"hostnameResolveRetries,omitempty"`
	HostnameChanges                 []types.HostnameChange `json:"hostnameChanges,omitempty"`
	DestIsIP                        *bool                  `json:"destinationIsIP,omitempty"`
	Time                            string                 `json:"time,omitempty"`
	LastSuccessfulProbe             string                 `json:"lastSuccessfulProbe,omitempty"`
	LastUnsuccessfulProbe           string                 `json:"lastUnsuccessfulProbe,omitempty"`
	LongestConsecutiveUptimeStart   string                 `json:"longestConsecutiveUptimeStart,omitempty"`
	LongestConsecutiveUptimeEnd     string                 `json:"longestConsecutiveUptimeEnd,omitempty"`
	LongestConsecutiveDowntimeStart string                 `json:"longestConsecutiveDowntimeStart,omitempty"`
	LongestConsecutiveDowntimeEnd   string                 `json:"longestConsecutiveDowntimeEnd,omitempty"`
	Latency                         float32                `json:"latency,omitempty"`    // Latency in ms for a successful probe messages.
	LatencyMin                      string                 `json:"latencyMin,omitempty"` // LatencyMin is a stringified 3 decimal places min latency for the stats event.
	LatencyAvg                      string                 `json:"latencyAvg,omitempty"` // LatencyAvg is a stringified 3 decimal places avg latency for the stats event.
	LatencyMax                      string                 `json:"latencyMax,omitempty"` // LatencyMax is a stringified 3 decimal places max latency for the stats event.
	OngoingSuccessfulProbes         uint                   `json:"ongoingSuccessfulProbes,omitempty"`
	OngoingUnsuccessfulProbes       uint                   `json:"ongoingUnsuccessfulProbes,omitempty"`
	StartTimestamp                  string                 `json:"startTime,omitempty"` // StartTime is used as a start time of TotalDuration for stats messages.
	EndTimestamp                    string                 `json:"endTime,omitempty"`   // EndTime is used as an end of TotalDuration for stats messages.
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
		Time:                    rtt,
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
			data.Timestamp = timestamp

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
			data.Timestamp = timestamp

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

// PrintProbeFailure prints a JSON message when a TCP probe fails.
func (p *JSONPrinter) PrintProbeFailure(startTime time.Time, opts types.Options, streak uint) {
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
		TotalDowntime: utils.DurationToString(downtime),
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
		Type:                     statisticsEvent,
		IPAddr:                   t.Options.IP.String(),
		Port:                     t.Options.Port,
		Hostname:                 t.Options.Hostname,
		TotalSuccessfulPackets:   t.TotalSuccessfulProbes,
		TotalUnsuccessfulPackets: t.TotalUnsuccessfulProbes,
		Timestamp:                time.Now().Format(consts.TimeFormat),
		StartTimestamp:           t.StartTime.Format(consts.TimeFormat),
		TotalUptime:              utils.DurationToString(t.TotalUptime),
		TotalDowntime:            utils.DurationToString(t.TotalDowntime),
		TotalPackets:             t.TotalSuccessfulProbes + t.TotalUnsuccessfulProbes,
	}

	if !t.DestIsIP {
		data.Message = fmt.Sprintf("%s (%s) TCPing statistics - ",
			t.Options.Hostname,
			t.Options.IP)
	} else {
		data.Message = fmt.Sprintf("%s TCPing statistics - ", t.Options.IP)
		data.Hostname = "" // to omit hostname from the output
	}

	data.Message += fmt.Sprintf("%d probes transmitted on port %d | %d received",
		data.TotalPackets,
		t.Options.Port,
		t.TotalSuccessfulProbes)

	if len(t.HostnameChanges) > 1 {
		data.HostnameChanges = t.HostnameChanges
	}

	totalPackets := t.TotalSuccessfulProbes + t.TotalUnsuccessfulProbes
	packetLoss := (float32(t.TotalUnsuccessfulProbes) / float32(totalPackets)) * 100

	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	data.TotalPacketLossPercent = fmt.Sprintf("%.2f", packetLoss)

	if !t.LastSuccessfulProbe.IsZero() {
		data.LastSuccessfulProbe = t.LastSuccessfulProbe.Format(consts.TimeFormat)
	}

	if !t.LastUnsuccessfulProbe.IsZero() {
		data.LastUnsuccessfulProbe = t.LastUnsuccessfulProbe.Format(consts.TimeFormat)
	}

	if t.LongestUptime.Duration != 0 {
		data.LongestUptime = fmt.Sprintf("%.0f", t.LongestUptime.Duration.Seconds())
		data.LongestConsecutiveUptimeStart = t.LongestUptime.Start.Format(consts.TimeFormat)
		data.LongestConsecutiveUptimeEnd = t.LongestUptime.End.Format(consts.TimeFormat)
	}

	if t.LongestDowntime.Duration != 0 {
		data.LongestDowntime = fmt.Sprintf("%.0f", t.LongestDowntime.Duration.Seconds())
		data.LongestConsecutiveDowntimeStart = t.LongestDowntime.Start.Format(consts.TimeFormat)
		data.LongestConsecutiveDowntimeEnd = t.LongestDowntime.End.Format(consts.TimeFormat)
	}

	if !t.DestIsIP {
		data.HostnameResolveRetries = t.RetriedHostnameLookups
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
