package printers

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/gookit/color"
	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
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
	// gookit/color caches the original stdout, so point it at the pipe too.
	color.SetOutput(w)

	defer func() {
		os.Stdout = original
		color.SetOutput(original)
	}()

	f()

	w.Close()
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
	p := NewColorPrinter(config.PrinterConfig{})
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
	p := NewColorPrinter(config.PrinterConfig{WithSourceAddress: true})
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
// thing PrintStatistics does, unless --omit-stats says otherwise.
func TestColorShutdownPrintsStatistics(t *testing.T) {
	s := plainTestStats()

	out := captureStdout(t, func() {
		NewColorPrinter(config.PrinterConfig{}).Shutdown(s)
	})

	wantLines(t, out, "TCPing statistics ---\n")
}

func TestColorShutdownOmitsStatistics(t *testing.T) {
	s := plainTestStats()

	out := captureStdout(t, func() {
		NewColorPrinter(config.PrinterConfig{OmitStatistics: true}).Shutdown(s)
	})

	if out != "" {
		t.Errorf("output = %q, want it to be empty", out)
	}
}
