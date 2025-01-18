package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/gookit/color"
	"github.com/stretchr/testify/assert"
)

// dummyPrinter is a fake test implementation
// of a printer that does nothing.
type dummyPrinter struct{}

func (fp *dummyPrinter) printStart(_ string, _ uint16)                              {}
func (fp *dummyPrinter) printProbeSuccess(_ string, _ userInput, _ uint, _ float32) {}
func (fp *dummyPrinter) printProbeFail(_ userInput, _ uint)                         {}
func (fp *dummyPrinter) printRetryingToResolve(_ string)                            {}
func (fp *dummyPrinter) printTotalDownTime(_ time.Duration)                         {}
func (fp *dummyPrinter) printStatistics(_ tcping)                                   {}
func (fp *dummyPrinter) printVersion()                                              {}
func (fp *dummyPrinter) printInfo(_ string, _ ...interface{})                       {}
func (fp *dummyPrinter) printError(_ string, _ ...interface{})                      {}

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
			if got := durationToString(tt.duration); got != tt.want {
				t.Errorf("calcTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getProbeSuccessTests() []struct {
	name              string
	showTimestamp     bool
	useHostname       bool
	showSourceAddress bool
	expectedOutput    string
} {
	return []struct {
		name              string
		showTimestamp     bool
		useHostname       bool
		showSourceAddress bool
		expectedOutput    string
	}{
		{
			name:              "With hostname, no timestamp",
			showTimestamp:     false,
			useHostname:       true,
			showSourceAddress: false,
			expectedOutput:    "Reply from %s (%s) on port %d TCP_conn=%d time=%.3f ms\n",
		},
		{
			name:              "With hostname, with timestamp",
			showTimestamp:     true,
			useHostname:       true,
			showSourceAddress: false,
			expectedOutput:    "%s Reply from %s (%s) on port %d TCP_conn=%d time=%.3f ms\n",
		},
		{
			name:              "Without hostname, with timestamp",
			showTimestamp:     true,
			useHostname:       false,
			showSourceAddress: false,
			expectedOutput:    "%s Reply from %s on port %d TCP_conn=%d time=%.3f ms\n",
		},
		{
			name:              "Without hostname, no timestamp",
			showTimestamp:     false,
			useHostname:       false,
			showSourceAddress: false,
			expectedOutput:    "Reply from %s on port %d TCP_conn=%d time=%.3f ms\n",
		},
		{
			name:              "Without hostname, no timestamp, with show source address",
			showTimestamp:     false,
			useHostname:       false,
			showSourceAddress: true,
			expectedOutput:    "Reply from %s on port %d using %s TCP_conn=%d time=%.3f ms\n",
		},
		{
			name:              "With hostname, no timestamp, with show source address",
			showTimestamp:     false,
			useHostname:       true,
			showSourceAddress: true,
			expectedOutput:    "Reply from %s (%s) on port %d using %s TCP_conn=%d time=%.3f ms\n",
		},
	}
}

func TestPrintProbeSuccess(t *testing.T) {
	testCases := getProbeSuccessTests()
	stats := createTestStats(t)
	stats.userInput.hostname = "example.com"
	streak := uint(5)
	rtt := float32(15.123)
	sourceAddr := "127.0.0.1:4567"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pp := newPlainPrinter(&tc.showTimestamp)

			read, write, _ := os.Pipe()
			os.Stdout = write

			if !tc.useHostname {
				stats.userInput.hostname = ""
			} else {
				stats.userInput.hostname = "example.com"
			}

			if tc.showSourceAddress {
				stats.userInput.showSourceAddress = true
			}

			pp.printProbeSuccess(sourceAddr, stats.userInput, streak, rtt)

			write.Close()

			var buf bytes.Buffer
			if _, err := io.Copy(&buf, read); err != nil {
				t.Fatalf("Failed to read from pipe: %v", err)
			}

			output := buf.String()

			var expected string
			if tc.showTimestamp {
				timestamp := time.Now().Format("2006-01-02 15:04:05")
				if tc.showSourceAddress && tc.useHostname {
					expected = fmt.Sprintf(tc.expectedOutput, timestamp, stats.userInput.hostname, stats.userInput.ip, stats.userInput.port, sourceAddr, streak, rtt)
				} else if tc.showSourceAddress {
					expected = fmt.Sprintf(tc.expectedOutput, timestamp, stats.userInput.ip, stats.userInput.port, sourceAddr, streak, rtt)
				} else if tc.useHostname {
					expected = fmt.Sprintf(tc.expectedOutput, timestamp, stats.userInput.hostname, stats.userInput.ip, stats.userInput.port, streak, rtt)
				} else {
					expected = fmt.Sprintf(tc.expectedOutput, timestamp, stats.userInput.ip, stats.userInput.port, streak, rtt)
				}
			} else {
				if tc.showSourceAddress && tc.useHostname {
					expected = fmt.Sprintf(tc.expectedOutput, stats.userInput.hostname, stats.userInput.ip, stats.userInput.port, sourceAddr, streak, rtt)
				} else if tc.showSourceAddress {
					expected = fmt.Sprintf(tc.expectedOutput, stats.userInput.ip, stats.userInput.port, sourceAddr, streak, rtt)
				} else if tc.useHostname {
					expected = fmt.Sprintf(tc.expectedOutput, stats.userInput.hostname, stats.userInput.ip, stats.userInput.port, streak, rtt)
				} else {
					expected = fmt.Sprintf(tc.expectedOutput, stats.userInput.ip, stats.userInput.port, streak, rtt)
				}
			}

			assert.Equal(t, expected, output)
		})
	}
}

func TestPrintProbeFail(t *testing.T) {
	stats := createTestStats(t)
	stats.userInput.hostname = "example.com"

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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pp := newColorPrinter(&tc.showTimestamp)

			read, write, _ := os.Pipe()
			os.Stdout = write
			color.SetOutput(write)
			color.Disable()

			if !tc.useHostname {
				stats.userInput.hostname = ""
			}

			pp.printProbeFail(stats.userInput, streak)

			write.Close()

			var buf bytes.Buffer
			if _, err := io.Copy(&buf, read); err != nil {
				t.Fatalf("Failed to read from pipe: %v", err)
			}

			output := buf.String()

			var expected string
			if tc.showTimestamp {
				timestamp := time.Now().Format("2006-01-02 15:04:05")
				if tc.useHostname {
					expected = fmt.Sprintf(tc.expectedOutput, timestamp, stats.userInput.hostname, stats.userInput.ip, stats.userInput.port, streak)
				} else {
					expected = fmt.Sprintf(tc.expectedOutput, timestamp, stats.userInput.ip, stats.userInput.port, streak)
				}
			} else {
				if tc.useHostname {
					expected = fmt.Sprintf(tc.expectedOutput, stats.userInput.hostname, stats.userInput.ip, stats.userInput.port, streak)
				} else {
					expected = fmt.Sprintf(tc.expectedOutput, stats.userInput.ip, stats.userInput.port, streak)
				}
			}
			assert.Equal(t, expected, output)
		})
	}
}
