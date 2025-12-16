package printers_test

import (
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/testdata"
	"github.com/pouriyajamshidi/tcping/v3/printers"
	"github.com/pouriyajamshidi/tcping/v3/statistics"
)


func TestNewJSONPrinter(t *testing.T) {
	p := printers.NewJSONPrinter()
	if p == nil {
		t.Fatal("NewJSONPrinter returned nil")
	}
}

func TestNewJSONPrinter_WithPrettyJSON(t *testing.T) {
	p := printers.NewJSONPrinter(printers.WithPrettyJSON())
	if p == nil {
		t.Fatal("NewJSONPrinter with pretty JSON returned nil")
	}
}

func TestJSONPrinter_PrintStart(t *testing.T) {
	stats := &statistics.Statistics{
		Hostname: "example.com",
		Port:     443,
	}

	data := testdata.CaptureJSONOutput(t, func() {
		p := printers.NewJSONPrinter()
		p.PrintStart(stats)
	})

	if data.Type != "start" {
		t.Errorf("expected type 'start', got %q", data.Type)
	}
	if data.Hostname != "example.com" {
		t.Errorf("expected hostname 'example.com', got %q", data.Hostname)
	}
	if data.Port != 443 {
		t.Errorf("expected port 443, got %d", data.Port)
	}
}

func TestJSONPrinter_PrintProbeSuccess(t *testing.T) {
	tests := []struct {
		name         string
		stats        *statistics.Statistics
		opts         []printers.JSONPrinterOption
		wantSuccess  *bool
		wantTimestamp bool
		wantSource   bool
		wantEmpty    bool // for failures-only mode
	}{
		{
			name: "basic success with hostname",
			stats: &statistics.Statistics{
				IP:                      netip.MustParseAddr("192.168.1.1"),
				Hostname:                "example.com",
				Port:                    443,
				OngoingSuccessfulProbes: 5,
				LatestRTT:               12.345,
			},
			wantSuccess: testdata.ToPtr(true),
		},
		{
			name: "success with IP only",
			stats: &statistics.Statistics{
				IP:                      netip.MustParseAddr("10.0.0.1"),
				Port:                    8080,
				OngoingSuccessfulProbes: 10,
				LatestRTT:               1.234,
			},
			wantSuccess: testdata.ToPtr(true),
		},
		{
			name: "with timestamp option",
			stats: &statistics.Statistics{
				IP:                      netip.MustParseAddr("192.168.1.1"),
				Hostname:                "test.local",
				Port:                    443,
				OngoingSuccessfulProbes: 3,
				LatestRTT:               8.901,
				StartTime:               time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
				LastSuccessfulProbe:     time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
			},
			opts: []printers.JSONPrinterOption{
				printers.WithTimestamp[*printers.JSONPrinter](),
			},
			wantSuccess:   testdata.ToPtr(true),
			wantTimestamp: true,
		},
		{
			name: "with source address option",
			stats: &statistics.Statistics{
				IP:                      netip.MustParseAddr("192.168.1.1"),
				Hostname:                "example.com",
				Port:                    443,
				OngoingSuccessfulProbes: 3,
				LatestRTT:               8.901,
				LocalAddr:               testdata.MockAddr{Addr: "10.0.0.1:12345"},
			},
			opts: []printers.JSONPrinterOption{
				printers.WithSourceAddress[*printers.JSONPrinter](),
			},
			wantSuccess: testdata.ToPtr(true),
			wantSource:  true,
		},
		{
			name: "failures only mode suppresses success",
			stats: &statistics.Statistics{
				IP:                      netip.MustParseAddr("192.168.1.1"),
				Hostname:                "example.com",
				Port:                    443,
				OngoingSuccessfulProbes: 5,
				LatestRTT:               12.345,
			},
			opts: []printers.JSONPrinterOption{
				printers.WithFailuresOnly[*printers.JSONPrinter](),
			},
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantEmpty {
				// capture raw output to verify nothing was printed
				output := testdata.CaptureOutput(t, func() {
					p := printers.NewJSONPrinter(tt.opts...)
					p.PrintProbeSuccess(tt.stats)
				})
				if output != "" {
					t.Errorf("expected no output in failures-only mode, got: %s", output)
				}
				return
			}

			data := testdata.CaptureJSONOutput(t, func() {
				p := printers.NewJSONPrinter(tt.opts...)
				p.PrintProbeSuccess(tt.stats)
			})

			if data.Type != "probe" {
				t.Errorf("expected type 'probe', got %q", data.Type)
			}
			if tt.wantSuccess != nil && (data.Success == nil || *data.Success != *tt.wantSuccess) {
				t.Errorf("expected success %v, got %v", *tt.wantSuccess, data.Success)
			}
			if tt.stats.Hostname != "" && data.Hostname != tt.stats.Hostname {
				t.Errorf("expected hostname %q, got %q", tt.stats.Hostname, data.Hostname)
			}
			if data.Port != tt.stats.Port {
				t.Errorf("expected port %d, got %d", tt.stats.Port, data.Port)
			}
			if data.Time == "" {
				t.Error("expected time to be set")
			}
			if tt.wantTimestamp && data.Timestamp == "" {
				t.Error("expected timestamp to be set")
			}
			if tt.wantSource && !strings.Contains(data.Message, "using") {
				t.Error("expected source address in message")
			}
		})
	}
}

func TestJSONPrinter_PrintProbeFailure(t *testing.T) {
	stats := &statistics.Statistics{
		IP:                        netip.MustParseAddr("192.168.1.1"),
		Hostname:                  "example.com",
		Port:                      443,
		OngoingUnsuccessfulProbes: 3,
	}

	data := testdata.CaptureJSONOutput(t, func() {
		p := printers.NewJSONPrinter()
		p.PrintProbeFailure(stats)
	})

	if data.Type != "probe" {
		t.Errorf("expected type 'probe', got %q", data.Type)
	}
	if data.Success == nil || *data.Success != false {
		t.Errorf("expected success false, got %v", data.Success)
	}
	if data.Hostname != "example.com" {
		t.Errorf("expected hostname 'example.com', got %q", data.Hostname)
	}
	if data.Port != 443 {
		t.Errorf("expected port 443, got %d", data.Port)
	}
}

func TestJSONPrinter_PrintStatistics(t *testing.T) {
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

	data := testdata.CaptureJSONOutput(t, func() {
		p := printers.NewJSONPrinter()
		p.PrintStatistics(stats)
	})

	if data.Type != "statistics" {
		t.Errorf("expected type 'statistics', got %q", data.Type)
	}
	if data.TotalSuccessfulPackets != 10 {
		t.Errorf("expected total successful packets 10, got %d", data.TotalSuccessfulPackets)
	}
	if data.TotalUnsuccessfulPackets != 2 {
		t.Errorf("expected total unsuccessful packets 2, got %d", data.TotalUnsuccessfulPackets)
	}
	if data.LatencyMin == "" {
		t.Error("expected latencyMin to be set")
	}
	if data.LatencyAvg == "" {
		t.Error("expected latencyAvg to be set")
	}
	if data.LatencyMax == "" {
		t.Error("expected latencyMax to be set")
	}
}

func TestJSONPrinter_Shutdown(t *testing.T) {
	p := printers.NewJSONPrinter()
	stats := &statistics.Statistics{}

	// smoke test - should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Shutdown panicked: %v", r)
		}
	}()

	p.Shutdown(stats)
}

