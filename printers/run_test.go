package printers

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/config"
	"github.com/pouriyajamshidi/tcping/v3/dns"
	"github.com/pouriyajamshidi/tcping/v3/nic"
	"github.com/pouriyajamshidi/tcping/v3/probe"
	"github.com/pouriyajamshidi/tcping/v3/stats"
)

// These tests replay a whole run through the real Prober and a real printer
// and check what it produced. The per-method tests elsewhere in this package
// cannot see a line that is right on its own but wrong for the run it appeared
// in, which is how the totals and the uptime and downtime reports drifted
// apart from what actually happened.

// scriptedPinger answers each probe from a fixed list of outcomes, so a run
// can be replayed without touching the network. after runs once the outcome
// at that index has been handed out, which is where a test presses "Enter".
type scriptedPinger struct {
	outcomes []bool
	next     int
	after    func(probeIndex int)
}

func (p *scriptedPinger) Ping(context.Context, netip.Addr) (probe.Result, error) {
	index := p.next
	p.next++

	if p.after != nil {
		defer p.after(index)
	}

	if index < len(p.outcomes) && p.outcomes[index] {
		return probe.Result{}, nil
	}

	return probe.Result{}, errors.New("connection refused")
}

// scriptedRun is one replayed run: which probes succeed, where their results
// go, and after how many probes the user presses "Enter" for a mid-run
// summary. Zero means they never do.
type scriptedRun struct {
	printer          probe.Printer
	outcomes         []bool
	showFailuresOnly bool
	enterAfter       int
	protocol         config.Protocol // TCP when not given.
	proxy            *url.URL
	resolver         *dns.Resolver // When given, looks the hostname up before every probe.
}

// run probes once per entry in outcomes and returns everything that reached
// the terminal, including the summary printed on the way out.
func (r scriptedRun) run(t *testing.T) string {
	t.Helper()

	protocol := r.protocol
	if protocol == "" {
		protocol = config.TCP
	}

	cfg := config.Config{
		Hostname:               "example.com",
		IP:                     netip.MustParseAddr("93.184.216.34"),
		Port:                   443,
		Protocol:               protocol,
		Proxy:                  r.proxy,
		NameResolutionDuration: 12 * time.Millisecond,
		IntervalBetweenProbes:  20 * time.Millisecond,
		ProbesBeforeQuit:       uint(len(r.outcomes)),
		ShowFailuresOnly:       r.showFailuresOnly,
		Resolver:               r.resolver,
		ResolveEveryProbe:      r.resolver != nil,
	}

	// Through a proxy, the proxy looks the hostname up, so there is no
	// address and no lookup time, the same as internal/cli leaves them.
	if r.proxy != nil {
		cfg.IP = netip.Addr{}
		cfg.NameResolutionDuration = 0
	}

	requests := make(chan struct{}, 1)

	pinger := &scriptedPinger{outcomes: r.outcomes}
	if r.enterAfter > 0 {
		pinger.after = func(index int) {
			if index+1 == r.enterAfter {
				requests <- struct{}{}
			}
		}
	}

	s := stats.NewStatistics(cfg)
	prober := probe.NewProber(pinger, r.printer, cfg, s, requests)

	return captureStdout(t, func() {
		if err := prober.Probe(context.Background()); err != nil {
			t.Errorf("Probe() failed: %v", err)
			return
		}

		r.printer.Shutdown(s)
	})
}

var (
	timestampPattern = regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`)
	rttPattern       = regexp.MustCompile(`[\d.]+/[\d.]+/[\d.]+/[\d.]+ ms`)
	// Every shape durationToString can produce, plus the millisecond form
	// the probe lines use.
	durationPattern = regexp.MustCompile(`\d+ hours? \d+ minutes? \d+ seconds?|\d+ minutes? \d+ seconds?|\d+ (hour|minute)|[\d.]+ (seconds?|ms)`)
)

// mask replaces everything that depends on the clock, so the rest of a
// transcript can be compared as plain text.
func mask(out string) string {
	out = timestampPattern.ReplaceAllString(out, "<time>")
	out = rttPattern.ReplaceAllString(out, "<rtt>")
	out = durationPattern.ReplaceAllString(out, "<duration>")

	return out
}

// probeLines is the part of a transcript before the first summary.
func probeLines(out string) string {
	lines, _, _ := strings.Cut(out, "\n---")

	return mask(lines)
}

// upDownUp is a run that is up, goes down for two probes, then comes back.
var upDownUp = []bool{true, true, false, false, true, true}

// The uptime that a failure ends, and the downtime that a success ends, are
// reported on the probe line that ended them and nowhere else. Reporting them
// on a line of their own used to repeat a period from an earlier cycle, so a
// recovery claimed an uptime that had ended minutes ago.
func TestRunTranscript_UpDownUp(t *testing.T) {
	out := scriptedRun{printer: NewPlainPrinter(Config{}), outcomes: upDownUp}.run(t)

	want := strings.Join([]string{
		"Probing example.com on port 443 over TCP (resolved in <duration>)",
		"Reply from example.com (93.184.216.34) on port 443 TCP_conn=1 time=<duration>",
		"Reply from example.com (93.184.216.34) on port 443 TCP_conn=2 time=<duration>",
		"No reply from example.com (93.184.216.34) on port 443 TCP_conn=1 (after <duration> of uptime)",
		"No reply from example.com (93.184.216.34) on port 443 TCP_conn=2",
		"Reply from example.com (93.184.216.34) on port 443 TCP_conn=1 time=<duration> (after <duration> of downtime)",
		"Reply from example.com (93.184.216.34) on port 443 TCP_conn=2 time=<duration>",
	}, "\n") + "\n"

	if got := probeLines(out); got != want {
		t.Errorf("transcript =\n%s\nwant =\n%s", got, want)
	}
}

// A summary with more downtime than its longest outage is not a bug, it is a
// run with more than one outage. The count is what says so, and it has to
// cover an opening failure, which starts a downtime before anything was up.
func TestRunTranscript_OutageCount(t *testing.T) {
	out := scriptedRun{
		printer:  NewPlainPrinter(Config{}),
		outcomes: []bool{false, true, true, false, false, true},
	}.run(t)

	wantLines(t, out,
		"outages:             2\n",
	)
}

// The colored output is the default one, so it is the one most people read.
// Only the color may differ from the plain output, never the words, and the
// escape codes are off while the output is captured.
func TestRunTranscript_ColorSaysTheSameAsPlain(t *testing.T) {
	plain := scriptedRun{printer: NewPlainPrinter(Config{}), outcomes: upDownUp, enterAfter: 4}.run(t)
	color := scriptedRun{printer: NewColorPrinter(Config{}), outcomes: upDownUp, enterAfter: 4}.run(t)

	if mask(plain) != mask(color) {
		t.Errorf("the plain and colored output differ.\nplain =\n%s\ncolor =\n%s", mask(plain), mask(color))
	}
}

// A run that never fails is up for its whole length, so its summary cannot
// report no uptime at all. It used to, because uptime was only added to the
// total once it ended.
func TestRunTranscript_SummaryOfARunThatNeverFailed(t *testing.T) {
	out := scriptedRun{printer: NewPlainPrinter(Config{}), outcomes: []bool{true, true, true, true}}.run(t)

	if strings.Contains(out, "total uptime:   0 seconds") {
		t.Errorf("summary reports no uptime after four successful probes:\n%s", out)
	}

	wantLines(t, out,
		"4 TCP probes transmitted on port 443 | 4 received, 0.00% packet loss\n",
		"total downtime: 0 seconds\n",
		"longest consecutive uptime:   ",
	)
}

// Pressing "Enter" mid-run has to report the run as it stands, not as if it
// had never started. The summary used to leave out the uptime it was in the
// middle of, and to work out its duration from an end time it did not have
// yet, which always printed the same nonsense value.
func TestRunTranscript_MidRunSummary(t *testing.T) {
	out := scriptedRun{
		printer:    NewPlainPrinter(Config{}),
		outcomes:   []bool{true, true, true, true, true, true},
		enterAfter: 3,
	}.run(t)

	// Everything up to the first "TCPing started at" is the mid-run summary
	// and the probe lines before it. The final summary follows.
	midRun, final, found := strings.Cut(out, "TCPing started at")
	if !found {
		t.Fatalf("no summary in the output:\n%s", out)
	}

	if strings.Contains(midRun, "total uptime:   0 seconds") {
		t.Errorf("mid-run summary reports no uptime after three successful probes:\n%s", midRun)
	}

	if !strings.Contains(midRun, "longest consecutive uptime:   ") {
		t.Errorf("mid-run summary leaves out the uptime it is in the middle of:\n%s", midRun)
	}

	// The run has not ended, so it has no end time to report.
	if strings.Contains(midRun, "TCPing ended at") {
		t.Errorf("mid-run summary claims the run has ended:\n%s", midRun)
	}

	if !strings.Contains(final, "TCPing ended at:   ") {
		t.Errorf("final summary leaves out the end time:\n%s", final)
	}

	// A run of six 20 ms probes takes well under ten seconds. The duration
	// used to come out of a zero end time, which always printed 00:12:43.
	duration := regexp.MustCompile(`duration \(HH:MM:SS\): (\S+)`)
	fewSeconds := regexp.MustCompile(`^00:00:0\d$`)

	for _, match := range duration.FindAllStringSubmatch(out, -1) {
		if !fewSeconds.MatchString(match[1]) {
			t.Errorf("duration = %q, want a few seconds at most", match[1])
		}
	}
}

// --failures-only means failures only: the successful probes are not printed,
// and neither is the outage they ended, since it is the success line that
// carries it.
func TestRunTranscript_ShowFailuresOnly(t *testing.T) {
	out := scriptedRun{
		printer:          NewPlainPrinter(Config{}),
		outcomes:         []bool{true, false, true, true},
		showFailuresOnly: true,
	}.run(t)

	want := strings.Join([]string{
		"Probing example.com on port 443 over TCP (resolved in <duration>)",
		"No reply from example.com (93.184.216.34) on port 443 TCP_conn=1 (after <duration> of uptime)",
	}, "\n") + "\n"

	if got := probeLines(out); got != want {
		t.Errorf("transcript =\n%s\nwant =\n%s", got, want)
	}
}

// Through a proxy, the connect time and the source address are those of the
// connection to the proxy, so the run has to say which proxy that was. The
// proxy URL can carry a password, which must never reach the output. The
// proxy looks the hostname up, so no address or lookup time of our own shows
// up anywhere, not even as an empty pair of brackets.
func TestRunTranscript_ThroughAProxy(t *testing.T) {
	proxy, _ := url.Parse("http://user:secret@proxy.example.com:3128")

	plain := scriptedRun{
		printer:  NewPlainPrinter(Config{}),
		outcomes: []bool{true},
		protocol: config.HTTPS,
		proxy:    proxy,
	}.run(t)

	want := strings.Join([]string{
		"Probing example.com on port 443 over HTTPS via proxy http://proxy.example.com:3128",
		"Reply from example.com on port 443 HTTPS_conn=1 time=<duration>",
	}, "\n") + "\n"

	if got := probeLines(plain); got != want {
		t.Errorf("transcript =\n%s\nwant =\n%s", got, want)
	}

	wantLines(t, plain, "\n--- example.com TCPing statistics ---\n")
	wantNoLines(t, plain, "()", "resolve", "IP address changes")

	color := scriptedRun{
		printer:  NewColorPrinter(Config{}),
		outcomes: []bool{true},
		protocol: config.HTTPS,
		proxy:    proxy,
	}.run(t)

	if mask(plain) != mask(color) {
		t.Errorf("the plain and colored output differ.\nplain =\n%s\ncolor =\n%s", mask(plain), mask(color))
	}

	wantNoLines(t, plain, "user", "secret")
}

// The JSON output is read by other programs, so the run has to come through
// as the same sequence of events the terminal shows as lines.
func TestRunTranscript_JSONEvents(t *testing.T) {
	var events bytes.Buffer

	scriptedRun{
		printer:    jsonTestPrinter(t, Config{Writer: &events}),
		outcomes:   upDownUp,
		enterAfter: 4,
	}.run(t)

	out := events.String()

	eventType := regexp.MustCompile(`"type":"([a-zA-Z]+)"`)

	var got []string
	for _, match := range eventType.FindAllStringSubmatch(out, -1) {
		got = append(got, match[1])
	}

	want := []string{
		"start",
		"probe", "probe", // up
		"probe", "probe", // down
		"statistics",     // "Enter" pressed after the fourth probe
		"probe", "probe", // up again
		"statistics", // on the way out
	}

	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("events = %v, want %v", got, want)
	}

	// The probe that ended each period is the one that reports its length.
	wantLines(t, out,
		`"success":false,"connections":1,"endedUptime":`,
		`"connections":1,"endedDowntime":`,
	)

	// The mid-run summary has the same gap to fall into as the terminal one.
	if strings.Contains(out, `"totalUptime":"0 seconds"`) {
		t.Errorf("a summary reports no uptime on a run that was up:\n%s", out)
	}
}

// Through a proxy there is no address of our own to report, so the JSON
// leaves the field out rather than sending an empty one.
func TestRunTranscript_JSONThroughAProxy(t *testing.T) {
	var events bytes.Buffer
	proxy, _ := url.Parse("http://proxy.example.com:3128")

	scriptedRun{
		printer:  jsonTestPrinter(t, Config{Writer: &events}),
		outcomes: []bool{true, false},
		protocol: config.HTTPS,
		proxy:    proxy,
	}.run(t)

	out := events.String()

	wantLines(t, out, `"hostname":"example.com"`, `"proxy":"http://proxy.example.com:3128"`)
	wantNoLines(t, out, "ipAddress", "nameResolution", "hostnameResolveRetries")
}

// columnOf is one named column of a CSV, header row included, so a test can
// name a column instead of counting to it.
func columnOf(t *testing.T, rows [][]string, name string) []string {
	t.Helper()

	index := slices.Index(rows[0], name)
	if index < 0 {
		t.Fatalf("no %q column in %v", name, rows[0])
	}

	values := make([]string, len(rows))
	for i, row := range rows {
		values[i] = row[index]
	}

	return values
}

// The CSV rows are what a run leaves behind for later, so they have to add up
// the same way the summary does.
func TestRunTranscript_CSVRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "run.csv")

	printer, err := NewCSVPrinter(Config{OutputCSVPath: path, CSVNoTimestamp: true})
	if err != nil {
		t.Fatalf("NewCSVPrinter() failed: %v", err)
	}

	scriptedRun{printer: printer, outcomes: upDownUp}.run(t)

	rows := readCSV(t, printer, path)

	// The header, then one row per probe, reachable or not.
	want := []string{"Reachable", "true", "true", "false", "false", "true", "true"}
	if len(rows) != len(want) {
		t.Fatalf("wrote %d rows, want %d: %v", len(rows), len(want), rows)
	}

	for i, row := range rows {
		if row[0] != want[i] {
			t.Errorf("row %d starts with %q, want %q", i, row[0], want[i])
		}
	}

	// The ended uptime sits on the first failure, the ended downtime on the
	// recovery, and both columns are empty everywhere else.
	endedUptime := columnOf(t, rows, "Ended Uptime")
	endedDowntime := columnOf(t, rows, "Ended Downtime")

	if endedUptime[3] == "" {
		t.Errorf("the failure that ended the uptime has no Ended Uptime: %v", rows[3])
	}

	if endedDowntime[5] == "" {
		t.Errorf("the probe that ended the outage has no Ended Downtime: %v", rows[5])
	}

	for _, row := range []int{1, 2, 4, 6} {
		if endedUptime[row] != "" || endedDowntime[row] != "" {
			t.Errorf("row %d ended no period but reports one: %v", row, rows[row])
		}
	}

	statistics := map[string]string{}
	for _, row := range readCSV(t, printer, strings.TrimSuffix(path, ".csv")+"_stats.csv") {
		if len(row) == 2 {
			statistics[row[0]] = row[1]
		}
	}

	if statistics["Total Uptime"] == "0 seconds" {
		t.Errorf("stats file reports no uptime on a run that was up: %v", statistics)
	}

	if statistics["Total Downtime"] == "0 seconds" {
		t.Errorf("stats file reports no downtime on a run that went down: %v", statistics)
	}

	if statistics["Outages"] != "1" {
		t.Errorf("stats file reports %q outages on a run with one: %v", statistics["Outages"], statistics)
	}
}

// failingResolver is a resolver whose lookups always time out, since the
// DNS server it asks reads every question and never answers.
func failingResolver(t *testing.T) *dns.Resolver {
	t.Helper()

	server, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatalf("failed to start the DNS server: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })

	go func() {
		buf := make([]byte, 512)
		for {
			if _, _, err := server.ReadFrom(buf); err != nil {
				return
			}
		}
	}()

	resolver, err := dns.NewResolver(server.LocalAddr().String(), 100*time.Millisecond, false, false, nic.NetworkInterface{})
	if err != nil {
		t.Fatalf("dns.NewResolver() failed: %v", err)
	}

	return resolver
}

// A failed name lookup has nothing to do with where the results go, so it
// must not show up as "InfluxDB Error: lookup example.com: i/o timeout". The
// prefix is only for errors from InfluxDB, the OTLP endpoint, the CSV files
// or the database themselves.
func TestRunTranscript_LookupErrorIsNotBlamedOnTheOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	tests := []struct {
		name       string
		newPrinter func() (probe.Printer, error)
	}{
		{"InfluxDB", func() (probe.Printer, error) {
			return NewInfluxDBPrinter(influxDBTestConfig(server.URL))
		}},
		{"OTLP", func() (probe.Printer, error) {
			return NewOTLPPrinter(Config{OTLPURL: server.URL})
		}},
		{"CSV", func() (probe.Printer, error) {
			return NewCSVPrinter(Config{OutputCSVPath: filepath.Join(t.TempDir(), "run.csv"), CSVNoTimestamp: true})
		}},
		{"SQLite", func() (probe.Printer, error) {
			return NewSQLitePrinter(Config{Target: "example.com", Port: 443, OutputSQLitePath: ":memory:"})
		}},
	}

	resolver := failingResolver(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printer, err := tt.newPrinter()
			if err != nil {
				t.Skipf("no %s printer here: %v", tt.name, err)
			}

			stderr := captureStderr(t, func() {
				scriptedRun{printer: printer, outcomes: []bool{true}, resolver: resolver}.run(t)
			})

			lookups := 0
			for line := range strings.Lines(stderr) {
				if !strings.Contains(line, "lookup example.com") {
					continue
				}

				lookups++
				if !strings.HasPrefix(line, "lookup example.com") {
					t.Errorf("the failed lookup was printed as %q", line)
				}
			}

			if lookups != 1 {
				t.Errorf("printed %d failed lookups, want 1:\n%s", lookups, stderr)
			}
		})
	}
}

// With --full-page, the time on the probe line has to cover the whole
// download, and -v has to say how long the download took and how big it was.
// Without it, the time stops at the server's answer. This runs the real HTTP
// probe against a server whose body takes 300ms to finish.
func TestRunTranscript_FullPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("a", 1000)))
		w.(http.Flusher).Flush()

		time.Sleep(300 * time.Millisecond)
		_, _ = w.Write([]byte(strings.Repeat("b", 1000)))
	}))
	t.Cleanup(srv.Close)

	u, _ := url.Parse(srv.URL)
	addrPort := netip.MustParseAddrPort(u.Host)

	run := func(fullPage bool) string {
		cfg := config.Config{
			URL:                   srv.URL,
			Hostname:              addrPort.Addr().String(),
			IP:                    addrPort.Addr(),
			TargetIsIP:            true,
			Port:                  addrPort.Port(),
			Protocol:              config.HTTP,
			Timeout:               5 * time.Second,
			FullPage:              fullPage,
			IntervalBetweenProbes: 20 * time.Millisecond,
			ProbesBeforeQuit:      1,
		}

		printer := NewPlainPrinter(Config{Verbose: true})
		s := stats.NewStatistics(cfg)
		prober := probe.NewProber(probe.NewHTTPing(cfg), printer, cfg, s, nil)

		return captureStdout(t, func() {
			if err := prober.Probe(context.Background()); err != nil {
				t.Errorf("Probe() failed: %v", err)
			}
		})
	}

	msOf := func(t *testing.T, out, pattern string) float64 {
		t.Helper()

		m := regexp.MustCompile(pattern + `=([\d.]+) ms`).FindStringSubmatch(out)
		if m == nil {
			t.Fatalf("no %s= in the output:\n%s", pattern, out)
		}

		ms, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			t.Fatalf("could not read %s= from %q: %v", pattern, m[0], err)
		}

		return ms
	}

	t.Run("full page", func(t *testing.T) {
		out := run(true)

		wantLines(t, out,
			"HTTP_conn=1 status=200 time=",
			"    HTTP/1.1 200 OK\n",
			" ms size=2000 bytes\n",
		)

		if ms := msOf(t, out, "time"); ms < 300 {
			t.Errorf("time=%v ms, want it to cover the 300ms the body took:\n%s", ms, out)
		}
		if ms := msOf(t, out, "download"); ms < 300 {
			t.Errorf("download=%v ms, want it to cover the 300ms the body took:\n%s", ms, out)
		}
	})

	t.Run("first byte only", func(t *testing.T) {
		out := run(false)

		wantNoLines(t, out, "download=", "size=")

		if ms := msOf(t, out, "time"); ms >= 300 {
			t.Errorf("time=%v ms, want it to stop at the server's answer:\n%s", ms, out)
		}
	})
}
