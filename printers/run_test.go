package printers

import (
	"bytes"
	"context"
	"errors"
	"net/netip"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/config"
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

func (p *scriptedPinger) Ping(context.Context, netip.Addr) (probe.ProbeResult, error) {
	index := p.next
	p.next++

	if p.after != nil {
		defer p.after(index)
	}

	if index < len(p.outcomes) && p.outcomes[index] {
		return probe.ProbeResult{}, nil
	}

	return probe.ProbeResult{}, errors.New("connection refused")
}

// scriptedRun is one replayed run: which probes succeed, where their results
// go, and after how many probes the user presses "Enter" for a mid-run
// summary. Zero means they never do.
type scriptedRun struct {
	printer          probe.Printer
	outcomes         []bool
	showFailuresOnly bool
	enterAfter       int
}

// run probes once per entry in outcomes and returns everything that reached
// the terminal, including the summary printed on the way out.
func (r scriptedRun) run(t *testing.T) string {
	t.Helper()

	cfg := config.Config{
		Hostname:               "example.com",
		IP:                     netip.MustParseAddr("93.184.216.34"),
		Port:                   443,
		Protocol:               config.TCP,
		NameResolutionDuration: 12 * time.Millisecond,
		IntervalBetweenProbes:  20 * time.Millisecond,
		ProbesBeforeQuit:       uint(len(r.outcomes)),
		ShowFailuresOnly:       r.showFailuresOnly,
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
		"No reply from example.com (93.184.216.34) on port 443 TCP_conn=1 (up for <duration>)",
		"No reply from example.com (93.184.216.34) on port 443 TCP_conn=2",
		"Reply from example.com (93.184.216.34) on port 443 TCP_conn=1 time=<duration> (down for <duration>)",
		"Reply from example.com (93.184.216.34) on port 443 TCP_conn=2 time=<duration>",
	}, "\n") + "\n"

	if got := probeLines(out); got != want {
		t.Errorf("transcript =\n%s\nwant =\n%s", got, want)
	}
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

// With --failures-only the success line is held back, so the downtime it
// would have carried is reported on a line of its own instead of being lost.
func TestRunTranscript_ShowFailuresOnly(t *testing.T) {
	out := scriptedRun{
		printer:          NewPlainPrinter(Config{ShowFailuresOnly: true}),
		outcomes:         []bool{true, false, true, true},
		showFailuresOnly: true,
	}.run(t)

	want := strings.Join([]string{
		"Probing example.com on port 443 over TCP (resolved in <duration>)",
		"No reply from example.com (93.184.216.34) on port 443 TCP_conn=1 (up for <duration>)",
		"No response received for <duration>",
	}, "\n") + "\n"

	if got := probeLines(out); got != want {
		t.Errorf("transcript =\n%s\nwant =\n%s", got, want)
	}
}

// The JSON output is read by other programs, so the run has to come through
// as the same sequence of events the terminal shows as lines.
func TestRunTranscript_JSONEvents(t *testing.T) {
	var events bytes.Buffer

	scriptedRun{
		printer:    NewJSONPrinter(Config{Writer: &events}),
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
		"probe", "uptimeDuration", // the first failure ends the uptime
		"probe",
		"statistics",                // "Enter" pressed after the fourth probe
		"probe", "downtimeDuration", // the recovery ends the downtime
		"probe",
		"statistics", // on the way out
	}

	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("events = %v, want %v", got, want)
	}

	// The mid-run summary has the same gap to fall into as the terminal one.
	if strings.Contains(out, `"totalUptime":"0 seconds"`) {
		t.Errorf("a summary reports no uptime on a run that was up:\n%s", out)
	}
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
}
