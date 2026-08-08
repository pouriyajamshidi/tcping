package config

import (
	"testing"
)

func TestFlagsRequiringValue(t *testing.T) {
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
