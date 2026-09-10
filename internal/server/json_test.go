package server

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// freeTCPAddress asks the OS for a free TCP port and releases it, so the
// server under test can bind it.
func freeTCPAddress(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to reserve a TCP port: %v", err)
	}
	address := listener.Addr().String()
	_ = listener.Close()

	return address
}

// captureStdout is where the events land, so a test has to read them from
// there rather than from a writer of its own.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}

	original := os.Stdout
	os.Stdout = w

	defer func() {
		os.Stdout = original
	}()

	f()

	_ = w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading captured output failed: %v", err)
	}

	return string(out)
}

// post sends one body to the server, retrying until it is listening, and
// returns the status it answered with.
func post(t *testing.T, address, body string) int {
	t.Helper()

	client := &http.Client{Timeout: 2 * time.Second}

	var lastErr error

	for range 20 {
		resp, err := client.Post("http://"+address+"/tcping", "application/json", strings.NewReader(body))
		if err != nil {
			// The server may not have bound the port yet.
			lastErr = err
			time.Sleep(50 * time.Millisecond)

			continue
		}

		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		return resp.StatusCode
	}

	t.Fatalf("the server never answered: %v", lastErr)

	return 0
}

// The events have to come out the way they went in, so that what a probing
// machine sends can be piped or redirected on the machine collecting it.
func TestListenJSON_PrintsTheEventsItReceives(t *testing.T) {
	address := freeTCPAddress(t)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	events := []string{
		`{"type":"start","data":{"hostname":"example.com","port":443}}`,
		`{"type":"probe","data":{"success":true,"latency":3.5}}`,
	}

	out := captureStdout(t, func() {
		go func() {
			if err := ListenJSON(ctx, address); err != nil {
				t.Errorf("ListenJSON() failed: %v", err)
			}
		}()

		for _, event := range events {
			// The event is printed before the sender is answered, so a
			// status in hand means the line is already out.
			if status := post(t, address, event+"\n"); status != http.StatusNoContent {
				t.Errorf("status = %d, want %d", status, http.StatusNoContent)
			}
		}
	})

	want := strings.Join(events, "\n") + "\n"
	if out != want {
		t.Errorf("printed\n%q\nwant\n%q", out, want)
	}
}

// Whatever reads the output should be able to count on every line being an
// event, so anything else is refused instead of printed.
func TestListenJSON_RefusesWhatIsNotAnEvent(t *testing.T) {
	address := freeTCPAddress(t)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	out := captureStdout(t, func() {
		go func() {
			if err := ListenJSON(ctx, address); err != nil {
				t.Errorf("ListenJSON() failed: %v", err)
			}
		}()

		if status := post(t, address, "not json at all"); status != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", status, http.StatusBadRequest)
		}

		client := &http.Client{Timeout: 2 * time.Second}

		resp, err := client.Get("http://" + address + "/tcping")
		if err != nil {
			t.Fatalf("GET failed: %v", err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("GET status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
		}
	})

	if out != "" {
		t.Errorf("printed something that was not an event: %q", out)
	}
}

func TestListenJSON_StopsWhenCancelled(t *testing.T) {
	address := freeTCPAddress(t)

	ctx, cancel := context.WithCancel(context.Background())

	stopped := make(chan error, 1)
	go func() {
		stopped <- ListenJSON(ctx, address)
	}()

	// Wait until it is really listening, so the cancel below cannot land
	// before the server exists. What is sent is refused on purpose, so
	// nothing of this test ends up on the terminal.
	post(t, address, "not an event")

	cancel()

	select {
	case err := <-stopped:
		if err != nil {
			t.Errorf("ListenJSON() returned %v, want nil after a cancel", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("ListenJSON() did not return after the context was cancelled")
	}
}
