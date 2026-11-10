package app

import (
	"bytes"
	"errors"
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// targetEnv holds the address the child test process should probe. It is
// empty in the parent.
const targetEnv = "TCPING_TEST_TARGET"

// runOneProbe runs tcping in a fresh copy of the test binary with one probe
// against addr, and returns what it printed and the status it left with.
// Run ends the program through os.Exit, so a child process is the only way
// to see the status.
func runOneProbe(t *testing.T, addr string) (output string, exitCode int) {
	t.Helper()

	cmd := exec.Command(os.Args[0], "-test.run=^TestExitStatus$")
	cmd.Env = append(os.Environ(), targetEnv+"="+addr)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()

	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return out.String(), 0
	case errors.As(err, &exitErr):
		return out.String(), exitErr.ExitCode()
	default:
		t.Fatalf("running tcping in a subprocess: %v", err)
		return "", 0
	}
}

func TestExitStatus(t *testing.T) {
	if addr := os.Getenv(targetEnv); addr != "" {
		host, port, _ := net.SplitHostPort(addr)
		os.Args = []string{"tcping", "-c", "1", "--no-color", host, port}
		Run()
		os.Exit(0)
	}

	open, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer open.Close()

	closed, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	closedAddr := closed.Addr().String()
	_ = closed.Close()

	tests := []struct {
		name     string
		addr     string
		want     int
		wantLine string
	}{
		{"the probe succeeds", open.Addr().String(), 0, "1 received"},
		{"no probe succeeds", closedAddr, 1, "0 received"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, code := runOneProbe(t, tt.addr)

			if code != tt.want {
				t.Errorf("exit status %d, want %d. Output:\n%s", code, tt.want, out)
			}

			if !strings.Contains(out, tt.wantLine) {
				t.Errorf("output does not contain %q:\n%s", tt.wantLine, out)
			}
		})
	}
}
