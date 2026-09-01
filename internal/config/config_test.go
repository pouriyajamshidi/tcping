package config

import (
	"flag"
	"testing"
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

// GetRetryResolveAfterNFailures must read the field that's actually
// populated from the -r flag, not an unrelated, never-set field.
func TestConfig_GetRetryResolveAfterNFailures(t *testing.T) {
	cfg := Config{RetryResolveAfterNFailures: 5}

	if got := cfg.GetRetryResolveAfterNFailures(); got != 5 {
		t.Errorf("GetRetryResolveAfterNFailures() = %d, want 5", got)
	}
}
