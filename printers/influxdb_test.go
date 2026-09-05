package printers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/config"

	"github.com/pouriyajamshidi/tcping/v3/stats"
)

// influxDBTestConfig is a complete configuration, so a test only has to say
// what it wants to be different.
func influxDBTestConfig(url string) Config {
	return Config{
		InfluxDBURL:    url,
		InfluxDBOrg:    "home",
		InfluxDBBucket: "tcping",
		InfluxDBToken:  "secret",
		SourceLabel:    "probe-1",
	}
}

// influxDBTestStats is a probe that succeeded in 3.5ms, which is enough to
// check what a successful probe writes.
func influxDBTestStats() *stats.Statistics {
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

func TestInfluxDBEndpoint(t *testing.T) {
	tests := []struct {
		given string
		want  string
	}{
		{"http://localhost:8086", "http://localhost:8086/api/v2/write?bucket=tcping&org=home&precision=ns"},
		{"http://localhost:8086/", "http://localhost:8086/api/v2/write?bucket=tcping&org=home&precision=ns"},
		{"http://localhost:8086/api/v2/write", "http://localhost:8086/api/v2/write?bucket=tcping&org=home&precision=ns"},
		{"localhost:8086", "http://localhost:8086/api/v2/write?bucket=tcping&org=home&precision=ns"},
	}

	for _, tt := range tests {
		p, err := NewInfluxDBPrinter(influxDBTestConfig(tt.given))
		if err != nil {
			t.Fatalf("NewInfluxDBPrinter(%q) returned %v", tt.given, err)
		}

		if p.endpoint != tt.want {
			t.Errorf("NewInfluxDBPrinter(%q) endpoint = %q, want %q", tt.given, p.endpoint, tt.want)
		}
	}
}

// Writing without a bucket or a token would only fail later, on the first
// probe, which is far too late to tell the user about it.
func TestInfluxDBNeedsOrgBucketAndToken(t *testing.T) {
	tests := []struct {
		name    string
		missing func(cfg *Config)
	}{
		{"no organization", func(cfg *Config) { cfg.InfluxDBOrg = "" }},
		{"no bucket", func(cfg *Config) { cfg.InfluxDBBucket = "" }},
		{"no token", func(cfg *Config) { cfg.InfluxDBToken = "" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := influxDBTestConfig("http://localhost:8086")
			tt.missing(&cfg)

			if _, err := NewInfluxDBPrinter(cfg); err == nil {
				t.Errorf("expected an error with %s", tt.name)
			}
		})
	}
}

// influxDBServer records every line it is written, so a test can look at
// what actually went over the wire.
func influxDBServer(t *testing.T, writes *[][]string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		if got := r.Header.Get("Authorization"); got != "Token secret" {
			t.Errorf("Authorization = %q, want %q", got, "Token secret")
		}

		*writes = append(*writes, strings.Split(strings.TrimSpace(string(body)), "\n"))

		w.WriteHeader(http.StatusNoContent)
	}))
}

func TestInfluxDBPrintProbeSuccess(t *testing.T) {
	var writes [][]string

	server := influxDBServer(t, &writes)
	defer server.Close()

	printer, err := NewInfluxDBPrinter(influxDBTestConfig(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	printer.PrintProbeSuccess(influxDBTestStats())

	probe := writes[0][0]

	want := "tcping_tcp,source=probe-1,target=example.com,port=443,protocol=TCP " +
		`success=1i,successful_probes=2i,unsuccessful_probes=0i,ip="93.184.216.34",rtt_ms=3.5 `

	if !strings.HasPrefix(probe, want) {
		t.Errorf("probe line = %q, want it to start with %q", probe, want)
	}

	// The line ends with a nanosecond timestamp, which is the only part we
	// cannot spell out here.
	fields := strings.Split(probe, " ")
	if len(fields) != 3 {
		t.Fatalf("probe line has %d space separated parts, want 3: %q", len(fields), probe)
	}

	if len(fields[2]) < 19 {
		t.Errorf("timestamp = %q, want nanoseconds since the epoch", fields[2])
	}
}

// A failed probe has no round trip time, so writing one would be a lie.
func TestInfluxDBPrintProbeFailureHasNoRTT(t *testing.T) {
	var writes [][]string

	server := influxDBServer(t, &writes)
	defer server.Close()

	printer, err := NewInfluxDBPrinter(influxDBTestConfig(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	printer.PrintProbeFailure(influxDBTestStats())

	if strings.Contains(writes[0][0], "rtt_ms") {
		t.Errorf("a failed probe should not write an RTT: %q", writes[0][0])
	}
}

// An HTTP probe writes its own measurement, holding everything the probe
// learned, so a query for HTTP timings does not have to sift TCP probes out.
func TestInfluxDBHTTPProbeWritesItsOwnMeasurement(t *testing.T) {
	var writes [][]string

	server := influxDBServer(t, &writes)
	defer server.Close()

	printer, err := NewInfluxDBPrinter(influxDBTestConfig(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	s := influxDBTestStats()
	s.Protocol = config.HTTPS
	s.HTTP.StatusCode = 200
	s.HTTP.TLSVersion = "TLS 1.3"
	s.HTTP.CertExpiry = time.Now().Add(30 * 24 * time.Hour)
	s.HTTP.ConnectDuration = 12 * time.Millisecond
	s.HTTP.TimeToFirstByte = 40 * time.Millisecond
	s.HTTP.TLSDuration = 20 * time.Millisecond

	printer.PrintProbeSuccess(s)

	line := writes[0][0]

	for _, want := range []string{
		"tcping_http,",
		"status_code=200i",
		"connect_ms=12",
		"ttfb_ms=40",
		"tls_handshake_ms=20",
		"certificate_days_remaining=29i",
	} {
		if !strings.Contains(line, want) {
			t.Errorf("HTTP line = %q, want it to contain %q", line, want)
		}
	}
}

// A UDP probe cannot say much, so the little it does learn has to be kept:
// whether the reply was our own payload coming back and whether the port
// refused us.
func TestInfluxDBUDPProbeWritesWhatItLearned(t *testing.T) {
	var writes [][]string

	server := influxDBServer(t, &writes)
	defer server.Close()

	printer, err := NewInfluxDBPrinter(influxDBTestConfig(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	s := influxDBTestStats()
	s.Protocol = config.UDP
	s.UDP.Echoed = true
	s.UDP.ProbeNumber = 7
	s.UDP.ReplySize = 4

	printer.PrintProbeSuccess(s)

	line := writes[0][0]

	for _, want := range []string{
		"tcping_udp,",
		"probe_number=7i",
		"reply_bytes=4i",
		"echoed=true",
		"rejected=false",
	} {
		if !strings.Contains(line, want) {
			t.Errorf("UDP line = %q, want it to contain %q", line, want)
		}
	}
}

// A hostname could hold a comma or a space, which would otherwise split the
// line in the wrong place and be rejected.
func TestInfluxDBEscapesTags(t *testing.T) {
	var writes [][]string

	server := influxDBServer(t, &writes)
	defer server.Close()

	printer, err := NewInfluxDBPrinter(influxDBTestConfig(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	s := influxDBTestStats()
	s.Hostname = "we ird,host=name"

	printer.PrintProbeSuccess(s)

	if !strings.Contains(writes[0][0], `target=we\ ird\,host\=name`) {
		t.Errorf("tags were not escaped: %q", writes[0][0])
	}
}

// An InfluxDB that is not there must not stop the probing, and must not
// repeat the same complaint on every probe.
func TestInfluxDBKeepsGoingWhenUnreachable(t *testing.T) {
	printer, err := NewInfluxDBPrinter(influxDBTestConfig("http://127.0.0.1:1"))
	if err != nil {
		t.Fatal(err)
	}

	printer.PrintProbeSuccess(influxDBTestStats())

	if !printer.warned {
		t.Error("expected the printer to warn about an unreachable InfluxDB")
	}

	printer.PrintProbeSuccess(influxDBTestStats())
}

// The run summary has to keep flowing on its own, otherwise a tcping that
// nobody stops never writes one.
func TestInfluxDBStatisticsRideAlongWithProbes(t *testing.T) {
	var writes [][]string

	server := influxDBServer(t, &writes)
	defer server.Close()

	cfg := influxDBTestConfig(server.URL)
	cfg.InfluxDBStatsInterval = 10 * time.Second

	printer, err := NewInfluxDBPrinter(cfg)
	if err != nil {
		t.Fatal(err)
	}

	hasSummary := func(lines []string) bool {
		for _, line := range lines {
			if strings.HasPrefix(line, "tcping_statistics,") {
				return true
			}
		}
		return false
	}

	// The first probe carries a summary, so the metrics show up right away.
	printer.PrintProbeSuccess(influxDBTestStats())
	if !hasSummary(writes[0]) {
		t.Error("the first probe should carry the run summary")
	}

	// The next one does not, since the interval has not passed.
	printer.PrintProbeSuccess(influxDBTestStats())
	if hasSummary(writes[1]) {
		t.Error("the summary should not be written with every probe")
	}

	// Once it has, it comes along again.
	printer.lastStats = time.Now().Add(-printer.statsInterval)
	printer.PrintProbeSuccess(influxDBTestStats())
	if !hasSummary(writes[2]) {
		t.Error("the summary should be written again once the interval passed")
	}
}

// The InfluxDB printer prints nothing after this line, so it is the only
// place the user gets to see which address is being probed.
func TestInfluxDBPrintStartShowsTheIP(t *testing.T) {
	p, err := NewInfluxDBPrinter(influxDBTestConfig("http://localhost:8086"))
	if err != nil {
		t.Fatalf("NewInfluxDBPrinter returned %v", err)
	}

	s := influxDBTestStats()
	s.NameResolutionDuration = 12 * time.Millisecond

	t.Run("hostname target", func(t *testing.T) {
		out := captureStdout(t, func() { p.PrintStart(s) })

		want := "Probing example.com (93.184.216.34) on port 443 over TCP (resolved in 12.000 ms) - sending metrics to: " + p.endpoint + "\n"
		if out != want {
			t.Errorf("output = %q, want %q", out, want)
		}
	})

	t.Run("IP target", func(t *testing.T) {
		ipTarget := influxDBTestStats()
		ipTarget.Hostname = "93.184.216.34"
		ipTarget.DestIsIP = true

		out := captureStdout(t, func() { p.PrintStart(ipTarget) })

		want := "Probing 93.184.216.34 on port 443 over TCP - sending metrics to: " + p.endpoint + "\n"
		if out != want {
			t.Errorf("output = %q, want %q", out, want)
		}
	})
}

// The resolved IP has to stay a field. As a tag it would identify the
// series, so a hostname that resolves somewhere else mid-run would leave
// the old series behind and start a new one.
func TestInfluxDBResolvedIPDoesNotSplitTheSeries(t *testing.T) {
	var writes [][]string

	server := influxDBServer(t, &writes)
	defer server.Close()

	printer, err := NewInfluxDBPrinter(influxDBTestConfig(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	s := influxDBTestStats()
	printer.PrintProbeSuccess(s)

	s.IP = netip.MustParseAddr("93.184.216.35")
	printer.PrintProbeSuccess(s)

	first := strings.Split(writes[0][0], " ")[0]
	second := strings.Split(writes[1][0], " ")[0]

	if first != second {
		t.Errorf("the address change moved the point to another series:\n%s\n%s", first, second)
	}

	if !strings.Contains(writes[1][0], `ip="93.184.216.35"`) {
		t.Errorf("the new address was not written: %q", writes[1][0])
	}
}
