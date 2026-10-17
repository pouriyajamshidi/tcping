package cli

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"testing"
)

// subprocessCase names the case a child test process should run. It is empty
// in the parent.
const subprocessCase = "TCPING_TEST_CASE"

// runCase re-runs the named test in a fresh copy of the test binary, with
// subprocessCase set to caseName. Bad input ends the program through
// os.Exit, so running the case in a child is the only way to see what it
// printed and which status it left with. It returns what went to stdout and
// to stderr apart, so a test can tell which one a message went to.
func runCase(t *testing.T, test, caseName string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command(os.Args[0], "-test.run=^"+test+"$")
	cmd.Env = append(os.Environ(), subprocessCase+"="+caseName)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()

	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return outBuf.String(), errBuf.String(), 0
	case errors.As(err, &exitErr):
		return outBuf.String(), errBuf.String(), exitErr.ExitCode()
	default:
		t.Fatalf("running %s in a subprocess: %v", test, err)
		return "", "", 0
	}
}
