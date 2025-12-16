package printers_test

import (
	"bytes"
	"io"
	"net"
	"net/netip"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/printers"
	"github.com/pouriyajamshidi/tcping/v3/statistics"
)

// mockAddr implements net.Addr for testing
type mockAddr struct {
	addr string
}

func (m mockAddr) Network() string { return "tcp" }
func (m mockAddr) String() string  { return m.addr }

var _ net.Addr = (*mockAddr)(nil)

// captureOutput captures stdout during function execution
func captureOutput(fn func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	w.Close()
	output := <-done
	os.Stdout = oldStdout

	return output
}

func TestNewPlainPrinter(t *testing.T) {
	p := printers.NewPlainPrinter()
	if p == nil {
		t.Fatal("NewPlainPrinter returned nil")
	}
}

func TestPlainPrinter_PrintStart(t *testing.T) {
	p := printers.NewPlainPrinter()
	stats := &statistics.Statistics{
		Hostname: "example.com",
		Port:     443,
	}

	output := captureOutput(func() {
		p.PrintStart(stats)
	})

	if !strings.Contains(output, "TCPinging example.com") {
		t.Errorf("expected 'TCPinging example.com', got: %q", output)
	}
	if !strings.Contains(output, "443") {
		t.Errorf("expected port 443, got: %q", output)
	}
}

func TestPlainPrinter_PrintProbeSuccess(t *testing.T) {
	tests := []struct {
		name            string
		stats           *statistics.Statistics
		opts            []printers.PlainPrinterOption
		wantInOutput    []string
		wantNotInOutput []string
	}{
		{
			name: "basic success with hostname",
			stats: &statistics.Statistics{
				IP:                      netip.MustParseAddr("192.168.1.1"),
				Port:                    443,
				Hostname:                "example.com",
				OngoingSuccessfulProbes: 5,
				LatestRTT:               12.345,
			},
			wantInOutput: []string{
				"Reply from example.com",
				"192.168.1.1",
				"443",
				"TCP_conn=5",
				"12.345",
			},
		},
		{
			name: "success with IP only (no hostname)",
			stats: &statistics.Statistics{
				IP:                      netip.MustParseAddr("192.168.1.1"),
				Port:                    80,
				Hostname:                "",
				OngoingSuccessfulProbes: 1,
				LatestRTT:               5.678,
			},
			wantInOutput: []string{
				"Reply from 192.168.1.1",
				"port 80",
				"TCP_conn=1",
				"5.678",
			},
			wantNotInOutput: []string{
				"(192.168.1.1)", // should not show IP in parens when no hostname
			},
		},
		{
			name: "success with timestamp",
			stats: &statistics.Statistics{
				IP:                      netip.MustParseAddr("10.0.0.1"),
				Port:                    8080,
				Hostname:                "test.local",
				StartTime:               time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
				OngoingSuccessfulProbes: 10,
				LatestRTT:               1.234,
			},
			opts: []printers.PlainPrinterOption{
				printers.WithTimestamp[*printers.PlainPrinter](),
			},
			wantInOutput: []string{
				"2024-01-15",
				"10:30:45",
				"Reply from test.local",
			},
		},
		{
			name: "success with source address",
			stats: &statistics.Statistics{
				IP:                      netip.MustParseAddr("192.168.1.1"),
				Port:                    443,
				Hostname:                "example.com",
				LocalAddr:               mockAddr{addr: "10.0.0.1:12345"},
				OngoingSuccessfulProbes: 3,
				LatestRTT:               8.901,
			},
			opts: []printers.PlainPrinterOption{
				printers.WithSourceAddress[*printers.PlainPrinter](),
			},
			wantInOutput: []string{
				"using 10.0.0.1",
			},
		},
		{
			name: "failures only mode suppresses success",
			stats: &statistics.Statistics{
				IP:                      netip.MustParseAddr("192.168.1.1"),
				Port:                    443,
				Hostname:                "example.com",
				OngoingSuccessfulProbes: 5,
				LatestRTT:               12.345,
			},
			opts: []printers.PlainPrinterOption{
				printers.WithFailuresOnly[*printers.PlainPrinter](),
			},
			wantNotInOutput: []string{
				"Reply from",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printers.NewPlainPrinter(tt.opts...)

			output := captureOutput(func() {
				p.PrintProbeSuccess(tt.stats)
			})

			for _, want := range tt.wantInOutput {
				if !strings.Contains(output, want) {
					t.Errorf("expected output to contain %q, got: %s", want, output)
				}
			}

			for _, notWant := range tt.wantNotInOutput {
				if strings.Contains(output, notWant) {
					t.Errorf("expected output to NOT contain %q, got: %s", notWant, output)
				}
			}
		})
	}
}

func TestPlainPrinter_PrintProbeFailure(t *testing.T) {
	tests := []struct {
		name         string
		stats        *statistics.Statistics
		opts         []printers.PlainPrinterOption
		wantInOutput []string
	}{
		{
			name: "basic failure with hostname",
			stats: &statistics.Statistics{
				IP:                        netip.MustParseAddr("192.168.1.1"),
				Port:                      443,
				Hostname:                  "example.com",
				OngoingUnsuccessfulProbes: 3,
			},
			wantInOutput: []string{
				"No reply from example.com",
				"192.168.1.1",
				"443",
				"TCP_conn=3",
			},
		},
		{
			name: "failure with IP only",
			stats: &statistics.Statistics{
				IP:                        netip.MustParseAddr("10.0.0.1"),
				Port:                      80,
				Hostname:                  "",
				OngoingUnsuccessfulProbes: 1,
			},
			wantInOutput: []string{
				"No reply from 10.0.0.1",
				"port 80",
				"TCP_conn=1",
			},
		},
		{
			name: "failure with timestamp",
			stats: &statistics.Statistics{
				IP:                        netip.MustParseAddr("192.168.1.1"),
				Port:                      443,
				Hostname:                  "test.local",
				StartTime:                 time.Date(2024, 1, 15, 14, 22, 10, 0, time.UTC),
				OngoingUnsuccessfulProbes: 5,
			},
			opts: []printers.PlainPrinterOption{
				printers.WithTimestamp[*printers.PlainPrinter](),
			},
			wantInOutput: []string{
				"2024-01-15",
				"14:22:10",
				"No reply from test.local",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printers.NewPlainPrinter(tt.opts...)

			output := captureOutput(func() {
				p.PrintProbeFailure(tt.stats)
			})

			for _, want := range tt.wantInOutput {
				if !strings.Contains(output, want) {
					t.Errorf("expected output to contain %q, got: %s", want, output)
				}
			}
		})
	}
}

func TestPlainPrinter_PrintTotalDownTime(t *testing.T) {
	p := printers.NewPlainPrinter()
	stats := &statistics.Statistics{
		DownTime: 5 * time.Second,
	}

	output := captureOutput(func() {
		p.PrintTotalDownTime(stats)
	})

	if !strings.Contains(output, "No response received") {
		t.Errorf("expected 'No response received', got: %s", output)
	}
	if !strings.Contains(output, "5 second") {
		t.Errorf("expected '5 second', got: %s", output)
	}
}

func TestPlainPrinter_PrintRetryingToResolve(t *testing.T) {
	p := printers.NewPlainPrinter()
	stats := &statistics.Statistics{
		Hostname: "flaky.example.com",
	}

	output := captureOutput(func() {
		p.PrintRetryingToResolve(stats)
	})

	if !strings.Contains(output, "Retrying to resolve") {
		t.Errorf("expected 'Retrying to resolve', got: %s", output)
	}
	if !strings.Contains(output, "flaky.example.com") {
		t.Errorf("expected hostname, got: %s", output)
	}
}

func TestPlainPrinter_PrintError(t *testing.T) {
	p := printers.NewPlainPrinter()

	output := captureOutput(func() {
		p.PrintError("connection timeout: %s", "deadline exceeded")
	})

	if !strings.Contains(output, "connection timeout") {
		t.Errorf("expected error message, got: %s", output)
	}
	if !strings.Contains(output, "deadline exceeded") {
		t.Errorf("expected formatted arg, got: %s", output)
	}
}

func TestPlainPrinter_PrintStatistics(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		stats        *statistics.Statistics
		wantInOutput []string
	}{
		{
			name: "basic statistics with hostname",
			stats: &statistics.Statistics{
				IP:                        netip.MustParseAddr("192.168.1.1"),
				Port:                      443,
				Hostname:                  "example.com",
				DestIsIP:                  false,
				TotalSuccessfulProbes:     10,
				TotalUnsuccessfulProbes:   2,
				StartTime:                 now,
				EndTime:                   now.Add(60 * time.Second),
				TotalUptime:               50 * time.Second,
				TotalDowntime:             10 * time.Second,
				LastSuccessfulProbe:       now,
				LastUnsuccessfulProbe:     now.Add(30 * time.Second),
				RTTResults: statistics.RttResult{
					HasResults: true,
					Min:        10.5,
					Average:    15.2,
					Max:        20.8,
				},
			},
			wantInOutput: []string{
				"example.com (192.168.1.1) TCPing statistics",
				"12 probes transmitted on port 443",
				"10 received",
				"16.67% packet loss",
				"successful probes:   10",
				"unsuccessful probes: 2",
				"rtt min/avg/max",
				"10.500",
				"15.200",
				"20.800",
				"TCPing started at:",
				"TCPing ended at:",
			},
		},
		{
			name: "IP only statistics",
			stats: &statistics.Statistics{
				IP:                      netip.MustParseAddr("10.0.0.1"),
				Port:                    80,
				Hostname:                "10.0.0.1",
				DestIsIP:                true,
				TotalSuccessfulProbes:   5,
				TotalUnsuccessfulProbes: 0,
				StartTime:               now,
				TotalUptime:             30 * time.Second,
				LastSuccessfulProbe:     now,
			},
			wantInOutput: []string{
				"10.0.0.1 TCPing statistics",
				"5 probes transmitted",
				"5 received",
				"0.00% packet loss",
				"Never failed",
			},
		},
		{
			name: "zero packet loss",
			stats: &statistics.Statistics{
				IP:                      netip.MustParseAddr("192.168.1.1"),
				Port:                    443,
				Hostname:                "example.com",
				TotalSuccessfulProbes:   100,
				TotalUnsuccessfulProbes: 0,
				StartTime:               now,
			},
			wantInOutput: []string{
				"0.00% packet loss",
			},
		},
		{
			name: "statistics without RTT",
			stats: &statistics.Statistics{
				IP:                    netip.MustParseAddr("192.168.1.1"),
				Port:                  443,
				Hostname:              "example.com",
				TotalSuccessfulProbes: 5,
				StartTime:             now,
				RTTResults: statistics.RttResult{
					HasResults: false,
				},
			},
			wantInOutput: []string{
				"example.com",
				"5 probes transmitted",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printers.NewPlainPrinter()

			output := captureOutput(func() {
				p.PrintStatistics(tt.stats)
			})

			for _, want := range tt.wantInOutput {
				if !strings.Contains(output, want) {
					t.Errorf("expected output to contain %q, got: %s", want, output)
				}
			}
		})
	}
}

func TestPlainPrinter_Shutdown(t *testing.T) {
	p := printers.NewPlainPrinter()
	stats := &statistics.Statistics{}

	// shutdown should not panic and should not modify stats
	p.Shutdown(stats)

	if !stats.EndTime.IsZero() {
		t.Error("Shutdown should not modify statistics")
	}
}
