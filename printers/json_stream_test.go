package printers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

// jsonStreamServer is a server that takes the events and keeps them, so a
// test can look at what really went over the wire.
type jsonStreamServer struct {
	*httptest.Server

	mu           sync.Mutex
	bodies       []string
	contentTypes []string
	// What each POST was answered with, in the same order as bodies.
	statuses []int
	// How many POSTs in a row get a 503 before the server answers normally.
	unavailable int
}

func newJSONStreamServer(t *testing.T, status int) *jsonStreamServer {
	t.Helper()

	s := &jsonStreamServer{}

	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		s.mu.Lock()
		answer := status
		if s.unavailable > 0 {
			s.unavailable--
			answer = http.StatusServiceUnavailable
		}
		s.bodies = append(s.bodies, string(body))
		s.contentTypes = append(s.contentTypes, r.Header.Get("Content-Type"))
		s.statuses = append(s.statuses, answer)
		s.mu.Unlock()

		w.WriteHeader(answer)
	}))

	t.Cleanup(s.Close)

	return s
}

func (s *jsonStreamServer) received() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.bodies
}

// eventTypes is the type of every event in the bodies, in order.
func eventTypes(t *testing.T, bodies []string) []string {
	t.Helper()

	var types []string

	for _, body := range bodies {
		var event struct {
			Type string `json:"type"`
		}

		if err := json.Unmarshal([]byte(body), &event); err != nil {
			t.Fatalf("body %q is not a single JSON event: %v", body, err)
		}

		types = append(types, event.Type)
	}

	return types
}

// upDownUpEvents are the events a streamed upDownUp run sends, with "Enter"
// pressed after the fourth probe.
var upDownUpEvents = []string{
	"start",
	"probe", "probe", // up
	"probe", "probe", // down
	"statistics",     // "Enter" pressed after the fourth probe
	"probe", "probe", // up again
	"statistics", // on the way out
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

	if got := eventTypes(t, server.received()); !slices.Equal(got, upDownUpEvents) {
		t.Errorf("events = %v, want %v", got, upDownUpEvents)
	}

	for _, contentType := range server.contentTypes {
		if contentType != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", contentType)
		}
	}
}

// A server that is not happy with what it got must not stop the probing, and
// every event it refused has to be reported. It would refuse the same event
// again, so none of them is sent twice.
func TestJSONStreamKeepsProbingWhenTheServerRejects(t *testing.T) {
	server := newJSONStreamServer(t, http.StatusBadRequest)

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

	if errors := strings.Count(stderr, "JSON stream Error:"); errors != 9 {
		t.Errorf("9 refused events printed %d errors, want 9:\n%s", errors, stderr)
	}
}

// A server that is restarting, or too busy to answer, gets the events it
// missed once it is back, in order and each one once.
func TestJSONStreamSendsWhatItMissedOnceBack(t *testing.T) {
	server := newJSONStreamServer(t, http.StatusOK)
	server.unavailable = 1

	stderr := captureStderr(t, func() {
		scriptedRun{
			printer:    jsonTestPrinter(t, Config{JSONURL: server.URL}),
			outcomes:   upDownUp,
			enterAfter: 4,
		}.run(t)
	})

	if !strings.Contains(stderr, "JSON stream Error: "+server.URL+" rejected the event with 503 Service Unavailable") {
		t.Errorf("the failed POST was not reported:\n%s", stderr)
	}

	var accepted []string
	for i, body := range server.received() {
		if server.statuses[i] == http.StatusOK {
			accepted = append(accepted, body)
		}
	}

	if got := eventTypes(t, accepted); !slices.Equal(got, upDownUpEvents) {
		t.Errorf("events = %v, want %v", got, upDownUpEvents)
	}
}

// A slow server is what used to slow the probing down: every event waited
// for its own POST to be answered. Now an event only goes in line, so the
// next one is written even while a POST is still waiting for an answer.
func TestJSONStreamDoesNotWaitOnASlowServer(t *testing.T) {
	entered := make(chan struct{}, 1)
	release := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		select {
		case entered <- struct{}{}:
		default:
		}
		<-release
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	writer, err := newJSONStreamWriter(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	_, _ = writer.Write([]byte(`{"type":"probe"}`))
	<-entered

	written := make(chan struct{})
	go func() {
		_, _ = writer.Write([]byte(`{"type":"probe"}`))
		close(written)
	}()

	select {
	case <-written:
	case <-time.After(time.Second):
		t.Error("an event waited on a POST that the server had not answered yet")
	}

	close(release)
	<-written
	writer.close()
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
