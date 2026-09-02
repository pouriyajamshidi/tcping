package printers

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/consts"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
)

func TestNewPrinter(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		cfg         config.PrinterConfig
		wantErr     bool
		expectedErr string
	}{
		{
			name: "JSON Printer Initialization",
			cfg: config.PrinterConfig{
				OutputJSON: true,
			},
			wantErr: false,
		},
		{
			name: "Database Printer Initialization",
			cfg: config.PrinterConfig{
				OutputDBPath: ":memory:",
				Target:       "example.com",
				Port:         443,
			},
			wantErr: false,
		},
		{
			name: "CSV Printer Initialization",
			cfg: config.PrinterConfig{
				OutputCSVPath: filepath.Join(tempDir, "test.csv"),
			},
			wantErr: false,
		},
		{
			name: "Alloy Printer Initialization",
			cfg: config.PrinterConfig{
				AlloyURL: "http://localhost:4318",
			},
			wantErr: false,
		},
		{
			name: "Plain Printer Initialization",
			cfg: config.PrinterConfig{
				NoColor: true,
			},
			wantErr: false,
		},
		{
			name: "Default Color Printer Initialization",
			cfg: config.PrinterConfig{
				NoColor: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printer, err := NewPrinter(tt.cfg)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewPrinter() expected an error, got nil")
				} else if err.Error() != tt.expectedErr {
					t.Errorf("NewPrinter() expected error %q, got %q", tt.expectedErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("NewPrinter() unexpected error: %v", err)
			}

			if printer == nil {
				t.Errorf("NewPrinter() returned nil printer for valid config")
			}

			if csvPrinter, ok := printer.(*CSVPrinter); ok {
				t.Cleanup(csvPrinter.done)
			}
		})
	}
}

func TestHTTPProbeSummary(t *testing.T) {
	tests := []struct {
		name string
		s    *stats.Statistics
		want string
	}{
		{
			name: "tcp probe says nothing",
			s:    &stats.Statistics{Protocol: consts.TCP, HTTP: stats.HTTPInfo{StatusCode: 200}},
			want: "",
		},
		{
			name: "http probe shows the status",
			s:    &stats.Statistics{Protocol: consts.HTTP, HTTP: stats.HTTPInfo{StatusCode: 200}},
			want: " status=200",
		},
		{
			name: "a probe that never got a response says nothing",
			s:    &stats.Statistics{Protocol: consts.HTTPS},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := httpProbeSummary(tt.s); got != tt.want {
				t.Errorf("httpProbeSummary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHTTPProbeDetails(t *testing.T) {
	https := func() *stats.Statistics {
		return &stats.Statistics{
			Protocol: consts.HTTPS,
			HTTP: stats.HTTPInfo{
				StatusCode:      200,
				Status:          "200 OK",
				Proto:           "HTTP/2.0",
				TLSVersion:      "TLS 1.3",
				TLSCipherSuite:  "TLS_AES_128_GCM_SHA256",
				CertExpiry:      time.Now().Add(48 * time.Hour),
				ConnectDuration: 10 * time.Millisecond,
				TLSDuration:     5 * time.Millisecond,
				TimeToFirstByte: 20 * time.Millisecond,
			},
		}
	}

	t.Run("off without verbose", func(t *testing.T) {
		s := https()

		if got := httpProbeDetails(s, false); got != "" {
			t.Errorf("httpProbeDetails() without -verbose = %q, want an empty string", got)
		}
	})

	t.Run("off for TCP", func(t *testing.T) {
		s := https()
		s.Protocol = consts.TCP

		if got := httpProbeDetails(s, true); got != "" {
			t.Errorf("httpProbeDetails() for a TCP probe = %q, want an empty string", got)
		}
	})

	t.Run("off without a response", func(t *testing.T) {
		s := https()
		s.HTTP.StatusCode = 0

		if got := httpProbeDetails(s, true); got != "" {
			t.Errorf("httpProbeDetails() without a response = %q, want an empty string", got)
		}
	})

	t.Run("https shows everything", func(t *testing.T) {
		got := httpProbeDetails(https(), true)

		for _, want := range []string{
			"HTTP/2.0 200 OK",
			"TLS 1.3 TLS_AES_128_GCM_SHA256",
			"certificate expires ",
			"connect=10.000 ms",
			"tls=5.000 ms",
			"ttfb=20.000 ms",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("httpProbeDetails() = %q, want it to contain %q", got, want)
			}
		}
	})

	t.Run("plain http leaves the TLS parts out", func(t *testing.T) {
		s := https()
		s.Protocol = consts.HTTP
		s.HTTP.Proto = "HTTP/1.1"
		s.HTTP.TLSVersion = ""
		s.HTTP.TLSCipherSuite = ""
		s.HTTP.CertExpiry = time.Time{}

		got := httpProbeDetails(s, true)

		for _, unwanted := range []string{"TLS", "certificate", "tls="} {
			if strings.Contains(got, unwanted) {
				t.Errorf("httpProbeDetails() = %q, want no %q for a plain HTTP probe", got, unwanted)
			}
		}

		if !strings.Contains(got, "connect=10.000 ms") || !strings.Contains(got, "ttfb=20.000 ms") {
			t.Errorf("httpProbeDetails() = %q, want the connect and first-byte timings", got)
		}
	})
}
