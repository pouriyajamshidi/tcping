package printers

import (
	"encoding/json"
	"io"
	"net"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/config"

	"github.com/pouriyajamshidi/tcping/v3/stats"
)

// jsonTestStats is a probe against a hostname that succeeded in 3.5 ms.
// Tests change only the fields they care about.
func jsonTestStats() *stats.Statistics {
	return &stats.Statistics{
		Hostname:                  "example.com",
		IP:                        netip.MustParseAddr("93.184.216.34"),
		Port:                      443,
		Protocol:                  config.TCP,
		LatestRTT:                 3.5,
		OngoingSuccessfulProbes:   2,
		OngoingUnsuccessfulProbes: 5,
		NameResolutionDuration:    12 * time.Millisecond,
		StartTime:                 time.Now(),
	}
}

// jsonEvents runs f and returns everything the printer wrote, decoded. The
// encoder is bound to stdout when the printer is made, so the printer has
// to be made inside the capture too.
func jsonEvents(t *testing.T, cfg Config, f func(p *JSONPrinter)) []map[string]any {
	t.Helper()

	out := captureStdout(t, func() {
		f(NewJSONPrinter(cfg))
	})

	var events []map[string]any

	decoder := json.NewDecoder(strings.NewReader(out))
	for {
		var event map[string]any

		err := decoder.Decode(&event)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("output is not valid JSON: %v, got %q", err, out)
		}

		events = append(events, event)
	}

	return events
}

// printOneJSONEvent runs f, which must print exactly one event, and returns its
// type and its data fields.
func printOneJSONEvent(t *testing.T, cfg Config, f func(p *JSONPrinter)) (string, map[string]any) {
	t.Helper()

	events := jsonEvents(t, cfg, f)
	if len(events) != 1 {
		t.Fatalf("printed %d events, want 1", len(events))
	}

	eventType, _ := events[0]["type"].(string)

	data, ok := events[0]["data"].(map[string]any)
	if !ok {
		data = map[string]any{}
	}

	return eventType, data
}

// wantFields checks the fields a test names and leaves the rest alone.
func wantFields(t *testing.T, data map[string]any, want map[string]any) {
	t.Helper()

	for key, wantValue := range want {
		got, ok := data[key]
		if !ok {
			t.Errorf("field %q is missing, want %v", key, wantValue)
			continue
		}
		if got != wantValue {
			t.Errorf("field %q = %v, want %v", key, got, wantValue)
		}
	}
}

// wantNoFields checks that the fields a test names were left out entirely,
// which is how the JSON output stays quiet about things that do not apply.
func wantNoFields(t *testing.T, data map[string]any, keys ...string) {
	t.Helper()

	for _, key := range keys {
		if got, ok := data[key]; ok {
			t.Errorf("field %q = %v, want it to be left out", key, got)
		}
	}
}

func TestJSONPrintStart(t *testing.T) {
	t.Run("hostname target reports its resolution time", func(t *testing.T) {
		eventType, data := printOneJSONEvent(t, Config{}, func(p *JSONPrinter) {
			p.PrintStart(jsonTestStats())
		})

		if eventType != "start" {
			t.Errorf("event type = %q, want %q", eventType, "start")
		}

		wantFields(t, data, map[string]any{
			"hostname":                 "example.com",
			"port":                     float64(443),
			"protocol":                 "TCP",
			"nameResolutionDurationMs": "12.000",
		})
	})

	t.Run("IP target has no resolution time", func(t *testing.T) {
		s := jsonTestStats()
		s.DestIsIP = true

		_, data := printOneJSONEvent(t, Config{}, func(p *JSONPrinter) {
			p.PrintStart(s)
		})

		wantNoFields(t, data, "nameResolutionDurationMs")
	})
}

func TestJSONPrintProbeSuccess(t *testing.T) {
	eventType, data := printOneJSONEvent(t, Config{}, func(p *JSONPrinter) {
		p.PrintProbeSuccess(jsonTestStats())
	})

	if eventType != "probe" {
		t.Errorf("event type = %q, want %q", eventType, "probe")
	}

	wantFields(t, data, map[string]any{
		"hostname":    "example.com",
		"ipAddress":   "93.184.216.34",
		"port":        float64(443),
		"success":     true,
		"latency":     3.5,
		"connections": float64(2),
	})

	wantNoFields(t, data, "timestamp", "sourceAddress", "http", "udp")
}

func TestJSONPrintProbeFailure(t *testing.T) {
	_, data := printOneJSONEvent(t, Config{}, func(p *JSONPrinter) {
		p.PrintProbeFailure(jsonTestStats())
	})

	wantFields(t, data, map[string]any{
		"success":     false,
		"connections": float64(5),
	})

	// A probe that failed has no round trip to report.
	wantNoFields(t, data, "latency")
}

// The IP is already in its own field, so repeating it as the hostname would
// only be noise.
func TestJSONProbeLeavesHostnameOutForIPTarget(t *testing.T) {
	s := jsonTestStats()
	s.DestIsIP = true

	_, data := printOneJSONEvent(t, Config{}, func(p *JSONPrinter) {
		p.PrintProbeSuccess(s)
	})

	wantNoFields(t, data, "hostname")
}

func TestJSONProbeTimestampAndSourceAddress(t *testing.T) {
	s := jsonTestStats()
	s.LocalAddr = &net.TCPAddr{IP: net.ParseIP("192.168.1.10")}

	cfg := Config{WithTimestamp: true, WithSourceAddress: true}

	for _, tt := range []struct {
		name  string
		print func(p *JSONPrinter)
	}{
		{"success", func(p *JSONPrinter) { p.PrintProbeSuccess(s) }},
		{"failure", func(p *JSONPrinter) { p.PrintProbeFailure(s) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, data := printOneJSONEvent(t, cfg, tt.print)

			wantFields(t, data, map[string]any{"sourceAddress": "192.168.1.10:0"})

			if _, ok := data["timestamp"]; !ok {
				t.Error("field \"timestamp\" is missing")
			}
		})
	}
}

func TestJSONProbeUDPDetails(t *testing.T) {
	s := jsonTestStats()
	s.Protocol = config.UDP
	s.UDP = stats.UDPInfo{ProbeNumber: 7, Echoed: true, ReplySize: 12}

	_, data := printOneJSONEvent(t, Config{}, func(p *JSONPrinter) {
		p.PrintProbeSuccess(s)
	})

	udp, ok := data["udp"].(map[string]any)
	if !ok {
		t.Fatalf("field \"udp\" is missing from %v", data)
	}

	wantFields(t, udp, map[string]any{
		"probeNumber": float64(7),
		"rejected":    false,
		"echoed":      true,
		"replyBytes":  float64(12),
	})
}

func TestJSONProbeHTTPDetails(t *testing.T) {
	https := func() *stats.Statistics {
		s := jsonTestStats()
		s.Protocol = config.HTTPS
		s.HTTP = stats.HTTPInfo{
			StatusCode:      200,
			Status:          "200 OK",
			Proto:           "HTTP/2.0",
			TLSVersion:      "TLS 1.3",
			TLSCipherSuite:  "TLS_AES_128_GCM_SHA256",
			CertExpiry:      time.Now().Add(48 * time.Hour),
			ConnectDuration: 10 * time.Millisecond,
			TLSDuration:     5 * time.Millisecond,
			TimeToFirstByte: 20 * time.Millisecond,
		}
		return s
	}

	t.Run("https reports the TLS details too", func(t *testing.T) {
		_, data := printOneJSONEvent(t, Config{}, func(p *JSONPrinter) {
			p.PrintProbeSuccess(https())
		})

		httpData, ok := data["http"].(map[string]any)
		if !ok {
			t.Fatalf("field \"http\" is missing from %v", data)
		}

		wantFields(t, httpData, map[string]any{
			"statusCode":        float64(200),
			"status":            "200 OK",
			"version":           "HTTP/2.0",
			"tlsVersion":        "TLS 1.3",
			"tlsCipherSuite":    "TLS_AES_128_GCM_SHA256",
			"connectMs":         "10.000",
			"tlsHandshakeMs":    "5.000",
			"timeToFirstByteMs": "20.000",
		})
	})

	t.Run("plain http has no TLS details", func(t *testing.T) {
		s := https()
		s.Protocol = config.HTTP
		s.HTTP.TLSVersion = ""
		s.HTTP.TLSCipherSuite = ""
		s.HTTP.CertExpiry = time.Time{}

		_, data := printOneJSONEvent(t, Config{}, func(p *JSONPrinter) {
			p.PrintProbeSuccess(s)
		})

		httpData, ok := data["http"].(map[string]any)
		if !ok {
			t.Fatalf("field \"http\" is missing from %v", data)
		}

		wantNoFields(t, httpData, "tlsVersion", "tlsCipherSuite", "certificateExpiry", "tlsHandshakeMs")
	})

	t.Run("a probe that got no response has no http block", func(t *testing.T) {
		s := https()
		s.HTTP.StatusCode = 0

		_, data := printOneJSONEvent(t, Config{}, func(p *JSONPrinter) {
			p.PrintProbeFailure(s)
		})

		wantNoFields(t, data, "http")
	})
}

func TestJSONPrintStatistics(t *testing.T) {
	s := jsonTestStats()
	s.TotalSuccessfulProbes = 3
	s.TotalUnsuccessfulProbes = 1
	s.LastSuccessfulProbe = time.Now()
	s.LastUnsuccessfulProbe = time.Now()
	s.TotalUptime = 3 * time.Second
	s.TotalDowntime = time.Second
	s.LongestUptime = stats.LongestTime{Start: time.Now(), End: time.Now(), Duration: 3 * time.Second}
	s.LongestDowntime = stats.LongestTime{Start: time.Now(), End: time.Now(), Duration: time.Second}
	s.RetriedHostnameLookups = 2
	s.RTTResults = stats.RTTResult{Min: 1, Average: 2, Max: 3, Mdev: 0.5}
	s.EndTime = s.StartTime.Add(4 * time.Second)

	eventType, data := printOneJSONEvent(t, Config{}, func(p *JSONPrinter) {
		p.PrintStatistics(s)
	})

	if eventType != "statistics" {
		t.Errorf("event type = %q, want %q", eventType, "statistics")
	}

	wantFields(t, data, map[string]any{
		"hostname":               "example.com",
		"ipAddress":              "93.184.216.34",
		"port":                   float64(443),
		"protocol":               "TCP",
		"totalProbes":            float64(4),
		"successfulProbes":       float64(3),
		"unsuccessfulProbes":     float64(1),
		"packetLossPercent":      float64(25),
		"totalUptime":            "3 seconds",
		"totalDowntime":          "1 second",
		"longestUptime":          "3 seconds",
		"longestDowntime":        "1 second",
		"hostnameResolveRetries": float64(2),
		"latencyMin":             float64(1),
		"latencyAvg":             float64(2),
		"latencyMax":             float64(3),
		"latencyMdev":            float64(0.5),
		"duration":               "00:00:04",
	})
}

// A run that never succeeded has no latencies to report, and a run that is
// still going has not ended yet.
func TestJSONStatisticsLeavesOutWhatDidNotHappen(t *testing.T) {
	s := jsonTestStats()
	s.TotalUnsuccessfulProbes = 2
	s.DestIsIP = true

	_, data := printOneJSONEvent(t, Config{}, func(p *JSONPrinter) {
		p.PrintStatistics(s)
	})

	wantNoFields(t,
		data,
		"latencyMin",
		"latencyAvg",
		"latencyMax",
		"latencyMdev",
		"lastSuccessfulProbe",
		"longestUptime",
		"longestDowntime",
		"hostnameResolveRetries",
		"hostnameChanges",
		"endTime",
	)
}

func TestJSONStatisticsHostnameChanges(t *testing.T) {
	s := jsonTestStats()
	s.HostnameChanges = []stats.HostnameChange{
		{Addr: netip.MustParseAddr("93.184.216.34"), When: time.Now(), Duration: 12 * time.Millisecond},
		{Addr: netip.MustParseAddr("93.184.216.35"), When: time.Now(), Duration: 9 * time.Millisecond},
	}

	_, data := printOneJSONEvent(t, Config{}, func(p *JSONPrinter) {
		p.PrintStatistics(s)
	})

	changes, ok := data["hostnameChanges"].([]any)
	if !ok {
		t.Fatalf("field \"hostnameChanges\" is missing from %v", data)
	}

	if len(changes) != 2 {
		t.Fatalf("reported %d hostname changes, want 2", len(changes))
	}

	first, _ := changes[0].(map[string]any)
	wantFields(t, first, map[string]any{"addr": "93.184.216.34", "durationMs": "12.000"})
}

func TestJSONSimpleEvents(t *testing.T) {
	s := jsonTestStats()
	s.CurrentDowntime = 2 * time.Second
	s.CurrentUptime = 5 * time.Second

	tests := []struct {
		name      string
		print     func(p *JSONPrinter)
		eventType string
		want      map[string]any
	}{
		{
			name:      "name resolution",
			print:     func(p *JSONPrinter) { p.PrintNameResolutionDuration(s) },
			eventType: "nameResolution",
			want:      map[string]any{"durationMs": "12.000"},
		},
		{
			name:      "retry",
			print:     func(p *JSONPrinter) { p.PrintRetryingToResolve("example.com") },
			eventType: "retry",
			want:      map[string]any{"hostname": "example.com"},
		},
		{
			name:      "downtime",
			print:     func(p *JSONPrinter) { p.PrintDownTimeDuration(s) },
			eventType: "downtimeDuration",
			want:      map[string]any{"duration": "2 seconds", "precededByUptime": "5 seconds"},
		},
		{
			name:      "uptime",
			print:     func(p *JSONPrinter) { p.PrintUpTimeDuration(s) },
			eventType: "uptimeDuration",
			want:      map[string]any{"duration": "5 seconds", "precededByDowntime": "2 seconds"},
		},
		{
			name:      "error",
			print:     func(p *JSONPrinter) { p.PrintError("could not resolve %s", "example.com") },
			eventType: "error",
			want:      map[string]any{"message": "could not resolve example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventType, data := printOneJSONEvent(t, Config{}, tt.print)

			if eventType != tt.eventType {
				t.Errorf("event type = %q, want %q", eventType, tt.eventType)
			}

			wantFields(t, data, tt.want)
		})
	}
}

// Every event has to be on a line of its own so the output can be piped
// into something that reads it a line at a time.
func TestJSONPrintsOneEventPerLine(t *testing.T) {
	s := jsonTestStats()

	out := captureStdout(t, func() {
		p := NewJSONPrinter(Config{})
		p.PrintStart(s)
		p.PrintProbeSuccess(s)
		p.PrintProbeFailure(s)
	})

	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("printed %d lines, want 3, got %q", len(lines), out)
	}

	for _, line := range lines {
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Errorf("line %q is not valid JSON on its own: %v", line, err)
		}
	}
}

func TestJSONPrettyPrint(t *testing.T) {
	out := captureStdout(t, func() {
		NewJSONPrinter(Config{PrettyJSON: true}).PrintStart(jsonTestStats())
	})

	if !strings.Contains(out, "\n\t\"type\"") {
		t.Errorf("output is not indented: %q", out)
	}
}

// Shutdown is what prints the summary on exit, so it has to produce the
// same event PrintStatistics does.
func TestJSONShutdownPrintsStatistics(t *testing.T) {
	eventType, _ := printOneJSONEvent(t, Config{}, func(p *JSONPrinter) {
		p.Shutdown(jsonTestStats())
	})

	if eventType != "statistics" {
		t.Errorf("event type = %q, want %q", eventType, "statistics")
	}
}

// With --no-stats, exiting emits no event at all.
func TestJSONShutdownOmitsStatistics(t *testing.T) {
	events := jsonEvents(t, Config{OmitStatistics: true}, func(p *JSONPrinter) {
		p.Shutdown(jsonTestStats())
	})

	if len(events) != 0 {
		t.Errorf("printed %d events, want 0", len(events))
	}
}
