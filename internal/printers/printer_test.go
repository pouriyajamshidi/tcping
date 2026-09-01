package printers

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
)

// mockPrinter implements the Printer interface for testing orchestration.
type mockPrinter struct {
	printStatisticsCalled bool
	lastReceivedStats     *stats.Statistics
}

func (m *mockPrinter) PrintStart(_ *stats.Statistics)            {}
func (m *mockPrinter) PrintProbeSuccess(_ *stats.Statistics)     {}
func (m *mockPrinter) PrintProbeFailure(_ *stats.Statistics)     {}
func (m *mockPrinter) PrintRetryingToResolve(_ string)           {}
func (m *mockPrinter) PrintDownTimeDuration(_ *stats.Statistics) {}
func (m *mockPrinter) PrintError(_ string, _ ...any)             {}
func (m *mockPrinter) Shutdown(_ *stats.Statistics)              {}

// PrintStatistics records that the method was called and saves the stats state.
func (m *mockPrinter) PrintStatistics(s *stats.Statistics) {
	m.printStatisticsCalled = true
	m.lastReceivedStats = s
}

func TestNewPrinter(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		cfg         PrinterConfig
		wantErr     bool
		expectedErr string
	}{
		{
			name: "JSON Printer Initialization",
			cfg: PrinterConfig{
				OutputJSON: true,
			},
			wantErr: false,
		},
		{
			name: "Database Printer Initialization",
			cfg: PrinterConfig{
				OutputDBPath: ":memory:",
				Target:       "example.com",
				Port:         443,
			},
			wantErr: false,
		},
		{
			name: "CSV Printer Initialization",
			cfg: PrinterConfig{
				OutputCSVPath: filepath.Join(tempDir, "test.csv"),
			},
			wantErr: false,
		},
		{
			name: "Plain Printer Initialization",
			cfg: PrinterConfig{
				NoColor: true,
			},
			wantErr: false,
		},
		{
			name: "Default Color Printer Initialization",
			cfg: PrinterConfig{
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
				t.Cleanup(csvPrinter.Done)
			}
		})
	}
}

func TestPrintStats(t *testing.T) {
	t.Run("Sets Longest Downtime on Failure", func(t *testing.T) {
		mp := &mockPrinter{}

		startTime := time.Now().Add(-10 * time.Second)
		s := &stats.Statistics{
			LastProbeHadFailed: true,
			StartOfDowntime:    startTime,
		}

		PrintStats(mp, s)

		if !mp.printStatisticsCalled {
			t.Errorf("Expected PrintStatistics to be called on the printer")
		}

		if mp.lastReceivedStats.LongestDowntime.Duration < 9*time.Second {
			t.Errorf("Expected LongestDowntime to be updated, got %v", mp.lastReceivedStats.LongestDowntime.Duration)
		}
	})

	t.Run("Sets Longest Uptime on Success", func(t *testing.T) {
		mp := &mockPrinter{}

		startTime := time.Now().Add(-5 * time.Second)
		s := &stats.Statistics{
			LastProbeHadFailed: false,
			StartOfUptime:      startTime,
		}

		PrintStats(mp, s)

		if !mp.printStatisticsCalled {
			t.Errorf("Expected PrintStatistics to be called on the printer")
		}

		if mp.lastReceivedStats.LongestUptime.Duration < 4*time.Second {
			t.Errorf("Expected LongestUptime to be updated, got %v", mp.lastReceivedStats.LongestUptime.Duration)
		}
	})
}
