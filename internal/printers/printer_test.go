package printers

import (
	"reflect"
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
	tests := []struct {
		name        string
		cfg         PrinterConfig
		wantErr     bool
		expectedErr string
	}{
		{
			name: "Error on PrettyJSON without OutputJSON",
			cfg: PrinterConfig{
				OutputJSON: false,
				PrettyJSON: true,
			},
			wantErr:     true,
			expectedErr: "--pretty has no effect without the -j flag",
		},
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
				OutputCSVPath: "test.csv",
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
		})
	}
}

func TestCalcMinAvgMaxRttTime(t *testing.T) {
	tests := []struct {
		name     string
		input    []float32
		expected stats.RTTResult
	}{
		{
			name:  "Empty slice",
			input: []float32{},
			expected: stats.RTTResult{
				HasResults: false,
				Min:        0,
				Max:        0,
				Average:    0,
			},
		},
		{
			name:  "Single value",
			input: []float32{15.5},
			expected: stats.RTTResult{
				HasResults: true,
				Min:        15.5,
				Max:        15.5,
				Average:    15.5,
			},
		},
		{
			name:  "Multiple values",
			input: []float32{10.0, 20.0, 30.0},
			expected: stats.RTTResult{
				HasResults: true,
				Min:        10.0,
				Max:        30.0,
				Average:    20.0,
			},
		},
		{
			name:  "Values with decimals",
			input: []float32{1.1, 2.2, 3.3, 4.4},
			expected: stats.RTTResult{
				HasResults: true,
				Min:        1.1,
				Max:        4.4,
				Average:    2.75,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcMinAvgMaxRttTime(tt.input)

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("calcMinAvgMaxRttTime() = %+v, want %+v", got, tt.expected)
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
			RTT:                []float32{10.0, 20.0},
		}

		PrintStats(mp, s)

		if !mp.printStatisticsCalled {
			t.Errorf("Expected PrintStatistics to be called on the printer")
		}

		if mp.lastReceivedStats.LongestDowntime.Duration < 9*time.Second {
			t.Errorf("Expected LongestDowntime to be updated, got %v", mp.lastReceivedStats.LongestDowntime.Duration)
		}

		if mp.lastReceivedStats.RTTResults.HasResults == false {
			t.Errorf("Expected RTTResults to be populated")
		}
	})

	t.Run("Sets Longest Uptime on Success", func(t *testing.T) {
		mp := &mockPrinter{}

		startTime := time.Now().Add(-5 * time.Second)
		s := &stats.Statistics{
			LastProbeHadFailed: false,
			StartOfUptime:      startTime,
			RTT:                []float32{5.0},
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
