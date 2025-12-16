// Package testdata provides shared test helpers and fixtures.
package testdata

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/netip"
	"os"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/printers"
)

// Common test fixture values
const (
	TestHostname = "example.com"
	TestPort     = uint16(443)
	TestPort80   = uint16(80)
	TestPort8080 = uint16(8080)
)

var (
	TestIP          = netip.MustParseAddr("192.168.1.1")
	TestIP2         = netip.MustParseAddr("10.0.0.1")
	TestTimestamp   = time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	TestTimestamp2  = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	TestTimestamp3  = time.Date(2024, 1, 15, 14, 22, 10, 0, time.UTC)
	TestSourceAddr  = "10.0.0.1:12345"
)

// MockAddr implements net.Addr for testing.
type MockAddr struct {
	Addr string
}

func (m MockAddr) Network() string { return "tcp" }
func (m MockAddr) String() string  { return m.Addr }

var _ net.Addr = (*MockAddr)(nil)

// ToPtr returns a pointer to the provided value.
func ToPtr[T any](v T) *T {
	return &v
}

// CaptureOutput captures stdout during function execution and returns it as a string.
func CaptureOutput(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	w.Close()
	output := <-done
	os.Stdout = oldStdout

	return output
}

// CaptureJSONOutput captures and parses JSON output from stdout.
func CaptureJSONOutput(t *testing.T, fn func()) printers.JSONData {
	t.Helper()

	output := CaptureOutput(t, fn)

	var data printers.JSONData
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		t.Fatalf("parse JSON: %v\nOutput: %s", err, output)
	}

	return data
}
