package printers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/version"

	"github.com/pouriyajamshidi/tcping/v3/stats"
)

// The OTLP JSON payload we POST to Alloy. It is the same envelope every
// time: one resource that says who we are, one scope, and the metrics we
// filled in for this event. Hand-written because OTLP over HTTP accepts
// plain JSON, so there is no need for protobuf or an SDK.
type otlpPayload struct {
	ResourceMetrics []otlpResourceMetrics `json:"resourceMetrics"`
}

type otlpResourceMetrics struct {
	Resource     otlpResource       `json:"resource"`
	ScopeMetrics []otlpScopeMetrics `json:"scopeMetrics"`
}

type otlpResource struct {
	Attributes []otlpAttr `json:"attributes"`
}

type otlpScopeMetrics struct {
	Metrics []otlpMetric `json:"metrics"`
}

// otlpMetric carries either a gauge or a sum, never both. A gauge is a value
// that goes up and down (a latency), a sum is a running total that only goes
// up (the probe count).
type otlpMetric struct {
	Name  string      `json:"name"`
	Unit  string      `json:"unit,omitempty"`
	Gauge *otlpPoints `json:"gauge,omitempty"`
	Sum   *otlpSum    `json:"sum,omitempty"`
}

type otlpPoints struct {
	DataPoints []otlpPoint `json:"dataPoints"`
}

type otlpSum struct {
	DataPoints             []otlpPoint `json:"dataPoints"`
	AggregationTemporality int         `json:"aggregationTemporality"`
	IsMonotonic            bool        `json:"isMonotonic"`
}

type otlpPoint struct {
	Value      float64    `json:"asDouble"`
	StartTime  string     `json:"startTimeUnixNano,omitempty"`
	Time       string     `json:"timeUnixNano"`
	Attributes []otlpAttr `json:"attributes,omitempty"`
}

type otlpAttr struct {
	Key   string        `json:"key"`
	Value otlpAttrValue `json:"value"`
}

type otlpAttrValue struct {
	String string `json:"stringValue"`
}

const (
	// OTLP's word for "this point carries the running total since
	// StartTime", which is exactly what our counters already hold.
	otlpCumulative = 2

	// Probes are usually a second apart, so a send that hangs longer than
	// this would hold up the next probe.
	alloyTimeout = 2 * time.Second

	otlpMetricsPath = "/v1/metrics"

	// Used when the caller did not ask for an interval, so a zero value
	// does not end up meaning "send the summary with every probe".
	defaultAlloyStatsInterval = 10 * time.Second
)

func attr(key, value string) otlpAttr {
	return otlpAttr{Key: key, Value: otlpAttrValue{String: value}}
}

// unixNano is how OTLP JSON wants a timestamp: nanoseconds since the epoch,
// as a string, because the number does not survive a JSON float.
func unixNano(t time.Time) string {
	return strconv.FormatInt(t.UnixNano(), 10)
}

// AlloyPrinter sends probe results to Grafana Alloy as OTLP metrics instead
// of printing them. Alloy forwards them to Prometheus, so a run shows up as
// a graph rather than as lines of text.
type AlloyPrinter struct {
	client    *http.Client
	endpoint  string
	startTime time.Time // Beginning of the run, which every counter is measured from.
	lastStats time.Time // When the run summary was last sent. Zero, so the first probe carries one.
	// How often the run summary rides along with a probe. Without it the
	// summary would only be sent when you press Enter or when tcping exits,
	// which never happens on a long run that no one is watching.
	statsInterval time.Duration
	warned        bool // Whether we already complained about a send that failed.
	cfg           Config
}

// NewAlloyPrinter creates an AlloyPrinter pointed at the given Alloy address.
// The address can be given with or without the OTLP path, so both
// "http://localhost:4318" and "http://localhost:4318/v1/metrics" work.
func NewAlloyPrinter(cfg Config) *AlloyPrinter {
	endpoint := cfg.AlloyURL

	if !strings.Contains(endpoint, "://") {
		endpoint = "http://" + endpoint
	}

	if !strings.HasSuffix(endpoint, otlpMetricsPath) {
		endpoint = strings.TrimSuffix(endpoint, "/") + otlpMetricsPath
	}

	statsInterval := cfg.AlloyStatsInterval
	if statsInterval <= 0 {
		statsInterval = defaultAlloyStatsInterval
	}

	return &AlloyPrinter{
		client:        &http.Client{Timeout: alloyTimeout},
		endpoint:      endpoint,
		startTime:     time.Now(),
		statsInterval: statsInterval,
		cfg:           cfg,
	}
}

// labels are the attributes put on every data point, which is what makes
// them Prometheus labels on the other end. Anything extra for a single
// metric is passed in, and each call builds its own slice so two callers
// can never scribble over each other's labels.
//
// The source label says which machine the probe was sent from. It has to be
// a data point attribute rather than a resource attribute, because Alloy's
// Prometheus exporter puts resource attributes on a separate target_info
// metric instead of on the series itself.
//
// The resolved IP is deliberately not here. It is part of what identifies a
// series, so a hostname that resolves to a different address mid-run would
// start a new series and leave the old one behind. It is sent on its own as
// tcping_target_address instead.
func (p *AlloyPrinter) labels(s *stats.Statistics, extra ...otlpAttr) []otlpAttr {
	labels := []otlpAttr{
		attr("source", p.cfg.SourceLabel),
		attr("target", s.Hostname),
		attr("port", s.PortStr()),
		attr("protocol", s.ProtocolStr()),
	}

	return append(labels, extra...)
}

// gauge is a single measurement taken right now.
func (p *AlloyPrinter) gauge(name, unit string, value float64, labels []otlpAttr) otlpMetric {
	return otlpMetric{
		Name: name,
		Unit: unit,
		Gauge: &otlpPoints{
			DataPoints: []otlpPoint{{
				Value:      value,
				Time:       unixNano(time.Now()),
				Attributes: labels,
			}},
		},
	}
}

// counter is a running total for the whole run. Prometheus works out the
// rate itself, so we always send the total rather than what changed.
func (p *AlloyPrinter) counter(name, unit string, points ...otlpPoint) otlpMetric {
	now := unixNano(time.Now())
	start := unixNano(p.startTime)

	for i := range points {
		points[i].Time = now
		points[i].StartTime = start
	}

	return otlpMetric{
		Name: name,
		Unit: unit,
		Sum: &otlpSum{
			DataPoints:             points,
			AggregationTemporality: otlpCumulative,
			IsMonotonic:            true,
		},
	}
}

// send POSTs one batch of metrics to Alloy. A failure does not stop the
// probing, and we say so only once, so an Alloy that is down does not fill
// the terminal with the same error every second.
func (p *AlloyPrinter) send(metrics []otlpMetric) {
	if len(metrics) == 0 {
		return
	}

	payload := otlpPayload{
		ResourceMetrics: []otlpResourceMetrics{{
			Resource: otlpResource{
				Attributes: []otlpAttr{attr("service.name", "tcping")},
			},
			ScopeMetrics: []otlpScopeMetrics{{Metrics: metrics}},
		}},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		p.warnOnce("could not encode metrics: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		p.warnOnce("could not build the request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", version.UserAgent)

	resp, err := p.client.Do(req)
	if err != nil {
		p.warnOnce("could not reach Alloy at %s: %v", p.endpoint, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.warnOnce("Alloy rejected the metrics with %s", resp.Status)
	}
}

func (p *AlloyPrinter) warnOnce(format string, args ...any) {
	if p.warned {
		return
	}

	p.warned = true
	fmt.Fprintf(os.Stderr, "Alloy Error: "+format+"\n", args...)
	fmt.Fprintln(os.Stderr, "Probing continues, but the metrics are being dropped.")
}

// httpMetrics are the extra timings an HTTP(S) probe learned. They are the
// reason to graph an HTTP target at all: a slow connect and a slow first
// byte mean different things.
func (p *AlloyPrinter) httpMetrics(s *stats.Statistics) []otlpMetric {
	if !s.IsHTTP() || !s.HasHTTPResponse() {
		return nil
	}

	labels := p.labels(s)

	metrics := []otlpMetric{
		p.gauge("tcping_http_status_code", "", float64(s.HTTP.StatusCode), labels),
		p.gauge("tcping_http_connect_milliseconds", "ms", msOf(s.HTTP.ConnectDuration), labels),
		p.gauge("tcping_http_time_to_first_byte_milliseconds", "ms", msOf(s.HTTP.TimeToFirstByte), labels),
	}

	if s.HTTP.TLSVersion == "" {
		return metrics
	}

	return append(metrics,
		p.gauge("tcping_http_tls_handshake_milliseconds", "ms", msOf(s.HTTP.TLSDuration), labels),
		p.gauge("tcping_certificate_days_remaining", "d", float64(s.CertDaysRemaining()), labels),
	)
}

// udpMetrics are what a UDP probe learned. A reply that carried our own
// payload back is the only proof that something is really listening, and an
// ICMP refusal the only proof that nothing is, so both are worth a graph.
func (p *AlloyPrinter) udpMetrics(s *stats.Statistics) []otlpMetric {
	if !s.IsUDP() {
		return nil
	}

	labels := p.labels(s)

	return []otlpMetric{
		p.gauge("tcping_udp_reply_echoed", "", oneIf(s.UDP.Echoed), labels),
		p.gauge("tcping_udp_port_unreachable", "", oneIf(s.UDP.Rejected), labels),
		p.gauge("tcping_udp_reply_bytes", "By", float64(s.UDP.ReplySize), labels),
	}
}

// oneIf is how a yes or no is sent, since a metric can only carry a number.
func oneIf(b bool) float64 {
	if b {
		return 1
	}

	return 0
}

// probeMetrics is what every probe sends, successful or not. The gauge says
// what just happened and the counter says how the run is going, so a graph
// can show both the last probe and the trend.
func (p *AlloyPrinter) probeMetrics(s *stats.Statistics, succeeded bool) []otlpMetric {
	metrics := []otlpMetric{
		p.gauge("tcping_probe_success", "", oneIf(succeeded), p.labels(s)),
		// Always 1. The value means nothing, the labels are the point: this
		// is where you look up which address the target resolved to.
		p.gauge("tcping_target_address", "", 1, p.labels(s, attr("ip", s.IPStr()))),
		p.counter("tcping_probes_total", "",
			otlpPoint{
				Value:      float64(s.TotalSuccessfulProbes),
				Attributes: p.labels(s, attr("result", "success")),
			},
			otlpPoint{
				Value:      float64(s.TotalUnsuccessfulProbes),
				Attributes: p.labels(s, attr("result", "failure")),
			},
		),
	}

	// A failed probe has no round trip time to report.
	if succeeded {
		metrics = append(metrics, p.gauge("tcping_probe_rtt_milliseconds", "ms", rttOf(s.LatestRTT), p.labels(s)))
	}

	metrics = append(metrics, p.httpMetrics(s)...)

	return append(metrics, p.udpMetrics(s)...)
}

// msOf turns a duration into milliseconds, the unit the rest of tcping
// reports latencies in.
func msOf(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}

// rttOf widens an RTT for a data point, which can only carry a float64.
// Rounding to three decimals, the precision the console printers use, keeps
// the widening from turning 2.1 into 2.0999999046325684.
func rttOf(rtt float32) float64 {
	return math.Round(float64(rtt)*1000) / 1000
}

// PrintStart says where the metrics are going, then leaves the terminal
// alone for the rest of the run.
func (p *AlloyPrinter) PrintStart(s *stats.Statistics) {
	if s.DestIsIP {
		fmt.Printf("Probing %s on port %d over %s - sending metrics to: %s\n",
			s.Hostname, s.Port, s.ProtocolStr(), p.endpoint)
		return
	}

	fmt.Printf("Probing %s (%s) on port %d over %s (resolved in %s ms) - sending metrics to: %s\n",
		s.Hostname, s.IPStr(), s.Port, s.ProtocolStr(), s.NameResolutionDurationStr(), p.endpoint)
}

// PrintNameResolutionDuration sends how long the hostname resolution took.
func (p *AlloyPrinter) PrintNameResolutionDuration(s *stats.Statistics) {
	p.send([]otlpMetric{
		p.gauge("tcping_name_resolution_milliseconds", "ms", msOf(s.NameResolutionDuration), p.labels(s)),
	})
}

// PrintProbeSuccess sends the metrics of a successful probe.
func (p *AlloyPrinter) PrintProbeSuccess(s *stats.Statistics) {
	p.send(append(p.probeMetrics(s, true), p.dueStatistics(s)...))
}

// PrintProbeFailure sends the metrics of a failed probe.
func (p *AlloyPrinter) PrintProbeFailure(s *stats.Statistics) {
	p.send(append(p.probeMetrics(s, false), p.dueStatistics(s)...))
}

// dueStatistics returns the run summary when enough time has passed since
// the last one, and nothing otherwise. Sending it along with a probe means
// no goroutine and no second request, and it keeps the summary flowing on a
// run that is never going to be stopped by hand.
func (p *AlloyPrinter) dueStatistics(s *stats.Statistics) []otlpMetric {
	if time.Since(p.lastStats) < p.statsInterval {
		return nil
	}

	p.lastStats = time.Now()

	return p.statisticsMetrics(s)
}

// PrintStatistics sends the summary of the run so far.
func (p *AlloyPrinter) PrintStatistics(s *stats.Statistics) {
	p.lastStats = time.Now()
	p.send(p.statisticsMetrics(s))
}

// statisticsMetrics is the summary of the run so far: the numbers you would
// otherwise only see when tcping exits. Every line of the statistics block
// the terminal prints has a metric here.
//
// Times are sent as milliseconds since the epoch rather than as text,
// because a metric can only carry a number. Milliseconds rather than seconds
// because that is what Grafana's date units read, and what the rest of
// tcping's timings are already in.
func (p *AlloyPrinter) statisticsMetrics(s *stats.Statistics) []otlpMetric {
	labels := p.labels(s)

	metrics := []otlpMetric{
		p.gauge("tcping_packet_loss_percent", "%", float64(s.PacketLoss()), labels),
		p.gauge("tcping_start_time_milliseconds", "ms", float64(s.StartTime.UnixMilli()), labels),
		p.gauge("tcping_run_duration_seconds", "s", s.RuntimeSeconds(), labels),
		p.counter("tcping_uptime_seconds_total", "s",
			otlpPoint{Value: s.TotalUptime.Seconds(), Attributes: p.labels(s)},
		),
		p.counter("tcping_downtime_seconds_total", "s",
			otlpPoint{Value: s.TotalDowntime.Seconds(), Attributes: p.labels(s)},
		),
		p.counter("tcping_hostname_resolution_retries_total", "",
			otlpPoint{Value: float64(s.RetriedHostnameLookups), Attributes: p.labels(s)},
		),
		p.counter("tcping_hostname_changes_total", "",
			otlpPoint{Value: float64(s.HostnameChangeCount()), Attributes: p.labels(s)},
		),
	}

	// A probe that has never happened has no timestamp, and sending a zero
	// one would put it in the year 1.
	if !s.LastSuccessfulProbe.IsZero() {
		metrics = append(metrics,
			p.gauge("tcping_last_successful_probe_milliseconds", "ms", float64(s.LastSuccessfulProbe.UnixMilli()), labels),
		)
	}

	if !s.LastUnsuccessfulProbe.IsZero() {
		metrics = append(metrics,
			p.gauge("tcping_last_unsuccessful_probe_milliseconds", "ms", float64(s.LastUnsuccessfulProbe.UnixMilli()), labels),
		)
	}

	// A streak is only known once it has ended, so a target that has not
	// changed state yet has neither of these.
	if s.LongestUptime.Duration != 0 {
		metrics = append(metrics,
			p.gauge("tcping_longest_uptime_seconds", "s", s.LongestUptime.Duration.Seconds(), labels),
			p.gauge("tcping_longest_uptime_start_milliseconds", "ms", float64(s.LongestUptime.Start.UnixMilli()), labels),
			p.gauge("tcping_longest_uptime_end_milliseconds", "ms", float64(s.LongestUptime.End.UnixMilli()), labels),
		)
	}

	if s.LongestDowntime.Duration != 0 {
		metrics = append(metrics,
			p.gauge("tcping_longest_downtime_seconds", "s", s.LongestDowntime.Duration.Seconds(), labels),
			p.gauge("tcping_longest_downtime_start_milliseconds", "ms", float64(s.LongestDowntime.Start.UnixMilli()), labels),
			p.gauge("tcping_longest_downtime_end_milliseconds", "ms", float64(s.LongestDowntime.End.UnixMilli()), labels),
		)
	}

	// Only the last summary of a run has an end time.
	if !s.EndTime.IsZero() {
		metrics = append(metrics,
			p.gauge("tcping_end_time_milliseconds", "ms", float64(s.EndTime.UnixMilli()), labels),
		)
	}

	// Without a single successful probe there is no latency to summarize,
	// and sending zeros would look like a very fast target.
	if s.TotalSuccessfulProbes > 0 {
		metrics = append(metrics,
			p.gauge("tcping_rtt_milliseconds", "ms", rttOf(s.RTTResults.Min), p.labels(s, attr("stat", "min"))),
			p.gauge("tcping_rtt_milliseconds", "ms", rttOf(s.RTTResults.Average), p.labels(s, attr("stat", "avg"))),
			p.gauge("tcping_rtt_milliseconds", "ms", rttOf(s.RTTResults.Max), p.labels(s, attr("stat", "max"))),
			p.gauge("tcping_rtt_milliseconds", "ms", rttOf(s.RTTResults.Mdev), p.labels(s, attr("stat", "mdev"))),
		)
	}

	return metrics
}

// PrintRetryingToResolve has no number behind it, so it goes to the terminal.
func (p *AlloyPrinter) PrintRetryingToResolve(hostname string) {
	fmt.Fprintf(os.Stderr, "retrying to resolve %s\n", hostname)
}

// PrintDownTimeDuration sends how long the outage that just ended lasted.
func (p *AlloyPrinter) PrintDownTimeDuration(s *stats.Statistics) {
	p.send([]otlpMetric{
		p.gauge("tcping_last_downtime_seconds", "s", s.CurrentDowntime.Seconds(), p.labels(s)),
	})
}

// PrintUpTimeDuration sends how long the target was up for, right as it
// stops responding.
func (p *AlloyPrinter) PrintUpTimeDuration(s *stats.Statistics) {
	p.send([]otlpMetric{
		p.gauge("tcping_last_uptime_seconds", "s", s.CurrentUptime.Seconds(), p.labels(s)),
	})
}

func (p *AlloyPrinter) PrintError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Alloy Error: "+format+"\n", args...)
}

// Shutdown sends the final statistics and closes the connection to Alloy.
// Statistics are already finalized by finalizeStatistics by the time this
// runs. It does not exit the program - that decision belongs to the caller,
// not the printer.
func (p *AlloyPrinter) Shutdown(s *stats.Statistics) {
	p.PrintStatistics(s)
	p.client.CloseIdleConnections()
}
