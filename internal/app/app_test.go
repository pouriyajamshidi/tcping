package app

import (
	"os"
	"testing"
	"time"
)

// withStdin points os.Stdin at a pipe for the duration of the test and
// returns the writing end, so a test can play the part of the user typing.
func withStdin(t *testing.T) *os.File {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}

	original := os.Stdin
	os.Stdin = r

	// Close the writing end first so the watcher goroutine sees EOF and
	// finishes before the next test swaps os.Stdin again.
	t.Cleanup(func() {
		w.Close()
		os.Stdin = original
		r.Close()
	})

	return w
}

func TestSummaryRequests_ReportsEmptyLines(t *testing.T) {
	stdin := withStdin(t)
	requests := SummaryRequests()

	if _, err := stdin.WriteString("\n"); err != nil {
		t.Fatalf("writing to stdin failed: %v", err)
	}

	select {
	case _, ok := <-requests:
		if !ok {
			t.Fatal("channel closed, want a request")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no request after an empty line")
	}
}

func TestSummaryRequests_IgnoresActualInput(t *testing.T) {
	stdin := withStdin(t)
	requests := SummaryRequests()

	if _, err := stdin.WriteString("hello\n"); err != nil {
		t.Fatalf("writing to stdin failed: %v", err)
	}

	select {
	case <-requests:
		t.Fatal("got a request for a non-empty line")
	case <-time.After(100 * time.Millisecond):
	}
}

// The old stdin loop treated a read error as something to retry, so once
// stdin hit EOF it spun on the error forever. Reaching the end of stdin
// has to end the goroutine instead.
func TestSummaryRequests_ClosesWhenStdinEnds(t *testing.T) {
	stdin := withStdin(t)
	requests := SummaryRequests()

	stdin.Close()

	select {
	case _, ok := <-requests:
		if ok {
			t.Fatal("got a request, want the channel closed")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("channel still open after stdin ended")
	}
}
