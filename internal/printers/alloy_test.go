package printers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/consts"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
)

func TestAlloyEndpoint(t *testing.T) {
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
		p := NewAlloyPrinter(config.PrinterConfig{AlloyURL: tt.given})

		if p.endpoint != tt.want {
			t.Errorf("NewAlloyPrinter(%q) endpoint = %q, want %q", tt.given, p.endpoint, tt.want)
		}
	}
}

// alloyTestStats is a probe that succeeded in 3.5ms, which is enough to
// check what a successful probe sends.
func alloyTestStats() *stats.Statistics {
	return &stats.Statistics{
		Hostname:              "example.com",
		IP:                    netip.MustParseAddr("93.184.216.34"),
		Port:                  443,
		Protocol:              consts.TCP,
		LatestRTT:             3.5,
		TotalSuccessfulProbes: 2,
		StartTime:             time.Now(),
	}
}

func TestAlloyPrintProbeSuccess(t *testing.T) {
	var got otlpPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		if err := json.Unmarshal(body, &got); err != nil {
			t.Errorf("Alloy received invalid JSON: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	printer := NewAlloyPrinter(config.PrinterConfig{AlloyURL: server.URL})
	printer.PrintProbeSuccess(alloyTestStats())

	if len(got.ResourceMetrics) != 1 {
		t.Fatalf("expected 1 resourceMetrics, got %d", len(got.ResourceMetrics))
	}

	metrics := got.ResourceMetrics[0].ScopeMetrics[0].Metrics

	found := make(map[string]float64)
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
func TestAlloyPrintProbeFailureHasNoRTT(t *testing.T) {
	var got otlpPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	printer := NewAlloyPrinter(config.PrinterConfig{AlloyURL: server.URL})
	printer.PrintProbeFailure(alloyTestStats())

	for _, m := range got.ResourceMetrics[0].ScopeMetrics[0].Metrics {
		if m.Name == "tcping_probe_rtt_milliseconds" {
			t.Error("a failed probe should not send an RTT")
		}
	}
}

// A UDP probe cannot say much, so the little it does learn has to be sent:
// whether the reply was our own payload coming back and whether the port
// refused us.
func TestAlloyUDPProbeSendsWhatItLearned(t *testing.T) {
	var got otlpPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	printer := NewAlloyPrinter(config.PrinterConfig{AlloyURL: server.URL})

	s := alloyTestStats()
	s.Protocol = consts.UDP
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

// An Alloy that is not there must not stop the probing, and must not repeat
// the same complaint on every probe.
func TestAlloyKeepsGoingWhenUnreachable(t *testing.T) {
	printer := NewAlloyPrinter(config.PrinterConfig{AlloyURL: "http://127.0.0.1:1"})

	printer.PrintProbeSuccess(alloyTestStats())

	if !printer.warned {
		t.Error("expected the printer to warn about an unreachable Alloy")
	}

	printer.PrintProbeSuccess(alloyTestStats())
}

// The run summary has to keep flowing on its own, otherwise a tcping that
// nobody stops never reports one.
func TestAlloyStatisticsRideAlongWithProbes(t *testing.T) {
	var payloads []otlpPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var got otlpPayload

		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &got)

		payloads = append(payloads, got)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	printer := NewAlloyPrinter(config.PrinterConfig{
		AlloyURL:           server.URL,
		AlloyStatsInterval: 10 * time.Second,
	})

	hasSummary := func(p otlpPayload) bool {
		for _, m := range p.ResourceMetrics[0].ScopeMetrics[0].Metrics {
			if m.Name == "tcping_packet_loss_percent" {
				return true
			}
		}
		return false
	}

	// The first probe carries a summary, so the metrics show up right away.
	printer.PrintProbeSuccess(alloyTestStats())
	if !hasSummary(payloads[0]) {
		t.Error("the first probe should carry the run summary")
	}

	// The next one does not, since the interval has not passed.
	printer.PrintProbeSuccess(alloyTestStats())
	if hasSummary(payloads[1]) {
		t.Error("the summary should not be sent with every probe")
	}

	// Once it has, it comes along again.
	printer.lastStats = time.Now().Add(-printer.statsInterval)
	printer.PrintProbeSuccess(alloyTestStats())
	if !hasSummary(payloads[2]) {
		t.Error("the summary should be sent again once the interval passed")
	}
}
