package cli

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestProcessUserInputSpotsIPLiterals(t *testing.T) {
	tests := []struct {
		target string
		want   bool
	}{
		{target: "127.0.0.1", want: true},
		{target: "::1", want: true},
		{target: "0:0:0:0:0:0:0:1", want: true},
		{target: "2001:0db8::1", want: true},
		{target: "localhost", want: false},
	}

	// ProcessUserInput registers its flags on the global flag set, which
	// panics when done twice, so each case runs in its own process.
	if target := os.Getenv(subprocessCase); target != "" {
		os.Args = []string{"tcping", target, "443"}
		cfg, _ := ProcessUserInput()
		fmt.Printf("TargetIsIP=%v\n", cfg.TargetIsIP)
		os.Exit(0)
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			out, code := runCase(t, "TestProcessUserInputSpotsIPLiterals", tt.target)
			if code != 0 {
				t.Fatalf("exited with %d. Output:\n%s", code, out)
			}

			want := fmt.Sprintf("TargetIsIP=%v", tt.want)
			if !strings.Contains(out, want) {
				t.Errorf("%s: want %s. Output:\n%s", tt.target, want, out)
			}
		})
	}
}
