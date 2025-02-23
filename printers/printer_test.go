package printers

import (
	"bytes"
	"fmt"
	"io"
	"net/netip"
	"os"
	"testing"
	"time"

	"github.com/gookit/color"
	"github.com/pouriyajamshidi/tcping/v2/consts"
	"github.com/pouriyajamshidi/tcping/v2/internal/utils"
	"github.com/pouriyajamshidi/tcping/v2/types"
	"github.com/stretchr/testify/assert"
)

// dummyPrinter is a fake test implementation
// of a printer that does nothing.
type dummyPrinter struct{}

func (fp *dummyPrinter) PrintStart(_ string, _ uint16)                                              {}
func (fp *dummyPrinter) PrintProbeSuccess(_ time.Time, _ string, _ types.Options, _ uint, _ string) {}
func (fp *dummyPrinter) PrintProbeFailure(_ time.Time, _ types.Options, _ uint)                     {}
func (fp *dummyPrinter) PrintRetryingToResolve(_ string)                                            {}
func (fp *dummyPrinter) PrintTotalDownTime(_ time.Duration)                                         {}
func (fp *dummyPrinter) PrintStatistics(_ types.Tcping)                                             {}
func (fp *dummyPrinter) PrintError(_ string, _ ...interface{})                                      {}

// createTestStats should be used to create new stats structs.
// it uses "127.0.0.1:12345" as default values, because
// [testServerListen] use the same values.
// It'll call t.Errorf if netip.ParseAddr has failed.
func createTestStats(t *testing.T) *types.Tcping {
	addr, err := netip.ParseAddr("127.0.0.1")

	s := types.Tcping{
		Printer: &dummyPrinter{},
		Options: types.Options{
			IP:                    addr,
			Port:                  12345,
			IntervalBetweenProbes: time.Second,
			Timeout:               time.Second,
		},
		Ticker: time.NewTicker(time.Second),
	}

	if err != nil {
		t.Errorf("ip parse: %v", err)
	}

	return &s
}

func TestDurationToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "1 second",
			duration: time.Duration(time.Second),
			want:     "1 second",
		},
		{
			name:     "59 seconds",
			duration: time.Duration(59 * time.Second),
			want:     "59 seconds",
		},
		{
			name:     "1 minute",
			duration: time.Duration(time.Minute),
			want:     "1 minute",
		},
		{
			name:     "59 minutes 0 seconds",
			duration: time.Duration(59 * time.Minute),
			want:     "59 minutes 0 seconds",
		},
		{
			name:     "1 minute 5 seconds",
			duration: time.Duration(time.Minute + 5*time.Second),
			want:     "1 minute 5 seconds",
		},
		{
			name:     "59 minutes 5 seconds",
			duration: time.Duration(59*time.Minute + 5*time.Second),
			want:     "59 minutes 5 seconds",
		},
		{
			name:     "1 hour",
			duration: time.Duration(time.Hour),
			want:     "1 hour",
		},
		{
			name:     "1 hour 10 minutes 5 seconds",
			duration: time.Duration(time.Hour + 10*time.Minute + 5*time.Second),
			want:     "1 hour 10 minutes 5 seconds",
		},
		{
			name:     "59 hours 0 minutes 0 seconds",
			duration: time.Duration(59 * time.Hour),
			want:     "59 hours 0 minutes 0 seconds",
		},
		{
			name:     "59 hours 10 minutes 5 seconds",
			duration: time.Duration(59*time.Hour + 10*time.Minute + 5*time.Second),
			want:     "59 hours 10 minutes 5 seconds",
		},
		{
			name:     "0.5 seconds",
			duration: time.Duration(500 * time.Millisecond),
			want:     "0.5 seconds",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.DurationToString(tt.duration); got != tt.want {
				t.Errorf("calcTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getProbeSuccessTests() []struct {
	name              string
	withTimestamp     bool
	useHostname       bool
	withSourceAddress bool
	expectedOutput    string
} {
	return []struct {
		name              string
		withTimestamp     bool
		useHostname       bool
		withSourceAddress bool
		expectedOutput    string
	}{
		{
			name:              "With hostname, no timestamp",
			withTimestamp:     false,
			useHostname:       true,
			withSourceAddress: false,
			expectedOutput:    "Reply from %s (%s) on port %d TCP_conn=%d time=%s ms\n",
		},
		{
			name:              "With hostname, with timestamp",
			withTimestamp:     true,
			useHostname:       true,
			withSourceAddress: false,
			expectedOutput:    "%s Reply from %s (%s) on port %d TCP_conn=%d time=%s ms\n",
		},
		{
			name:              "Without hostname, with timestamp",
			withTimestamp:     true,
			useHostname:       false,
			withSourceAddress: false,
			expectedOutput:    "%s Reply from %s on port %d TCP_conn=%d time=%s ms\n",
		},
		{
			name:              "Without hostname, no timestamp",
			withTimestamp:     false,
			useHostname:       false,
			withSourceAddress: false,
			expectedOutput:    "Reply from %s on port %d TCP_conn=%d time=%s ms\n",
		},
		{
			name:              "Without hostname, no timestamp, with show source address",
			withTimestamp:     false,
			useHostname:       false,
			withSourceAddress: true,
			expectedOutput:    "Reply from %s on port %d using %s TCP_conn=%d time=%s ms\n",
		},
		{
			name:              "With hostname, no timestamp, with show source address",
			withTimestamp:     false,
			useHostname:       true,
			withSourceAddress: true,
			expectedOutput:    "Reply from %s (%s) on port %d using %s TCP_conn=%d time=%s ms\n",
		},
	}
}

func TestPrintProbeSuccess(t *testing.T) {
	testCases := getProbeSuccessTests()
	stats := createTestStats(t)
	stats.Options.Hostname = "example.com"
	streak := uint(5)
	rtt := fmt.Sprintf("%.3f", 15.123)
	sourceAddr := "127.0.0.1:4567"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := PrinterConfig{}

			if !tc.useHostname {
				stats.Options.Hostname = ""
			} else {
				stats.Options.Hostname = "example.com"
			}

			if tc.withTimestamp {
				cfg.WithTimestamp = true
			}

			if tc.withSourceAddress {
				cfg.WithSourceAddress = true
			}

			read, write, _ := os.Pipe()
			os.Stdout = write

			p := NewPlainPrinter(cfg)
			p.PrintProbeSuccess(time.Now(), sourceAddr, stats.Options, streak, rtt)

			write.Close()

			var buf bytes.Buffer
			if _, err := io.Copy(&buf, read); err != nil {
				t.Fatalf("Failed to read from pipe: %v", err)
			}

			output := buf.String()

			var expected string

			if cfg.WithTimestamp {
				timestamp := time.Now().Format(consts.TimeFormat)
				if cfg.WithSourceAddress && tc.useHostname {
					expected = fmt.Sprintf(tc.expectedOutput, timestamp, stats.Options.Hostname, stats.Options.IP, stats.Options.Port, sourceAddr, streak, rtt)
				} else if cfg.WithSourceAddress {
					expected = fmt.Sprintf(tc.expectedOutput, timestamp, stats.Options.IP, stats.Options.Port, sourceAddr, streak, rtt)
				} else if tc.useHostname {
					expected = fmt.Sprintf(tc.expectedOutput, timestamp, stats.Options.Hostname, stats.Options.IP, stats.Options.Port, streak, rtt)
				} else {
					expected = fmt.Sprintf(tc.expectedOutput, timestamp, stats.Options.IP, stats.Options.Port, streak, rtt)
				}
			} else {
				if cfg.WithSourceAddress && tc.useHostname {
					expected = fmt.Sprintf(tc.expectedOutput, stats.Options.Hostname, stats.Options.IP, stats.Options.Port, sourceAddr, streak, rtt)
				} else if tc.withSourceAddress {
					expected = fmt.Sprintf(tc.expectedOutput, stats.Options.IP, stats.Options.Port, sourceAddr, streak, rtt)
				} else if tc.useHostname {
					expected = fmt.Sprintf(tc.expectedOutput, stats.Options.Hostname, stats.Options.IP, stats.Options.Port, streak, rtt)
				} else {
					expected = fmt.Sprintf(tc.expectedOutput, stats.Options.IP, stats.Options.Port, streak, rtt)
				}
			}

			assert.Equal(t, expected, output)
		})
	}
}

func TestPrintProbeFail(t *testing.T) {
	stats := createTestStats(t)
	stats.Options.Hostname = "example.com"

	streak := uint(5)

	testCases := []struct {
		name           string
		showTimestamp  bool
		useHostname    bool
		expectedOutput string
	}{
		{
			name:           "With hostname, no timestamp",
			showTimestamp:  false,
			useHostname:    true,
			expectedOutput: "No reply from %s (%s) on port %d TCP_conn=%d\n",
		},
		{
			name:           "With hostname, with timestamp",
			showTimestamp:  true,
			useHostname:    true,
			expectedOutput: "%s No reply from %s (%s) on port %d TCP_conn=%d\n",
		},
		{
			name:           "Without hostname, with timestamp",
			showTimestamp:  true,
			useHostname:    false,
			expectedOutput: "%s No reply from %s on port %d TCP_conn=%d\n",
		},
		{
			name:           "Without hostname, no timestamp",
			showTimestamp:  false,
			useHostname:    false,
			expectedOutput: "No reply from %s on port %d TCP_conn=%d\n",
		},
	}

	cfg := PrinterConfig{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg.WithTimestamp = tc.showTimestamp
			pp := NewPlainPrinter(cfg)

			read, write, _ := os.Pipe()
			os.Stdout = write
			color.SetOutput(write)

			if !tc.useHostname {
				stats.Options.Hostname = ""
			}

			pp.PrintProbeFailure(time.Now(), stats.Options, streak)

			write.Close()

			var buf bytes.Buffer
			if _, err := io.Copy(&buf, read); err != nil {
				t.Fatalf("Failed to read from pipe: %v", err)
			}

			output := buf.String()

			var expected string
			if tc.showTimestamp {
				timestamp := time.Now().Format(consts.TimeFormat)
				if tc.useHostname {
					expected = fmt.Sprintf(tc.expectedOutput, timestamp, stats.Options.Hostname, stats.Options.IP, stats.Options.Port, streak)
				} else {
					expected = fmt.Sprintf(tc.expectedOutput, timestamp, stats.Options.IP, stats.Options.Port, streak)
				}
			} else {
				if tc.useHostname {
					expected = fmt.Sprintf(tc.expectedOutput, stats.Options.Hostname, stats.Options.IP, stats.Options.Port, streak)
				} else {
					expected = fmt.Sprintf(tc.expectedOutput, stats.Options.IP, stats.Options.Port, streak)
				}
			}
			assert.Equal(t, expected, output)
		})
	}
}
