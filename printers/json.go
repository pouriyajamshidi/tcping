package printers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pouriyajamshidi/tcping/v3/stats"
)

// JSONPrinter writes one JSON event per line, for piping a run into
// another program.
type JSONPrinter struct {
	encoder *json.Encoder
	cfg     Config
}

// NewJSONPrinter creates a JSON printer writing to the destination in cfg.
func NewJSONPrinter(cfg Config) *JSONPrinter {
	encoder := json.NewEncoder(writerOrStdout(cfg))

	if cfg.PrettyJSON {
		encoder.SetIndent("", "\t")
	}

	return &JSONPrinter{encoder: encoder, cfg: cfg}
}

type jsonEvent struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

type jsonStart struct {
	Hostname               string `json:"hostname"`
	Port                   uint16 `json:"port"`
	Protocol               string `json:"protocol"`
	NameResolutionDuration string `json:"nameResolutionDurationMs,omitempty"`
}

type jsonProbe struct {
	Hostname    string  `json:"hostname,omitempty"`
	IP          string  `json:"ipAddress"`
	Port        uint16  `json:"port"`
	Success     bool    `json:"success"`
	Latency     float32 `json:"latency,omitempty"`
	Source      string  `json:"sourceAddress,omitempty"`
	Connections uint    `json:"connections"`
	// How long the target had been up, or down, when this probe ended it.
	// Left out of every probe that ended neither, which is most of them.
	EndedUptime   string    `json:"endedUptime,omitempty"`
	EndedDowntime string    `json:"endedDowntime,omitempty"`
	Timestamp     string    `json:"timestamp,omitempty"`
	HTTP          *jsonHTTP `json:"http,omitempty"`
	UDP           *jsonUDP  `json:"udp,omitempty"`
}

// jsonUDP is attached to a probe only when the target is UDP, so the output
// of the other probe types is unchanged. It is what tells a reply apart from
// a refusal and from silence, which for UDP is the whole story.
type jsonUDP struct {
	ProbeNumber uint64 `json:"probeNumber"`
	Rejected    bool   `json:"rejected"`
	Echoed      bool   `json:"echoed"`
	ReplyBytes  int    `json:"replyBytes"`
}

func newJSONUDP(s *stats.Statistics) *jsonUDP {
	if !s.IsUDP() {
		return nil
	}

	return &jsonUDP{
		ProbeNumber: s.UDP.ProbeNumber,
		Rejected:    s.UDP.Rejected,
		Echoed:      s.UDP.Echoed,
		ReplyBytes:  s.UDP.ReplySize,
	}
}

// jsonHTTP is attached to a probe only when the target is HTTP(S) and the
// server actually responded, so TCP output is unchanged.
type jsonHTTP struct {
	StatusCode     int    `json:"statusCode"`
	Status         string `json:"status"`
	Version        string `json:"version"`
	TLSVersion     string `json:"tlsVersion,omitempty"`
	TLSCipherSuite string `json:"tlsCipherSuite,omitempty"`
	CertExpiry     string `json:"certificateExpiry,omitempty"`
	ConnectMs      string `json:"connectMs"`
	TLSMs          string `json:"tlsHandshakeMs,omitempty"`
	TTFBMs         string `json:"timeToFirstByteMs"`
}

func newJSONHTTP(s *stats.Statistics) *jsonHTTP {
	if !s.IsHTTP() || !s.HasHTTPResponse() {
		return nil
	}

	h := &jsonHTTP{
		StatusCode: s.HTTP.StatusCode,
		Status:     s.HTTP.Status,
		Version:    s.HTTP.Proto,
		ConnectMs:  s.ConnectDurationStr(),
		TTFBMs:     s.TimeToFirstByteStr(),
	}

	if s.HTTP.TLSVersion != "" {
		h.TLSVersion = s.HTTP.TLSVersion
		h.TLSCipherSuite = s.HTTP.TLSCipherSuite
		h.CertExpiry = s.CertExpiryStr()
		h.TLSMs = s.TLSDurationStr()
	}

	return h
}

type jsonRetry struct {
	Hostname string `json:"hostname"`
}

type jsonNameResolution struct {
	DurationMs string `json:"durationMs"`
}

type jsonError struct {
	Message string `json:"message"`
}

type jsonHostnameChange struct {
	Addr       string `json:"addr"`
	When       string `json:"when"`
	DurationMs string `json:"durationMs"`
}

type jsonStatistics struct {
	Hostname               string               `json:"hostname,omitempty"`
	IP                     string               `json:"ipAddress"`
	Port                   uint16               `json:"port"`
	Protocol               string               `json:"protocol"`
	TotalProbes            uint                 `json:"totalProbes"`
	SuccessfulProbes       uint                 `json:"successfulProbes"`
	UnsuccessfulProbes     uint                 `json:"unsuccessfulProbes"`
	PacketLoss             float32              `json:"packetLossPercent"`
	LastSuccessfulProbe    string               `json:"lastSuccessfulProbe,omitempty"`
	LastUnsuccessfulProbe  string               `json:"lastUnsuccessfulProbe,omitempty"`
	TotalUptime            string               `json:"totalUptime"`
	TotalDowntime          string               `json:"totalDowntime"`
	LongestUptime          string               `json:"longestUptime,omitempty"`
	LongestDowntime        string               `json:"longestDowntime,omitempty"`
	HostnameResolveRetries uint                 `json:"hostnameResolveRetries,omitempty"`
	HostnameChanges        []jsonHostnameChange `json:"hostnameChanges,omitempty"`
	LatencyMin             float32              `json:"latencyMin,omitempty"`
	LatencyAvg             float32              `json:"latencyAvg,omitempty"`
	LatencyMax             float32              `json:"latencyMax,omitempty"`
	LatencyMdev            float32              `json:"latencyMdev,omitempty"`
	StartTime              string               `json:"startTime"`
	EndTime                string               `json:"endTime,omitempty"`
	Duration               string               `json:"duration"`
}

func (p *JSONPrinter) encode(event string, data any) {
	_ = p.encoder.Encode(jsonEvent{
		Type: event,
		Data: data,
	})
}

// PrintStart emits the event that opens a run.
func (p *JSONPrinter) PrintStart(s *stats.Statistics) {
	start := jsonStart{
		Hostname: s.Hostname,
		Port:     s.Port,
		Protocol: s.ProtocolStr(),
	}
	if !s.DestIsIP {
		start.NameResolutionDuration = s.NameResolutionDurationStr()
	}
	p.encode("start", start)
}

// PrintNameResolutionDuration prints how long the initial hostname resolution took.
func (p *JSONPrinter) PrintNameResolutionDuration(s *stats.Statistics) {
	p.encode("nameResolution", jsonNameResolution{
		DurationMs: s.NameResolutionDurationStr(),
	})
}

// PrintProbeSuccess emits a successful probe.
func (p *JSONPrinter) PrintProbeSuccess(s *stats.Statistics) {
	hostname := s.Hostname
	if s.DestIsIP {
		hostname = ""
	}

	latency, _ := strconv.ParseFloat(s.RTTStr(), 32)

	data := jsonProbe{
		Hostname:    hostname,
		IP:          s.IPStr(),
		Port:        s.Port,
		Success:     true,
		Latency:     float32(latency),
		Connections: s.OngoingSuccessfulProbes,
		HTTP:        newJSONHTTP(s),
		UDP:         newJSONUDP(s),
	}

	if s.EndedDowntime != 0 {
		data.EndedDowntime = s.EndedDowntimeDuration()
	}

	if p.cfg.WithTimestamp {
		data.Timestamp = s.CurrentTimestamp()
	}

	if p.cfg.WithSourceAddress && s.SourceAddr() != "" {
		data.Source = s.SourceAddr()
	}

	p.encode("probe", data)
}

// PrintProbeFailure emits a failed probe.
func (p *JSONPrinter) PrintProbeFailure(s *stats.Statistics) {
	hostname := s.Hostname
	if s.DestIsIP {
		hostname = ""
	}

	data := jsonProbe{
		Hostname:    hostname,
		IP:          s.IPStr(),
		Port:        s.Port,
		Success:     false,
		Connections: s.OngoingUnsuccessfulProbes,
		HTTP:        newJSONHTTP(s),
		UDP:         newJSONUDP(s),
	}

	if s.EndedUptime != 0 {
		data.EndedUptime = s.EndedUptimeDuration()
	}

	if p.cfg.WithTimestamp {
		data.Timestamp = s.CurrentTimestamp()
	}

	if p.cfg.WithSourceAddress && s.SourceAddr() != "" {
		data.Source = s.SourceAddr()
	}

	p.encode("probe", data)
}

// PrintStatistics emits the summary of the run so far.
func (p *JSONPrinter) PrintStatistics(s *stats.Statistics) {
	hostname := s.Hostname
	if s.DestIsIP {
		hostname = ""
	}

	data := jsonStatistics{
		Hostname:           hostname,
		IP:                 s.IPStr(),
		Port:               s.Port,
		Protocol:           s.ProtocolStr(),
		TotalProbes:        s.TotalProbes(),
		SuccessfulProbes:   s.TotalSuccessfulProbes,
		UnsuccessfulProbes: s.TotalUnsuccessfulProbes,
		PacketLoss:         s.PacketLoss(),
		TotalUptime:        s.TotalUptimeDuration(),
		TotalDowntime:      s.TotalDowntimeDuration(),
		StartTime:          s.StartTimeFormatted(),
		Duration:           s.RuntimeDuration(),
	}

	if !s.LastSuccessfulProbe.IsZero() {
		data.LastSuccessfulProbe = s.LastSuccessfulProbeFormatted()
	}

	if !s.LastUnsuccessfulProbe.IsZero() {
		data.LastUnsuccessfulProbe = s.LastUnsuccessfulProbeFormatted()
	}

	if s.LongestUptime.Duration != 0 {
		data.LongestUptime = s.LongestUptimeDuration()
	}

	if s.LongestDowntime.Duration != 0 {
		data.LongestDowntime = s.LongestDowntimeDuration()
	}

	if !s.DestIsIP {
		data.HostnameResolveRetries = s.RetriedHostnameLookups

		if len(s.HostnameChanges) > 1 {
			for _, change := range s.HostnameChanges {
				data.HostnameChanges = append(data.HostnameChanges, jsonHostnameChange{
					Addr:       change.Addr.String(),
					When:       change.WhenFormatted(),
					DurationMs: change.DurationStr(),
				})
			}
		}
	}

	if s.TotalSuccessfulProbes > 0 {
		data.LatencyMin = s.RTTResults.Min
		data.LatencyAvg = s.RTTResults.Average
		data.LatencyMax = s.RTTResults.Max
		data.LatencyMdev = s.RTTResults.Mdev
	}

	if !s.EndTime.IsZero() {
		data.EndTime = s.EndTimeFormatted()
	}

	p.encode("statistics", data)
}

// PrintRetryingToResolve emits the hostname being looked up again.
func (p *JSONPrinter) PrintRetryingToResolve(hostname string) {
	p.encode("retry", jsonRetry{
		Hostname: hostname,
	})
}

// PrintError emits an error as an event, so it stays in the same stream.
func (p *JSONPrinter) PrintError(format string, args ...any) {
	p.encode("error", jsonError{
		Message: fmt.Sprintf(format, args...),
	})
}

// Shutdown prints statistics. Statistics are already finalized by
// finalizeStatistics by the time this runs. It does not exit the program -
// that decision belongs to the caller, not the printer.
func (p *JSONPrinter) Shutdown(s *stats.Statistics) {
	if !p.cfg.OmitStatistics {
		p.PrintStatistics(s)
	}
}
