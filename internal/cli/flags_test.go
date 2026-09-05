package cli

import (
	"flag"
	"os"
	"slices"
	"strings"
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

// useTestFlagSet swaps flag.CommandLine for a fresh set holding a handful of
// tcping's real flags, so permuteArgs can tell which of them take a value.
// The original set is put back when the test ends.
func useTestFlagSet(t *testing.T) {
	t.Helper()

	oldCommandLine := flag.CommandLine
	t.Cleanup(func() {
		flag.CommandLine = oldCommandLine
	})

	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	flag.Uint("c", 0, "")
	flag.Float64("i", 0, "")
	flag.Float64("t", 0, "")
	flag.String("dns-server", "", "")
	flag.String("csv", "", "")

	flag.Bool("4", false, "")
	flag.Bool("6", false, "")
	flag.Bool("D", false, "")
	flag.Bool("v", false, "")
}

func TestPermuteArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "flags after the target are moved in front of it",
			args: []string{"example.com", "443", "-4", "-c", "5"},
			want: []string{"-4", "-c", "5", "example.com", "443"},
		},
		{
			name: "flags already in front are left where they are",
			args: []string{"-4", "-c", "5", "example.com", "443"},
			want: []string{"-4", "-c", "5", "example.com", "443"},
		},
		{
			name: "flags on both sides of the target are gathered up",
			args: []string{"-4", "example.com", "-c", "5", "443"},
			want: []string{"-4", "-c", "5", "example.com", "443"},
		},
		{
			name: "the order of the flags themselves is kept",
			args: []string{"example.com", "443", "-c", "5", "-i", "0.5", "-D"},
			want: []string{"-c", "5", "-i", "0.5", "-D", "example.com", "443"},
		},
		{
			name: "the order of the positional arguments is kept",
			args: []string{"-4", "example.com", "443"},
			want: []string{"-4", "example.com", "443"},
		},
		{
			name: "nothing but positional arguments",
			args: []string{"example.com", "443"},
			want: []string{"example.com", "443"},
		},
		{
			name: "nothing but flags",
			args: []string{"-4", "-D"},
			want: []string{"-4", "-D"},
		},
		{
			name: "no arguments at all",
			args: []string{},
			want: []string{},
		},
		{
			name: "the host:port form is a single positional argument",
			args: []string{"example.com:443", "-c", "5"},
			want: []string{"-c", "5", "example.com:443"},
		},
		{
			name: "a long flag written with two dashes still takes its value",
			args: []string{"example.com", "53", "--dns-server", "1.1.1.1"},
			want: []string{"--dns-server", "1.1.1.1", "example.com", "53"},
		},
		{
			name: "a bool flag does not swallow the argument after it",
			args: []string{"-4", "example.com", "-D", "443"},
			want: []string{"-4", "-D", "example.com", "443"},
		},
		{
			name: "-c=5 is one argument, so it is left for flag.Parse to split",
			args: []string{"example.com", "443", "-c=5"},
			want: []string{"-c=5", "example.com", "443"},
		},
		{
			name: "an unknown flag is passed through for flag.Parse to reject",
			args: []string{"example.com", "443", "-nope"},
			want: []string{"-nope", "example.com", "443"},
		},
		{
			name: "a value that looks like a hostname is still a value",
			args: []string{"-csv", "out.csv", "example.com", "443"},
			want: []string{"-csv", "out.csv", "example.com", "443"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useTestFlagSet(t)

			args := slices.Clone(tt.args)
			permuteArgs(args)

			if !slices.Equal(args, tt.want) {
				t.Errorf("permuteArgs(%q) = %q, want %q", tt.args, args, tt.want)
			}
		})
	}
}

func TestPermuteArgsRejectsAFlagWithoutItsValue(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "nothing follows the flag",
			args: []string{"example.com", "443", "-c"},
		},
		{
			name: "another flag follows it",
			args: []string{"-c", "-4", "example.com", "443"},
		},
		{
			name: "a long flag with nothing after it",
			args: []string{"example.com", "53", "--dns-server"},
		},
	}

	// The child process runs the one case it was asked for and lets
	// permuteArgs end the program.
	if name := os.Getenv(subprocessCase); name != "" {
		for _, tt := range tests {
			if tt.name == name {
				useTestFlagSet(t)
				permuteArgs(tt.args)
				t.Fatalf("permuteArgs(%q) returned instead of exiting", tt.args)
			}
		}
		t.Fatalf("unknown case %q", name)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, code := runCase(t, "TestPermuteArgsRejectsAFlagWithoutItsValue", tt.name)

			if code == 0 {
				t.Errorf("permuteArgs(%q) exited with 0, want a failure. Output:\n%s", tt.args, out)
			}

			if !strings.Contains(out, "option requires a value") {
				t.Errorf("permuteArgs(%q) did not say the option requires a value. Output:\n%s", tt.args, out)
			}
		})
	}
}

func TestValidateRejectsFlagsThatDoNotGoTogether(t *testing.T) {
	tests := []struct {
		name  string
		flags flags
		want  string
	}{
		{
			name:  "both IP versions at once",
			flags: flags{useIPv4: true, useIPv6: true, intervalBetweenProbes: 1},
			want:  "Only one IP version can be specified",
		},
		{
			name:  "pretty without JSON",
			flags: flags{prettyJSON: true, intervalBetweenProbes: 1},
			want:  "--pretty has no effect without the -j flag",
		},
		{
			name:  "omitting statistics that go to a database",
			flags: flags{omitStatistics: true, DBPath: "out.db", intervalBetweenProbes: 1},
			want:  "--omit-stats has no effect",
		},
		{
			name:  "omitting statistics that go to a CSV file",
			flags: flags{omitStatistics: true, CSVPath: "out.csv", intervalBetweenProbes: 1},
			want:  "--omit-stats has no effect",
		},
		{
			name:  "omitting statistics that go to Alloy",
			flags: flags{omitStatistics: true, alloyURL: "http://localhost:4318", alloyStatsInterval: 10, intervalBetweenProbes: 1},
			want:  "--omit-stats has no effect",
		},
		{
			name:  "omitting statistics that go to InfluxDB",
			flags: flags{omitStatistics: true, influxDBURL: "http://localhost:8086", influxDBStatsInterval: 10, intervalBetweenProbes: 1},
			want:  "--omit-stats has no effect",
		},
		{
			name:  "a zero probe interval, which would panic the ticker",
			flags: flags{intervalBetweenProbes: 0},
			want:  "Interval between probes should be more than 0 seconds",
		},
		{
			name:  "a negative probe interval",
			flags: flags{intervalBetweenProbes: -1},
			want:  "Interval between probes should be more than 0 seconds",
		},
		{
			name:  "a zero Alloy statistics interval",
			flags: flags{alloyURL: "http://localhost:4318", alloyStatsInterval: 0, intervalBetweenProbes: 1},
			want:  "Alloy statistics interval should be more than 0 seconds",
		},
		{
			name:  "a zero InfluxDB statistics interval",
			flags: flags{influxDBURL: "http://localhost:8086", influxDBStatsInterval: 0, intervalBetweenProbes: 1},
			want:  "InfluxDB statistics interval should be more than 0 seconds",
		},
	}

	// The child process runs the one case it was asked for and lets
	// validate end the program.
	if name := os.Getenv(subprocessCase); name != "" {
		for _, tt := range tests {
			if tt.name == name {
				tt.flags.validate()
				t.Fatalf("validate returned instead of exiting for %q", name)
			}
		}
		t.Fatalf("unknown case %q", name)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, code := runCase(t, "TestValidateRejectsFlagsThatDoNotGoTogether", tt.name)

			if code == 0 {
				t.Errorf("validate exited with 0, want a failure. Output:\n%s", out)
			}

			if !strings.Contains(out, tt.want) {
				t.Errorf("validate did not print %q. Output:\n%s", tt.want, out)
			}
		})
	}
}

func TestValidateAcceptsFlagsThatGoTogether(t *testing.T) {
	tests := []struct {
		name  string
		flags flags
	}{
		{
			name:  "nothing but a probe interval",
			flags: flags{intervalBetweenProbes: 1},
		},
		{
			name:  "a fraction of a second between probes",
			flags: flags{intervalBetweenProbes: 0.1},
		},
		{
			name:  "one IP version",
			flags: flags{useIPv4: true, intervalBetweenProbes: 1},
		},
		{
			name:  "pretty alongside JSON",
			flags: flags{outputJSON: true, prettyJSON: true, intervalBetweenProbes: 1},
		},
		{
			name:  "omitting statistics on terminal output",
			flags: flags{omitStatistics: true, intervalBetweenProbes: 1},
		},
		{
			name:  "an Alloy statistics interval of zero without an Alloy URL",
			flags: flags{alloyStatsInterval: 0, intervalBetweenProbes: 1},
		},
		{
			name:  "an InfluxDB statistics interval of zero without an InfluxDB URL",
			flags: flags{influxDBStatsInterval: 0, intervalBetweenProbes: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// validate exits on anything it does not like, so simply
			// returning is the whole assertion.
			tt.flags.validate()
		})
	}
}
