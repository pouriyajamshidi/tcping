package printers_test

import (
	"net/netip"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/printers"
	"github.com/pouriyajamshidi/tcping/v3/statistics"
)

// note: ColorPrinter is functionally identical to PlainPrinter, just with ANSI color codes.
// comprehensive logic is tested in plain_test.go
// these are minimal smoke tests to ensure the printer doesn't panic

func TestNewColorPrinter(t *testing.T) {
	p := printers.NewColorPrinter()
	if p == nil {
		t.Fatal("NewColorPrinter returned nil")
	}
}

func TestColorPrinter_WithOptions(t *testing.T) {
	opts := []printers.ColorPrinterOption{
		printers.WithTimestamp[*printers.ColorPrinter](),
		printers.WithSourceAddress[*printers.ColorPrinter](),
		printers.WithFailuresOnly[*printers.ColorPrinter](),
	}
	p := printers.NewColorPrinter(opts...)
	if p == nil {
		t.Fatal("NewColorPrinter with options returned nil")
	}
}

func TestColorPrinter_PrintStart(t *testing.T) {
	p := printers.NewColorPrinter()
	stats := &statistics.Statistics{
		Hostname: "example.com",
		Port:     443,
	}

	// smoke test - just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintStart panicked: %v", r)
		}
	}()
	p.PrintStart(stats)
}

func TestColorPrinter_PrintProbeSuccess(t *testing.T) {
	p := printers.NewColorPrinter()
	stats := &statistics.Statistics{
		IP:                      netip.MustParseAddr("192.168.1.1"),
		Hostname:                "example.com",
		Port:                    443,
		OngoingSuccessfulProbes: 5,
		LatestRTT:               12.345,
	}

	// smoke test - just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintProbeSuccess panicked: %v", r)
		}
	}()
	p.PrintProbeSuccess(stats)
}

func TestColorPrinter_PrintProbeFailure(t *testing.T) {
	p := printers.NewColorPrinter()
	stats := &statistics.Statistics{
		IP:                        netip.MustParseAddr("192.168.1.1"),
		Hostname:                  "example.com",
		Port:                      443,
		OngoingUnsuccessfulProbes: 3,
	}

	// smoke test - just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintProbeFailure panicked: %v", r)
		}
	}()
	p.PrintProbeFailure(stats)
}

func TestColorPrinter_PrintTotalDownTime(t *testing.T) {
	p := printers.NewColorPrinter()
	stats := &statistics.Statistics{
		DownTime: 5 * time.Second,
	}

	// smoke test - just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintTotalDownTime panicked: %v", r)
		}
	}()
	p.PrintTotalDownTime(stats)
}

func TestColorPrinter_PrintRetryingToResolve(t *testing.T) {
	p := printers.NewColorPrinter()
	stats := &statistics.Statistics{
		Hostname: "flaky.example.com",
	}

	// smoke test - just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintRetryingToResolve panicked: %v", r)
		}
	}()
	p.PrintRetryingToResolve(stats)
}

func TestColorPrinter_PrintError(t *testing.T) {
	p := printers.NewColorPrinter()

	// smoke test - just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintError panicked: %v", r)
		}
	}()
	p.PrintError("connection timeout: %s", "deadline exceeded")
}

func TestColorPrinter_PrintStatistics(t *testing.T) {
	p := printers.NewColorPrinter()
	stats := &statistics.Statistics{
		IP:                        netip.MustParseAddr("192.168.1.1"),
		Hostname:                  "example.com",
		Port:                      443,
		TotalSuccessfulProbes:     10,
		TotalUnsuccessfulProbes:   2,
		LastSuccessfulProbe:       time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		LastUnsuccessfulProbe:     time.Date(2024, 1, 15, 12, 0, 30, 0, time.UTC),
		TotalUptime:               50 * time.Second,
		TotalDowntime:             10 * time.Second,
		StartTime:                 time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		EndTime:                   time.Date(2024, 1, 15, 12, 1, 0, 0, time.UTC),
		RTTResults:                statistics.RttResult{HasResults: true, Min: 10.5, Average: 15.2, Max: 20.8},
	}

	// smoke test - just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintStatistics panicked: %v", r)
		}
	}()
	p.PrintStatistics(stats)
}

func TestColorPrinter_Shutdown(t *testing.T) {
	p := printers.NewColorPrinter()
	stats := &statistics.Statistics{}

	// smoke test - just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Shutdown panicked: %v", r)
		}
	}()
	p.Shutdown(stats)
}
