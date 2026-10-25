package printers

import (
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"
)

// queueRecorder is a backend for a sendQueue that keeps every batch it was
// sent and every error the queue printed.
type queueRecorder struct {
	mu      sync.Mutex
	batches [][]int
	errors  []string
	// How many sends in a row fail with a "try again" before they work.
	failures int
}

func (r *queueRecorder) send(batch []int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.failures > 0 {
		r.failures--
		return true
	}

	r.batches = append(r.batches, slices.Clone(batch))

	return false
}

func (r *queueRecorder) printError(format string, args ...any) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.errors = append(r.errors, fmt.Sprintf(format, args...))
}

func (r *queueRecorder) sent() []int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return slices.Concat(r.batches...)
}

// This is the whole point of the queue: a backend that takes its time must
// not hold up the probe that hands it the next results.
func TestSendQueueAddDoesNotWaitOnASend(t *testing.T) {
	entered := make(chan struct{}, 1)
	release := make(chan struct{})

	q := newSendQueue(time.Hour, 1000, func([]int) bool {
		select {
		case entered <- struct{}{}:
		default:
		}
		<-release
		return false
	}, func(string, ...any) {})

	q.add(1)
	go q.flush()
	<-entered

	added := make(chan struct{})
	go func() {
		q.add(2)
		close(added)
	}()

	select {
	case <-added:
	case <-time.After(time.Second):
		t.Error("add waited on a send that was still going")
	}

	close(release)
}

// With no interval, what is added goes out right away, which is what keeps
// a JSON stream live.
func TestSendQueueWithNoIntervalSendsRightAway(t *testing.T) {
	sent := make(chan []int, 1)

	q := newSendQueue(0, 1, func(batch []int) bool {
		sent <- slices.Clone(batch)
		return false
	}, func(string, ...any) {})

	q.add(7)

	select {
	case got := <-sent:
		if !slices.Equal(got, []int{7}) {
			t.Errorf("sent %v, want [7]", got)
		}
	case <-time.After(time.Second):
		t.Error("nothing was sent without a flush being asked for")
	}

	q.close()
}

// A backend that was down for a while gets everything once it is back, in
// the order it happened.
func TestSendQueueTriesAgainAfterAFailure(t *testing.T) {
	r := &queueRecorder{failures: 1}
	q := newSendQueue(time.Hour, 1000, r.send, r.printError)

	q.add(1, 2)
	if q.flush() {
		t.Fatal("flush said everything went out, but the send failed")
	}

	q.add(3)
	if !q.flush() {
		t.Fatal("flush said something was left, but the send worked")
	}

	if got, want := r.sent(), []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Errorf("sent %v, want %v", got, want)
	}
}

// A backend that stays down must not make the queue grow forever, and losing
// results is never done quietly.
func TestSendQueueDropsTheOldestWhenFull(t *testing.T) {
	r := &queueRecorder{}
	q := newSendQueue(time.Hour, 1000, r.send, r.printError)
	q.limit = 3

	q.add(1, 2, 3, 4, 5)
	q.flush()

	if got, want := r.sent(), []int{3, 4, 5}; !slices.Equal(got, want) {
		t.Errorf("sent %v, want %v", got, want)
	}

	want := []string{"the buffer of unsent results was full, so the 2 oldest were dropped"}
	if !slices.Equal(r.errors, want) {
		t.Errorf("errors = %q, want %q", r.errors, want)
	}
}

// A queue that filled up while the backend was down goes out in requests of
// a sane size, not in one huge one.
func TestSendQueueSendsInBatches(t *testing.T) {
	r := &queueRecorder{}
	q := newSendQueue(time.Hour, 1000, r.send, r.printError)

	for i := range 2*1000 + 1 {
		q.add(i)
	}
	q.flush()

	var sizes []int
	for _, batch := range r.batches {
		sizes = append(sizes, len(batch))
	}

	if want := []int{1000, 1000, 1}; !slices.Equal(sizes, want) {
		t.Errorf("batch sizes = %v, want %v", sizes, want)
	}
}

// The last probes and the final summary are still waiting when tcping exits,
// so closing has to send them.
func TestSendQueueCloseSendsWhatIsLeft(t *testing.T) {
	r := &queueRecorder{}
	q := newSendQueue(time.Hour, 1000, r.send, r.printError)

	q.add(1, 2)
	q.close()

	if got, want := r.sent(), []int{1, 2}; !slices.Equal(got, want) {
		t.Errorf("sent %v, want %v", got, want)
	}

	if len(r.errors) != 0 {
		t.Errorf("unexpected errors: %q", r.errors)
	}
}

// Nothing tries again after close, so what it could not send is lost, and
// that has to be said.
func TestSendQueueCloseReportsWhatIsLost(t *testing.T) {
	r := &queueRecorder{failures: 1}
	q := newSendQueue(time.Hour, 1000, r.send, r.printError)

	q.add(1, 2)
	q.close()

	want := []string{"gave up on 2 results that could not be sent"}
	if !slices.Equal(r.errors, want) {
		t.Errorf("errors = %q, want %q", r.errors, want)
	}
}
