package printers

import (
	"sync"
	"time"
)

// The most results a sendQueue holds while the backend cannot be reached. At
// one probe a second that is somewhere between a quarter of an hour and an
// hour of them, depending on the backend and the probe type. Past that the
// oldest are dropped, so a backend that stays down does not make tcping grow
// forever.
const sendQueueLimit = 10_000

// sendQueue holds the results a printer has not sent yet and sends them from
// a goroutine of its own, so the probing never waits on the network. A
// result is one item of whatever the backend takes: a line for InfluxDB, a
// metric for OTLP, an event for a JSON stream.
type sendQueue[T any] struct {
	// send sends one batch, reports its own errors, and says whether the
	// batch is worth trying again, which is the case when the backend could
	// not be reached.
	send       func(batch []T) (retry bool)
	printError func(format string, args ...any)

	// How often what is waiting is sent. Zero sends it as soon as it is
	// added.
	every time.Duration
	// The most results a single send carries.
	batchSize int

	mu      sync.Mutex
	pending []T
	dropped int // How many were thrown away since that was last reported.
	limit   int

	// Held for the whole of a flush, so two of them never send at once.
	sending sync.Mutex

	// Told about every add, in case it has to be sent right away.
	added   chan struct{}
	stop    chan struct{}
	stopped chan struct{}
}

// newSendQueue starts a sendQueue that hands what it is given to send, every
// so often or right away, at most batchSize results at a time.
func newSendQueue[T any](every time.Duration, batchSize int, send func([]T) bool, printError func(string, ...any)) *sendQueue[T] {
	q := &sendQueue[T]{
		send:       send,
		printError: printError,
		every:      every,
		batchSize:  batchSize,
		limit:      sendQueueLimit,
		added:      make(chan struct{}, 1),
		stop:       make(chan struct{}),
		stopped:    make(chan struct{}),
	}

	go q.run()

	return q
}

// add puts results in line to be sent. It never waits on the network.
func (q *sendQueue[T]) add(items ...T) {
	q.mu.Lock()
	q.pending = append(q.pending, items...)
	q.trim()
	q.mu.Unlock()

	// One waiting signal is enough, since a flush sends everything.
	select {
	case q.added <- struct{}{}:
	default:
	}
}

// trim drops the oldest results once there are more than the limit, and
// counts them so flush can say so. It must be called with mu held.
func (q *sendQueue[T]) trim() {
	if over := len(q.pending) - q.limit; over > 0 {
		q.pending = q.pending[over:]
		q.dropped += over
	}
}

// run sends what is waiting, every so often or as soon as it is added,
// until close is called. Only one of the two channels is ever set, and a nil
// one never fires.
func (q *sendQueue[T]) run() {
	defer close(q.stopped)

	var tick <-chan time.Time
	added := q.added

	if q.every > 0 {
		ticker := time.NewTicker(q.every)
		defer ticker.Stop()

		tick = ticker.C
		added = nil
	}

	for {
		select {
		case <-q.stop:
			return

		case <-tick:
			q.flush()

		case <-added:
			q.flush()
		}
	}
}

// flush sends everything waiting, a batch at a time, and reports whether all
// of it went out. A batch that has to be tried again goes back to the front
// of the line and ends the flush, so a backend that is down is asked once
// per interval, or once per add, rather than over and over.
func (q *sendQueue[T]) flush() bool {
	q.sending.Lock()
	defer q.sending.Unlock()

	for {
		q.mu.Lock()
		dropped := q.dropped
		q.dropped = 0
		n := min(len(q.pending), q.batchSize)
		// Capped at n, so putting it back below cannot write over what
		// add appended in the meantime.
		batch := q.pending[:n:n]
		q.pending = q.pending[n:]
		q.mu.Unlock()

		if dropped > 0 {
			q.printError("the buffer of unsent results was full, so the %d oldest were dropped", dropped)
		}

		if len(batch) == 0 {
			return true
		}

		if !q.send(batch) {
			continue
		}

		q.mu.Lock()
		q.pending = append(batch, q.pending...)
		q.trim()
		q.mu.Unlock()

		return false
	}
}

// close stops the goroutine and sends what is still waiting, so the last
// probes and the final summary are not lost. Whatever cannot be sent by then
// is reported, since nothing is going to try again.
func (q *sendQueue[T]) close() {
	close(q.stop)
	<-q.stopped

	if q.flush() {
		return
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	q.printError("gave up on %d results that could not be sent", len(q.pending)+q.dropped)
}
