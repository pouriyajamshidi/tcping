package printers

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/utils"
	"github.com/pouriyajamshidi/tcping/v3/probes/statistics"
	"github.com/pouriyajamshidi/tcping/v3/types"
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
	LongestUp                       string                 `json:"longestUp,omitempty"`              // LongestUp is the longest uptime in seconds.
	LongestDown                     string                 `json:"longestDowntime,omitempty"`        // LongestDown is the longest downtime in seconds.
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
}

// NewJSONPrinter creates a new JSONPrinter instance.
// If prettify is true, the JSON output will be formatted with indentation.
func NewJSONPrinter(pretty bool) *JSONPrinter {
	encoder := json.NewEncoder(os.Stdout)

	if pretty {
		encoder.SetIndent("", "\t")
	}

	return &JSONPrinter{encoder: encoder}
}

// Shutdown sets the end time, prints statistics, and exits the program.
func (p *JSONPrinter) Shutdown(s *statistics.Statistics) {
	s.EndTime = time.Now()
	PrintStats(p, s)
	os.Exit(0)
}

// PrintStart prints the initial message before doing probes.
func (p *JSONPrinter) PrintStart(s *statistics.Statistics) {
	p.encoder.Encode(JSONData{
		Type:     startEvent,
		Message:  fmt.Sprintf("TCPinging %s on port %d", s.Hostname, s.Port),
		Hostname: s.Hostname,
		Port:     s.Port,
	})
}

// PrintProbeSuccess prints successful TCP probe replies in JSON format.
func (p *JSONPrinter) PrintProbeSuccess(s *statistics.Statistics) {
	// so that *bool fields do not get omitted
	f := false
	t := true

	data := JSONData{
		Type:                    probeEvent,
		Hostname:                s.Hostname,
		IPAddr:                  s.IPStr(),
		Port:                    s.Port,
		Time:                    s.RTTStr(),
		DestIsIP:                &t,
		Success:                 &t,
		OngoingSuccessfulProbes: s.OngoingSuccessfulProbes,
	}

	timestamp := ""
	if s.WithTimestamp {
		timestamp = s.StartTimeFormatted()
	}

	if s.Hostname == s.IPStr() {
		data.Hostname = "" // to omit it from the output

		if timestamp == "" {
			if s.WithSourceAddress {
				data.Message = fmt.Sprintf("Reply from %s on port %d using %s TCP_conn=%d time=%s ms",
					s.IP.String(),
					s.Port,
					s.SourceAddr(),
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			} else {
				data.Message = fmt.Sprintf("Reply from %s on port %d TCP_conn=%d time=%s ms",
					s.IP.String(),
					s.Port,
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			}
		} else {
			data.Timestamp = timestamp

			if s.WithSourceAddress {
				data.Message = fmt.Sprintf("%s Reply from %s on port %d using %s TCP_conn=%d time=%s ms",
					timestamp,
					s.IP.String(),
					s.Port,
					s.SourceAddr(),
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			} else {
				data.Message = fmt.Sprintf("%s Reply from %s on port %d TCP_conn=%d time=%s ms",
					timestamp,
					s.IP.String(),
					s.Port,
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			}
		}
	} else {
		data.DestIsIP = &f

		if timestamp == "" {
			if s.WithSourceAddress {
				data.Message = fmt.Sprintf("Reply from %s (%s) on port %d using %s TCP_conn=%d time=%s ms",
					s.Hostname,
					s.IP.String(),
					s.Port,
					s.SourceAddr(),
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			} else {
				data.Message = fmt.Sprintf("Reply from %s (%s) on port %d TCP_conn=%d time=%s ms",
					s.Hostname,
					s.IP.String(),
					s.Port,
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			}
		} else {
			data.Timestamp = timestamp

			if s.WithSourceAddress {
				data.Message = fmt.Sprintf("%s Reply from %s (%s) on port %d using %s TCP_conn=%d time=%s ms",
					timestamp,
					s.Hostname,
					s.IP.String(),
					s.Port,
					s.SourceAddr(),
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			} else {
				data.Message = fmt.Sprintf("%s Reply from %s (%s) on port %d TCP_conn=%d time=%s ms",
					timestamp,
					s.Hostname,
					s.IP.String(),
					s.Port,
					s.OngoingSuccessfulProbes,
					s.RTTStr())
			}
		}
	}

	p.encoder.Encode(data)
}

// PrintProbeFailure prints a JSON message when a TCP probe fails.
func (p *JSONPrinter) PrintProbeFailure(s *statistics.Statistics) {
	// so that *bool fields not get omitted
	f := false
	t := true

	data := JSONData{
		Type:                      probeEvent,
		Hostname:                  s.Hostname,
		IPAddr:                    s.IP.String(),
		Port:                      s.Port,
		DestIsIP:                  &t,
		Success:                   &f,
		OngoingUnsuccessfulProbes: s.OngoingUnsuccessfulProbes,
	}

	timestamp := ""
	if s.WithTimestamp {
		timestamp = s.StartTimeFormatted()
	}

	if s.Hostname == s.IP.String() {
		data.Hostname = "" // to omit it from the output

		if timestamp == "" {
			data.Message = fmt.Sprintf("No reply from %s on port %d",
				s.IP.String(),
				s.Port)
		} else {
			data.Message = fmt.Sprintf("%s No reply from %s on port %d",
				timestamp,
				s.IP.String(),
				s.Port)
		}
	} else {
		data.DestIsIP = &f

		if timestamp == "" {
			data.Message = fmt.Sprintf("No reply from %s (%s) on port %d",
				s.Hostname,
				s.IP.String(),
				s.Port)
		} else {
			data.Message = fmt.Sprintf("%s No reply from %s (%s) on port %d",
				timestamp,
				s.Hostname,
				s.IP.String(),
				s.Port)
		}
	}

	p.encoder.Encode(data)
}

// PrintTotalDownTime prints the total downtime,
// if the next retry was successful.
func (p *JSONPrinter) PrintTotalDownTime(s *statistics.Statistics) {
	p.encoder.Encode(JSONData{
		Type:          retrySuccessEvent,
		Message:       fmt.Sprintf("No response received for %s", utils.DurationToString(s.DownTime)),
		TotalDowntime: utils.DurationToString(s.TotalDowntime),
	})
}

// PrintRetryingToResolve print the message retrying to resolve,
// after n failed probes.
func (p *JSONPrinter) PrintRetryingToResolve(s *statistics.Statistics) {
	p.encoder.Encode(JSONData{
		Type:     retryEvent,
		Message:  fmt.Sprintf("Retrying to resolve %s", s.Hostname),
		Hostname: s.Hostname,
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
func (p *JSONPrinter) PrintStatistics(s *statistics.Statistics) {
	data := JSONData{
		Type:                     statisticsEvent,
		IPAddr:                   s.IPStr(),
		Port:                     s.Port,
		Hostname:                 s.Hostname,
		TotalSuccessfulPackets:   s.TotalSuccessfulProbes,
		TotalUnsuccessfulPackets: s.TotalUnsuccessfulProbes,
		Timestamp:                time.Now().Format(time.DateTime),
		StartTimestamp:           s.StartTime.Format(time.DateTime),
		TotalUptime:              utils.DurationToString(s.TotalUptime),
		TotalDowntime:            utils.DurationToString(s.TotalDowntime),
		TotalPackets:             s.TotalSuccessfulProbes + s.TotalUnsuccessfulProbes,
	}

	if !s.DestIsIP {
		data.Message = fmt.Sprintf("%s (%s) TCPing statistics - ",
			s.Hostname,
			s.IP)
	} else {
		data.Message = fmt.Sprintf("%s TCPing statistics - ", s.IP)
		data.Hostname = "" // to omit hostname from the output
	}

	data.Message += fmt.Sprintf("%d probes transmitted on port %d | %d received",
		data.TotalPackets,
		s.Port,
		s.TotalSuccessfulProbes)

	if len(s.HostnameChanges) > 1 {
		data.HostnameChanges = s.HostnameChanges
	}

	totalPackets := s.TotalSuccessfulProbes + s.TotalUnsuccessfulProbes
	packetLoss := (float32(s.TotalUnsuccessfulProbes) / float32(totalPackets)) * 100

	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	data.TotalPacketLossPercent = fmt.Sprintf("%.2f", packetLoss)

	if !s.LastSuccessfulProbe.IsZero() {
		data.LastSuccessfulProbe = s.LastSuccessfulProbe.Format(time.DateTime)
	}

	if !s.LastUnsuccessfulProbe.IsZero() {
		data.LastUnsuccessfulProbe = s.LastUnsuccessfulProbe.Format(time.DateTime)
	}

	if s.LongestUp.Duration != 0 {
		data.LongestUp = fmt.Sprintf("%.0f", s.LongestUp.Duration.Seconds())
		data.LongestConsecutiveUptimeStart = s.LongestUp.Start.Format(time.DateTime)
		data.LongestConsecutiveUptimeEnd = s.LongestUp.End.Format(time.DateTime)
	}

	if s.LongestDown.Duration != 0 {
		data.LongestDown = fmt.Sprintf("%.0f", s.LongestDown.Duration.Seconds())
		data.LongestConsecutiveDowntimeStart = s.LongestDown.Start.Format(time.DateTime)
		data.LongestConsecutiveDowntimeEnd = s.LongestDown.End.Format(time.DateTime)
	}

	if !s.DestIsIP {
		data.HostnameResolveRetries = s.RetriedHostnameLookups
	}

	if s.RTTResults.HasResults {
		data.LatencyMin = fmt.Sprintf("%.3f", s.RTTResults.Min)
		data.LatencyAvg = fmt.Sprintf("%.3f", s.RTTResults.Average)
		data.LatencyMax = fmt.Sprintf("%.3f", s.RTTResults.Max)
	}

	if !s.EndTime.IsZero() {
		data.EndTimestamp = s.EndTime.Format(time.DateTime)
	}

	totalDuration := s.TotalDowntime + s.TotalUptime
	data.TotalDuration = fmt.Sprintf("%.0f", totalDuration.Seconds())

	p.encoder.Encode(data)
}
