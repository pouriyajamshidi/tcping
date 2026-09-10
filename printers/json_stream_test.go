package printers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
)

// jsonStreamServer is a server that takes the events and keeps them, so a
// test can look at what really went over the wire.
type jsonStreamServer struct {
	*httptest.Server

	mu           sync.Mutex
	bodies       []string
	contentTypes []string
}

func newJSONStreamServer(t *testing.T, status int) *jsonStreamServer {
	t.Helper()

	s := &jsonStreamServer{}

	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		s.mu.Lock()
		s.bodies = append(s.bodies, string(body))
		s.contentTypes = append(s.contentTypes, r.Header.Get("Content-Type"))
		s.mu.Unlock()

		w.WriteHeader(status)
	}))

	t.Cleanup(s.Close)

	return s
}

func (s *jsonStreamServer) received() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.bodies
}

// captureStderr is captureStdout for the other stream, which is where a
// printer complains when its destination is not answering.
func captureStderr(t *testing.T, f func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}

	original := os.Stderr
	os.Stderr = w

	defer func() {
		os.Stderr = original
	}()

	f()

	_ = w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading captured output failed: %v", err)
	}

	return string(out)
}

// A streamed run has to reach the server as the same events a piped one
// writes to the terminal, one JSON event per POST, so the receiving side can
// decode a request body without having to split it first.
func TestJSONStreamSendsEveryEvent(t *testing.T) {
	server := newJSONStreamServer(t, http.StatusOK)

	scriptedRun{
		printer:    jsonTestPrinter(t, Config{JSONURL: server.URL}),
		outcomes:   upDownUp,
		enterAfter: 4,
	}.run(t)

	var got []string

	for _, body := range server.received() {
		var event struct {
			Type string `json:"type"`
		}

		if err := json.Unmarshal([]byte(body), &event); err != nil {
			t.Fatalf("body %q is not a single JSON event: %v", body, err)
		}

		got = append(got, event.Type)
	}

	want := []string{
		"start",
		"probe", "probe", // up
		"probe", "probe", // down
		"statistics",     // "Enter" pressed after the fourth probe
		"probe", "probe", // up again
		"statistics", // on the way out
	}

	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("events = %v, want %v", got, want)
	}

	for _, contentType := range server.contentTypes {
		if contentType != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", contentType)
		}
	}
}

// A server that is not answering, or not happy with what it got, must not
// stop the probing, and must not repeat itself every second either.
func TestJSONStreamKeepsProbingWhenTheServerRejects(t *testing.T) {
	server := newJSONStreamServer(t, http.StatusInternalServerError)

	var out string

	stderr := captureStderr(t, func() {
		out = scriptedRun{
			printer:    jsonTestPrinter(t, Config{JSONURL: server.URL}),
			outcomes:   upDownUp,
			enterAfter: 4,
		}.run(t)
	})

	if len(server.received()) != 9 {
		t.Errorf("the run stopped after %d events, want all 9", len(server.received()))
	}

	if out != "" {
		t.Errorf("the events fell back to the terminal:\n%s", out)
	}

	if warnings := strings.Count(stderr, "JSON stream Error:"); warnings != 1 {
		t.Errorf("complained %d times, want once:\n%s", warnings, stderr)
	}
}

func TestJSONStreamAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    string
		wantErr bool
	}{
		{
			name:    "a full address is left alone",
			address: "http://localhost:8000/tcping",
			want:    "http://localhost:8000/tcping",
		},
		{
			name:    "a missing scheme is filled in",
			address: "localhost:8000/tcping",
			want:    "http://localhost:8000/tcping",
		},
		{
			name:    "https is kept",
			address: "https://example.com/tcping",
			want:    "https://example.com/tcping",
		},
		{
			name:    "an address with no host is rejected",
			address: "http:///tcping",
			wantErr: true,
		},
		{
			name:    "a malformed address is rejected",
			address: "http://local host:8000",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writer, err := newJSONStreamWriter(test.address)

			if test.wantErr {
				if err == nil {
					t.Fatalf("newJSONStreamWriter(%q) succeeded, want an error", test.address)
				}
				return
			}

			if err != nil {
				t.Fatalf("newJSONStreamWriter(%q) failed: %v", test.address, err)
			}

			if writer.url != test.want {
				t.Errorf("url = %q, want %q", writer.url, test.want)
			}
		})
	}
}

// The printer is created before the first probe, so a typo in the address
// has to be an error the run can report rather than events dropped one at a
// time.
func TestJSONPrinterRejectsBadStreamAddress(t *testing.T) {
	if _, err := NewJSONPrinter(Config{JSONURL: "http://local host:8000"}); err == nil {
		t.Error("NewJSONPrinter() accepted an address it cannot POST to")
	}
}
