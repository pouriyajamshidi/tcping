package printers

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/pouriyajamshidi/tcping/v3/stats"
)

// captureStdout runs f and returns whatever it wrote to stdout.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}

	original := os.Stdout
	os.Stdout = w
	// Escape codes would only get in the way of comparing the output.
	originalColorEnabled := colorEnabled
	colorEnabled = false

	defer func() {
		os.Stdout = original
		colorEnabled = originalColorEnabled
	}()

	f()

	_ = w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading captured output failed: %v", err)
	}

	return string(out)
}

// An IPv6 link-local address carries a zone ID, e.g. fe80::1%eth0. The
// probe line is built first and then printed, so handing it to a Printf as
// the format string turned that % into "%!e(MISSING)".
func TestColorPrinter_ZoneIDInTargetIsNotTreatedAsAVerb(t *testing.T) {
	p := NewColorPrinter(Config{})
	s := &stats.Statistics{
		Hostname:                "fe80::1%eth0",
		Port:                    80,
		Protocol:                "TCP",
		OngoingSuccessfulProbes: 1,
	}

	for _, tc := range []struct {
		name  string
		print func()
	}{
		{"success", func() { p.PrintProbeSuccess(s) }},
		{"failure", func() { p.PrintProbeFailure(s) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := captureStdout(t, tc.print)

			if strings.Contains(out, "%!") {
				t.Errorf("output has a formatting error in it: %q", out)
			}
			if !strings.Contains(out, "fe80::1%eth0") {
				t.Errorf("output does not contain the target verbatim: %q", out)
			}
		})
	}
}

// Without -I there is no source address to show, so -D has to leave the
// "using" part out rather than print an empty one.
func TestColorProbeWithoutASourceAddress(t *testing.T) {
	p := NewColorPrinter(Config{WithSourceAddress: true})
	s := &stats.Statistics{
		Hostname:                "example.com",
		Port:                    443,
		Protocol:                "TCP",
		OngoingSuccessfulProbes: 1,
	}

	for _, tc := range []struct {
		name  string
		print func()
	}{
		{"success", func() { p.PrintProbeSuccess(s) }},
		{"failure", func() { p.PrintProbeFailure(s) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if out := captureStdout(t, tc.print); strings.Contains(out, "using ") {
				t.Errorf("output = %q, want no source address", out)
			}
		})
	}
}

// Shutdown is what prints the summary on exit, so it has to print the same
// thing PrintStatistics does, unless --no-stats says otherwise.
func TestColorShutdownPrintsStatistics(t *testing.T) {
	s := plainTestStats()

	out := captureStdout(t, func() {
		NewColorPrinter(Config{}).Shutdown(s)
	})

	wantLines(t, out, "TCPing statistics ---\n")
}

func TestColorShutdownOmitsStatistics(t *testing.T) {
	s := plainTestStats()

	out := captureStdout(t, func() {
		NewColorPrinter(Config{OmitStatistics: true}).Shutdown(s)
	})

	if out != "" {
		t.Errorf("output = %q, want it to be empty", out)
	}
}

// captureColoredStdout runs f with color turned on and returns whatever it
// wrote to stdout, escape codes and all.
func captureColoredStdout(t *testing.T, f func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}

	original := os.Stdout
	os.Stdout = w
	originalColorEnabled := colorEnabled
	colorEnabled = true

	defer func() {
		os.Stdout = original
		colorEnabled = originalColorEnabled
	}()

	f()

	_ = w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading captured output failed: %v", err)
	}

	return string(out)
}

// The color a line comes out in is what tells the user at a glance whether a
// probe answered, so the ones that carry that meaning are checked here. The
// labels in the summary are decoration and are left alone on purpose.
func TestColorProbeLineColors(t *testing.T) {
	for _, tc := range []struct {
		name  string
		want  string
		print func(p *ColorPrinter, s *stats.Statistics)
	}{
		{"reply is light green", lightGreenCode, (*ColorPrinter).PrintProbeSuccess},
		{"no reply is red", redCode, (*ColorPrinter).PrintProbeFailure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := captureColoredStdout(t, func() {
				tc.print(NewColorPrinter(Config{}), plainTestStats())
			})

			if !strings.HasPrefix(out, tc.want) {
				t.Errorf("output = %q, want it to start with %q", out, tc.want)
			}
		})
	}
}

func TestColorPacketLossColors(t *testing.T) {
	for _, tc := range []struct {
		name         string
		successful   uint
		unsuccessful uint
		want         string
	}{
		{"no loss is green", 10, 0, greenCode + "0.00%"},
		{"some loss is light yellow", 9, 1, lightYellowCode + "10.00%"},
		{"heavy loss is red", 5, 5, redCode + "50.00%"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := plainTestStats()
			s.TotalSuccessfulProbes = tc.successful
			s.TotalUnsuccessfulProbes = tc.unsuccessful

			out := captureColoredStdout(t, func() {
				NewColorPrinter(Config{}).PrintStatistics(s)
			})

			wantLines(t, out, tc.want+resetCode)
		})
	}
}

func TestColorProbesThatNeverHappenedKeepTheirMeaning(t *testing.T) {
	out := captureColoredStdout(t, func() {
		NewColorPrinter(Config{}).PrintStatistics(plainTestStats())
	})

	wantLines(t,
		out,
		redCode+"Never succeeded\n"+resetCode,
		greenCode+"Never failed\n"+resetCode,
	)
}

func TestColorErrorAndRetryColors(t *testing.T) {
	p := NewColorPrinter(Config{})

	out := captureColoredStdout(t, func() { p.PrintError("could not connect") })
	wantLines(t, out, redCode+"could not connect\n"+resetCode)

	out = captureColoredStdout(t, func() { p.PrintRetryingToResolve("example.com") })
	wantLines(t, out, lightYellowCode+"Retrying to resolve example.com\n"+resetCode)
}
