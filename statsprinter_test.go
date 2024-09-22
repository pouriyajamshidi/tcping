package main

import (
	"testing"
	"time"
)

// dummyPrinter is a fake test implementation
// of a printer that does nothing.
type dummyPrinter struct{}

func (fp *dummyPrinter) printStart(_ string, _ uint16)                              {}
func (fp *dummyPrinter) printProbeFail(_, _ string, _ uint16, _ uint)               {}
func (fp *dummyPrinter) printRetryingToResolve(_ string)                            {}
func (fp *dummyPrinter) printTotalDownTime(_ time.Duration)                         {}
func (fp *dummyPrinter) printStatistics(_ tcping)                                   {}
func (fp *dummyPrinter) printVersion()                                              {}
func (fp *dummyPrinter) printInfo(_ string, _ ...interface{})                       {}
func (fp *dummyPrinter) printError(_ string, _ ...interface{})                      {}
func (fp *dummyPrinter) printProbeSuccess(_, _ string, _ uint16, _ uint, _ float32) {}

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
			name:     "0.5 milliseconds",
			duration: time.Duration(500 * time.Millisecond),
			want:     "0.5 milliseconds",
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
