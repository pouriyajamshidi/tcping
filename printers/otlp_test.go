package printers

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"slices"
	"strings"
	"sync"
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
	printer.queue.flush()

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
				printer.queue.flush()
			})

			if warnings := strings.Count(stderr, "OTLP Error:"); warnings != tt.warnings {
				t.Errorf("complained %d times, want %d:\n%s", warnings, tt.warnings, stderr)
			}
		})
	}
}

// Sends go out every few seconds, so every one of them opening a new
// connection would be a lot of wasted handshakes on a long run.
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
		printer.queue.flush()
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
	printer.queue.flush()

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
	printer.queue.flush()

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
	printer.queue.flush()

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

// The download time and the page size are only sent with --full-page, and
// an empty page is still a page, so it is sent as 0.
func TestOTLPHTTPFullPageSendsTheDownload(t *testing.T) {
	tests := []struct {
		name     string
		bodyRead bool
		size     int64
	}{
		{"without --full-page", false, 0},
		{"with --full-page", true, 48213},
		{"with --full-page and an empty page", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			s.Protocol = config.HTTP
			s.HTTP.StatusCode = 200
			s.HTTP.BodyRead = tt.bodyRead
			s.HTTP.DownloadDuration = 30 * time.Millisecond
			s.HTTP.BodySize = tt.size

			printer.PrintProbeSuccess(s)
			printer.queue.flush()

			found := map[string]float64{}
			for _, m := range got.ResourceMetrics[0].ScopeMetrics[0].Metrics {
				if m.Gauge != nil {
					found[m.Name] = m.Gauge.DataPoints[0].Value
				}
			}

			download, hasDownload := found["tcping_http_download_milliseconds"]
			pageBytes, hasPageBytes := found["tcping_http_page_bytes"]

			if !tt.bodyRead {
				if hasDownload || hasPageBytes {
					t.Errorf("sent the download time or page size without --full-page: %v", found)
				}
				return
			}

			if !hasDownload || download != 30 {
				t.Errorf("tcping_http_download_milliseconds = %v (sent: %v), want 30", download, hasDownload)
			}
			if !hasPageBytes || pageBytes != float64(tt.size) {
				t.Errorf("tcping_http_page_bytes = %v (sent: %v), want %d", pageBytes, hasPageBytes, tt.size)
			}
		})
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
			printer.queue.flush()
		}
	})

	if errors := strings.Count(stderr, "OTLP Error:"); errors != 3 {
		t.Errorf("3 failed sends printed %d errors, want 3:\n%s", errors, stderr)
	}
}

// A slow endpoint is what used to slow the probing down: every probe waited
// for its own send to be answered. Now a probe only puts its metrics in
// line, so it is done even while a send is still waiting for an answer.
func TestOTLPProbesDoNotWaitOnASlowSend(t *testing.T) {
	var (
		mu       sync.Mutex
		payloads []otlpPayload
	)

	entered := make(chan struct{}, 1)
	release := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case entered <- struct{}{}:
		default:
		}
		<-release

		var got otlpPayload
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &got); err != nil {
			t.Errorf("the OTLP endpoint received invalid JSON: %v", err)
		}

		mu.Lock()
		payloads = append(payloads, got)
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	printer, err := NewOTLPPrinter(Config{OTLPURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}

	printer.PrintProbeSuccess(otlpTestStats())
	go printer.queue.flush()
	<-entered

	probed := make(chan struct{})
	go func() {
		printer.PrintProbeSuccess(otlpTestStats())
		close(probed)
	}()

	select {
	case <-probed:
	case <-time.After(time.Second):
		t.Error("a probe waited on a send that the endpoint had not answered yet")
	}

	close(release)
	<-probed
	printer.Shutdown(otlpTestStats())

	mu.Lock()
	defer mu.Unlock()

	if got := otlpProbeResults(payloads...); len(got) != 2 {
		t.Errorf("the endpoint got %d probes, want 2", len(got))
	}
}

// otlpProbeResults is the tcping_probe_success value of every probe in the
// payloads, in the order they were sent.
func otlpProbeResults(payloads ...otlpPayload) []float64 {
	var results []float64

	for _, p := range payloads {
		for _, m := range p.ResourceMetrics[0].ScopeMetrics[0].Metrics {
			if m.Name == "tcping_probe_success" {
				results = append(results, m.Gauge.DataPoints[0].Value)
			}
		}
	}

	return results
}

// otlpStatusServer answers every send with the next status in the list, and
// with 200 once the list runs out. It records the payloads it accepted.
func otlpStatusServer(t *testing.T, statuses []int, payloads *[]otlpPayload) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var got otlpPayload
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &got); err != nil {
			t.Errorf("the OTLP endpoint received invalid JSON: %v", err)
		}

		status := http.StatusOK
		if len(statuses) > 0 {
			status, statuses = statuses[0], statuses[1:]
		}

		if status/100 == 2 {
			*payloads = append(*payloads, got)
		}

		w.WriteHeader(status)
	}))
}

// An endpoint that is restarting, or too busy to answer, gets the metrics it
// missed once it is back, so the graph has no hole in it.
func TestOTLPSendsWhatItMissedOnceBack(t *testing.T) {
	var payloads []otlpPayload

	server := otlpStatusServer(t, []int{http.StatusServiceUnavailable}, &payloads)
	defer server.Close()

	printer, err := NewOTLPPrinter(Config{OTLPURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}

	stderr := captureStderr(t, func() {
		printer.PrintProbeFailure(otlpTestStats())
		printer.queue.flush()
	})

	if !strings.Contains(stderr, "OTLP Error: the OTLP endpoint rejected the metrics with 503 Service Unavailable") {
		t.Errorf("the failed send was not reported:\n%s", stderr)
	}

	printer.PrintProbeSuccess(otlpTestStats())
	printer.queue.flush()

	if len(payloads) != 1 {
		t.Fatalf("got %d sends, want 1", len(payloads))
	}

	if got, want := otlpProbeResults(payloads...), []float64{0, 1}; !slices.Equal(got, want) {
		t.Errorf("the send should hold the missed probe and then the new one, got %v, want %v", got, want)
	}
}

// Metrics the endpoint rejected for what they hold would be rejected again,
// so they are reported and not sent a second time.
func TestOTLPDoesNotRetryARejectedSend(t *testing.T) {
	var payloads []otlpPayload

	server := otlpStatusServer(t, []int{http.StatusBadRequest}, &payloads)
	defer server.Close()

	printer, err := NewOTLPPrinter(Config{OTLPURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}

	captureStderr(t, func() {
		printer.PrintProbeFailure(otlpTestStats())
		printer.queue.flush()
	})

	printer.PrintProbeSuccess(otlpTestStats())
	printer.queue.flush()

	if got, want := otlpProbeResults(payloads...), []float64{1}; !slices.Equal(got, want) {
		t.Errorf("only the second probe should have been sent, got %v, want %v", got, want)
	}
}

// Metrics wait a few seconds before they are sent, so the last probes of a
// run are still waiting when it ends, and Shutdown has to send them along
// with the final summary.
func TestOTLPShutdownSendsWhatIsWaiting(t *testing.T) {
	var payloads []otlpPayload

	server := otlpStatusServer(t, nil, &payloads)
	defer server.Close()

	printer, err := NewOTLPPrinter(Config{OTLPURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}

	printer.PrintProbeSuccess(otlpTestStats())
	printer.PrintProbeFailure(otlpTestStats())
	printer.Shutdown(otlpTestStats())

	if len(payloads) != 1 {
		t.Fatalf("got %d sends, want 1", len(payloads))
	}

	if got, want := otlpProbeResults(payloads...), []float64{1, 0}; !slices.Equal(got, want) {
		t.Errorf("sent the probes %v, want %v", got, want)
	}

	var summaries int
	for _, m := range payloads[0].ResourceMetrics[0].ScopeMetrics[0].Metrics {
		if m.Name == "tcping_packet_loss_percent" {
			summaries++
		}
	}

	// The first probe carries a summary of its own, see
	// TestOTLPStatisticsRideAlongWithProbes.
	if summaries != 2 {
		t.Errorf("sent %d summaries, want 2", summaries)
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
	printer.queue.flush()
	if !hasSummary(payloads[0]) {
		t.Error("the first probe should carry the run summary")
	}

	// The next one does not, since the interval has not passed.
	printer.PrintProbeSuccess(otlpTestStats())
	printer.queue.flush()
	if hasSummary(payloads[1]) {
		t.Error("the summary should not be sent with every probe")
	}

	// Once it has, it comes along again.
	printer.lastStats = time.Now().Add(-printer.statsInterval)
	printer.PrintProbeSuccess(otlpTestStats())
	printer.queue.flush()
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
	printer.queue.flush()

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
	printer.queue.flush()

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
	printer.queue.flush()

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
	printer.queue.flush()

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
