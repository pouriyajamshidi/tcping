package printers

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/config"

	"github.com/pouriyajamshidi/tcping/v3/stats"
)

func TestOTLPEndpoint(t *testing.T) {
	tests := []struct {
		given string
		want  string
	}{
		{"http://localhost:4318", "http://localhost:4318/v1/metrics"},
		{"http://localhost:4318/", "http://localhost:4318/v1/metrics"},
		{"http://localhost:4318/v1/metrics", "http://localhost:4318/v1/metrics"},
		{"localhost:4318", "http://localhost:4318/v1/metrics"},
	}

	for _, tt := range tests {
		p, err := NewOTLPPrinter(Config{OTLPURL: tt.given})
		if err != nil {
			t.Fatalf("NewOTLPPrinter(%q) returned %v", tt.given, err)
		}

		if p.endpoint != tt.want {
			t.Errorf("NewOTLPPrinter(%q) endpoint = %q, want %q", tt.given, p.endpoint, tt.want)
		}
	}
}

// Hosted backends each want their token in a header of their own choosing,
// so the header given has to arrive as it was written.
func TestOTLPSendsTheGivenHeader(t *testing.T) {
	var got string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Honeycomb-Team")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	printer, err := NewOTLPPrinter(Config{OTLPURL: server.URL, OTLPHeader: "X-Honeycomb-Team: secret: with colon"})
	if err != nil {
		t.Fatal(err)
	}
	printer.PrintProbeSuccess(otlpTestStats())

	if got != "secret: with colon" {
		t.Errorf("X-Honeycomb-Team = %q, want %q", got, "secret: with colon")
	}
}

// Any 2xx means the metrics got through. Collectors and hosted backends do not
// all answer with 200, and a 202 or 204 must not be reported as a rejection.
func TestOTLPAcceptsAny2xx(t *testing.T) {
	tests := []struct {
		status   int
		warnings int
	}{
		{http.StatusOK, 0},
		{http.StatusAccepted, 0},
		{http.StatusNoContent, 0},
		{http.StatusBadRequest, 1},
		{http.StatusInternalServerError, 1},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer server.Close()

			printer, err := NewOTLPPrinter(Config{OTLPURL: server.URL})
			if err != nil {
				t.Fatal(err)
			}

			stderr := captureStderr(t, func() {
				printer.PrintProbeSuccess(otlpTestStats())
			})

			if warnings := strings.Count(stderr, "OTLP Error:"); warnings != tt.warnings {
				t.Errorf("complained %d times, want %d:\n%s", warnings, tt.warnings, stderr)
			}
		})
	}
}

// Probes are a second apart, so every one of them opening a new connection
// would be a lot of wasted handshakes on a long run.
func TestOTLPReusesTheConnection(t *testing.T) {
	var connections atomic.Int32

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"partialSuccess":{}}`))
	}))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			connections.Add(1)
		}
	}
	server.Start()
	defer server.Close()

	printer, err := NewOTLPPrinter(Config{OTLPURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}

	for range 3 {
		printer.PrintProbeSuccess(otlpTestStats())
	}

	if got := connections.Load(); got != 1 {
		t.Errorf("3 probes opened %d connections, want 1", got)
	}
}

func TestOTLPRejectsABadHeader(t *testing.T) {
	for _, header := range []string{"no colon here", ": no name"} {
		if _, err := NewOTLPPrinter(Config{OTLPURL: "http://localhost:4318", OTLPHeader: header}); err == nil {
			t.Errorf("expected an error for the header %q", header)
		}
	}
}

// otlpTestStats is a probe that succeeded in 3.5ms, which is enough to
// check what a successful probe sends.
func otlpTestStats() *stats.Statistics {
	return &stats.Statistics{
		Hostname:              "example.com",
		IP:                    netip.MustParseAddr("93.184.216.34"),
		Port:                  443,
		Protocol:              config.TCP,
		LatestRTT:             3.5,
		TotalSuccessfulProbes: 2,
		StartTime:             time.Now(),
	}
}

func TestOTLPPrintProbeSuccess(t *testing.T) {
	var got otlpPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		if err := json.Unmarshal(body, &got); err != nil {
			t.Errorf("the OTLP endpoint received invalid JSON: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	printer, err := NewOTLPPrinter(Config{OTLPURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	printer.PrintProbeSuccess(otlpTestStats())

	if len(got.ResourceMetrics) != 1 {
		t.Fatalf("expected 1 resourceMetrics, got %d", len(got.ResourceMetrics))
	}

	metrics := got.ResourceMetrics[0].ScopeMetrics[0].Metrics

	found := map[string]float64{}
	for _, m := range metrics {
		if m.Gauge != nil {
			found[m.Name] = m.Gauge.DataPoints[0].Value
		}
	}

	if found["tcping_probe_success"] != 1 {
		t.Errorf("tcping_probe_success = %v, want 1", found["tcping_probe_success"])
	}

	if found["tcping_probe_rtt_milliseconds"] != 3.5 {
		t.Errorf("tcping_probe_rtt_milliseconds = %v, want 3.5", found["tcping_probe_rtt_milliseconds"])
	}
}

// A failed probe has no round trip time, so sending one would be a lie.
func TestOTLPPrintProbeFailureHasNoRTT(t *testing.T) {
	var got otlpPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &got); err != nil {
			t.Errorf("the OTLP endpoint received invalid JSON: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	printer, err := NewOTLPPrinter(Config{OTLPURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	printer.PrintProbeFailure(otlpTestStats())

	for _, m := range got.ResourceMetrics[0].ScopeMetrics[0].Metrics {
		if m.Name == "tcping_probe_rtt_milliseconds" {
			t.Error("a failed probe should not send an RTT")
		}
	}
}

// A UDP probe cannot say much, so the little it does learn has to be sent:
// whether the reply was our own payload coming back and whether the port
// refused us.
func TestOTLPUDPProbeSendsWhatItLearned(t *testing.T) {
	var got otlpPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &got); err != nil {
			t.Errorf("the OTLP endpoint received invalid JSON: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	printer, err := NewOTLPPrinter(Config{OTLPURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}

	s := otlpTestStats()
	s.Protocol = config.UDP
	s.UDP.Echoed = true
	s.UDP.ReplySize = 4

	printer.PrintProbeSuccess(s)

	found := map[string]float64{}
	for _, m := range got.ResourceMetrics[0].ScopeMetrics[0].Metrics {
		if m.Gauge != nil {
			found[m.Name] = m.Gauge.DataPoints[0].Value
		}
	}

	if found["tcping_udp_reply_echoed"] != 1 {
		t.Errorf("tcping_udp_reply_echoed = %v, want 1", found["tcping_udp_reply_echoed"])
	}

	if found["tcping_udp_port_unreachable"] != 0 {
		t.Errorf("tcping_udp_port_unreachable = %v, want 0", found["tcping_udp_port_unreachable"])
	}

	if found["tcping_udp_reply_bytes"] != 4 {
		t.Errorf("tcping_udp_reply_bytes = %v, want 4", found["tcping_udp_reply_bytes"])
	}
}

// An OTLP endpoint that is not there must not stop the probing, and every
// failed send has to be reported, so an outage that lasts is never hidden.
func TestOTLPKeepsGoingWhenUnreachable(t *testing.T) {
	printer, err := NewOTLPPrinter(Config{OTLPURL: "http://127.0.0.1:1"})
	if err != nil {
		t.Fatal(err)
	}

	stderr := captureStderr(t, func() {
		for range 3 {
			printer.PrintProbeSuccess(otlpTestStats())
		}
	})

	if errors := strings.Count(stderr, "OTLP Error:"); errors != 3 {
		t.Errorf("3 failed sends printed %d errors, want 3:\n%s", errors, stderr)
	}
}

// The run summary has to keep flowing on its own, otherwise a tcping that
// nobody stops never reports one.
func TestOTLPStatisticsRideAlongWithProbes(t *testing.T) {
	var payloads []otlpPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var got otlpPayload

		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &got); err != nil {
			t.Errorf("the OTLP endpoint received invalid JSON: %v", err)
		}

		payloads = append(payloads, got)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	printer, err := NewOTLPPrinter(Config{
		OTLPURL:       server.URL,
		StatsInterval: 10 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}

	hasSummary := func(p otlpPayload) bool {
		for _, m := range p.ResourceMetrics[0].ScopeMetrics[0].Metrics {
			if m.Name == "tcping_packet_loss_percent" {
				return true
			}
		}
		return false
	}

	// The first probe carries a summary, so the metrics show up right away.
	printer.PrintProbeSuccess(otlpTestStats())
	if !hasSummary(payloads[0]) {
		t.Error("the first probe should carry the run summary")
	}

	// The next one does not, since the interval has not passed.
	printer.PrintProbeSuccess(otlpTestStats())
	if hasSummary(payloads[1]) {
		t.Error("the summary should not be sent with every probe")
	}

	// Once it has, it comes along again.
	printer.lastStats = time.Now().Add(-printer.statsInterval)
	printer.PrintProbeSuccess(otlpTestStats())
	if !hasSummary(payloads[2]) {
		t.Error("the summary should be sent again once the interval passed")
	}
}

// The OTLP printer prints nothing after this line, so it is the only place
// the user gets to see which address is being probed.
func TestOTLPPrintStartShowsTheIP(t *testing.T) {
	p, err := NewOTLPPrinter(Config{OTLPURL: "http://localhost:4318"})
	if err != nil {
		t.Fatal(err)
	}

	s := otlpTestStats()
	s.NameResolutionDuration = 12 * time.Millisecond

	t.Run("hostname target", func(t *testing.T) {
		out := captureStdout(t, func() { p.PrintStart(s) })

		want := "Probing example.com (93.184.216.34) on port 443 over TCP (resolved in 12.000 ms) - sending metrics to: " + p.endpoint + "\n"
		if out != want {
			t.Errorf("output = %q, want %q", out, want)
		}
	})

	t.Run("IP target", func(t *testing.T) {
		ipTarget := otlpTestStats()
		ipTarget.Hostname = "93.184.216.34"
		ipTarget.DestIsIP = true

		out := captureStdout(t, func() { p.PrintStart(ipTarget) })

		want := "Probing 93.184.216.34 on port 443 over TCP - sending metrics to: " + p.endpoint + "\n"
		if out != want {
			t.Errorf("output = %q, want %q", out, want)
		}
	})
}

// Several machines can send to the same OTLP endpoint, so every data point
// has to say which one it came from. It has to be on the point itself, not on the
// resource, or Prometheus would not have it as a label.
func TestOTLPSourceLabelIsOnEveryDataPoint(t *testing.T) {
	var got otlpPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &got); err != nil {
			t.Errorf("the OTLP endpoint received invalid JSON: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	printer, err := NewOTLPPrinter(Config{OTLPURL: server.URL, SourceLabel: "probe-1"})
	if err != nil {
		t.Fatal(err)
	}
	printer.PrintProbeSuccess(otlpTestStats())

	for _, m := range got.ResourceMetrics[0].ScopeMetrics[0].Metrics {
		points := []otlpPoint{}
		if m.Gauge != nil {
			points = m.Gauge.DataPoints
		}
		if m.Sum != nil {
			points = m.Sum.DataPoints
		}

		for _, point := range points {
			var source string
			for _, a := range point.Attributes {
				if a.Key == "source" {
					source = a.Value.String
				}
			}

			if source != "probe-1" {
				t.Errorf("%s has source %q, want %q", m.Name, source, "probe-1")
			}
		}
	}
}

// The resolved IP has to stay off the probe metrics. As a label it would
// identify the series, so a hostname that resolves somewhere else mid-run
// would leave the old series behind and start a new one. It gets a metric
// of its own instead.
func TestOTLPResolvedIPHasItsOwnMetric(t *testing.T) {
	var got otlpPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &got); err != nil {
			t.Errorf("the OTLP endpoint received invalid JSON: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	printer, err := NewOTLPPrinter(Config{OTLPURL: server.URL, SourceLabel: "probe-1"})
	if err != nil {
		t.Fatal(err)
	}
	printer.PrintProbeSuccess(otlpTestStats())

	var address *otlpMetric

	for i, m := range got.ResourceMetrics[0].ScopeMetrics[0].Metrics {
		points := []otlpPoint{}
		if m.Gauge != nil {
			points = m.Gauge.DataPoints
		}
		if m.Sum != nil {
			points = m.Sum.DataPoints
		}

		for _, point := range points {
			for _, a := range point.Attributes {
				if a.Key == "ip" && m.Name != "tcping_target_address" {
					t.Errorf("%s carries the IP as a label", m.Name)
				}
			}
		}

		if m.Name == "tcping_target_address" {
			address = &got.ResourceMetrics[0].ScopeMetrics[0].Metrics[i]
		}
	}

	if address == nil {
		t.Fatal("no tcping_target_address metric was sent")
	}

	if address.Gauge.DataPoints[0].Value != 1 {
		t.Errorf("tcping_target_address = %v, want 1", address.Gauge.DataPoints[0].Value)
	}
}

// Everything the statistics block prints in the terminal has to be in the
// summary too, otherwise a graph cannot show it.
func TestOTLPStatisticsCarryTheWholeSummary(t *testing.T) {
	var got otlpPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &got); err != nil {
			t.Errorf("the OTLP endpoint received invalid JSON: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	printer, err := NewOTLPPrinter(Config{OTLPURL: server.URL, SourceLabel: "probe-1"})
	if err != nil {
		t.Fatal(err)
	}
	printer.PrintStatistics(statisticsTestStats())

	values := map[string]float64{}

	for _, m := range got.ResourceMetrics[0].ScopeMetrics[0].Metrics {
		if m.Gauge != nil {
			values[m.Name] = m.Gauge.DataPoints[0].Value
		}
		if m.Sum != nil {
			values[m.Name] = m.Sum.DataPoints[0].Value
		}
	}

	want := map[string]float64{
		"tcping_packet_loss_percent":                  2.5,
		"tcping_start_time_milliseconds":              statisticsTestEpoch,
		"tcping_run_duration_seconds":                 60,
		"tcping_outages_total":                        1,
		"tcping_hostname_resolution_retries_total":    3,
		"tcping_hostname_changes_total":               1,
		"tcping_last_successful_probe_milliseconds":   statisticsTestEpoch + 40_000,
		"tcping_last_unsuccessful_probe_milliseconds": statisticsTestEpoch + 20_000,
		"tcping_longest_uptime_seconds":               20,
		"tcping_longest_uptime_start_milliseconds":    statisticsTestEpoch,
		"tcping_longest_uptime_end_milliseconds":      statisticsTestEpoch + 20_000,
		"tcping_longest_downtime_seconds":             5,
		"tcping_longest_downtime_start_milliseconds":  statisticsTestEpoch + 20_000,
		"tcping_longest_downtime_end_milliseconds":    statisticsTestEpoch + 25_000,
		"tcping_end_time_milliseconds":                statisticsTestEpoch + 60_000,
	}

	for name, w := range want {
		if values[name] != w {
			t.Errorf("%s = %v, want %v", name, values[name], w)
		}
	}
}

// A run that has not gone down yet has no streaks and no failed probe, and
// sending zeros for those would claim things that never happened.
func TestOTLPStatisticsOmitWhatHasNotHappened(t *testing.T) {
	var got otlpPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &got); err != nil {
			t.Errorf("the OTLP endpoint received invalid JSON: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	printer, err := NewOTLPPrinter(Config{OTLPURL: server.URL, SourceLabel: "probe-1"})
	if err != nil {
		t.Fatal(err)
	}
	printer.PrintStatistics(otlpTestStats())

	for _, m := range got.ResourceMetrics[0].ScopeMetrics[0].Metrics {
		switch m.Name {
		case "tcping_last_unsuccessful_probe_milliseconds",
			"tcping_longest_uptime_seconds",
			"tcping_longest_downtime_seconds",
			"tcping_end_time_milliseconds":
			t.Errorf("the summary should not carry %s yet", m.Name)

		default:
			// every other metric is expected in the summary
		}
	}
}
