package cli

import (
	"flag"
	"testing"
	"time"
)

func TestFlagsRequiringValue(t *testing.T) {
	oldCommandLine := flag.CommandLine
	defer func() {
		flag.CommandLine = oldCommandLine
	}()

	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	flag.Uint("c", 0, "")
	flag.Float64("t", 0, "")
	flag.Uint("r", 0, "")
	flag.Float64("i", 0, "")
	flag.Int("I", 0, "")
	flag.String("dns-server", "", "")
	flag.String("csv", "", "")
	flag.String("db", "", "")

	flag.Bool("4", false, "")
	flag.Bool("6", false, "")
	flag.Bool("D", false, "")
	flag.Bool("j", false, "")
	flag.Bool("pretty", false, "")
	flag.Bool("non-interactive", false, "")
	flag.Bool("no-color", false, "")
	flag.Bool("v", false, "")
	flag.Bool("u", false, "")

	fv := flagsRequiringValue()

	wantValue := []string{"c", "t", "r", "i", "I", "dns-server", "csv", "db"}
	for _, name := range wantValue {
		if !fv[name] {
			t.Errorf("expected %q to require a value", name)
		}
	}

	wantBool := []string{"4", "6", "D", "j", "pretty", "non-interactive", "no-color", "v", "u"}
	for _, name := range wantBool {
		if fv[name] {
			t.Errorf("expected %q to be a bool flag", name)
		}
	}
}

func TestSecondsToDuration(t *testing.T) {
	tests := []struct {
		name    string
		seconds float64
		want    time.Duration
	}{
		{
			name:    "whole second",
			seconds: 1,
			want:    time.Second,
		},
		{
			name:    "zero",
			seconds: 0,
			want:    0,
		},
		{
			name:    "fraction of a second",
			seconds: 0.5,
			want:    500 * time.Millisecond,
		},
		{
			name:    "fraction that is not exact in binary floating point",
			seconds: 0.3,
			want:    300 * time.Millisecond,
		},
		{
			name:    "more than a second",
			seconds: 2.5,
			want:    2500 * time.Millisecond,
		},
		{
			name:    "one millisecond",
			seconds: 0.001,
			want:    time.Millisecond,
		},
		{
			name:    "sub-millisecond is kept, not floored to zero",
			seconds: 0.0005,
			want:    500 * time.Microsecond,
		},
		{
			name:    "sub-millisecond part of a larger value is kept",
			seconds: 1.0005,
			want:    time.Second + 500*time.Microsecond,
		},
		{
			name:    "microsecond",
			seconds: 0.000001,
			want:    time.Microsecond,
		},
		{
			name:    "an hour",
			seconds: 3600,
			want:    time.Hour,
		},
		{
			name:    "negative",
			seconds: -1,
			want:    -time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := secondsToDuration(tt.seconds); got != tt.want {
				t.Errorf("secondsToDuration(%v) = %v, want %v", tt.seconds, got, tt.want)
			}
		})
	}
}
