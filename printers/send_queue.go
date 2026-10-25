package printers

import (
	"sync"
	"time"
)

const (
	// How often the metrics waiting in a sendQueue are sent. One request
	// every few seconds instead of one per probe is far less work for the
	// backend, and lets InfluxDB hold a write until its next fsync without
	// slowing the probing down.
	sendQueueInterval = 5 * time.Second

	// The most metrics a single request carries, so a queue that filled up
	// while the backend was down goes out in requests of a sane size once it
	// is back.
	sendQueueBatchSize = 1000

	// The most metrics a sendQueue holds while the backend cannot be
	// reached. At one probe a second that is somewhere between a quarter of
	// an hour and an hour of them, depending on the backend and the probe
	// type. Past that the oldest are dropped, so a backend that stays down
	// does not make tcping grow forever.
	sendQueueLimit = 10_000
)

// sendQueue holds the metrics a printer has not sent yet and sends them from
// a goroutine of its own, so the probing never waits on the network. A
// metric is one item of whatever the backend takes: a line for InfluxDB, a
// metric for OTLP.
type sendQueue[T any] struct {
	// send sends one batch, reports its own errors, and says whether the
	// batch is worth trying again, which is the case when the backend could
	// not be reached.
	send       func(batch []T) (retry bool)
	printError func(format string, args ...any)

	mu      sync.Mutex
	pending []T
	dropped int // How many were thrown away since that was last reported.
	limit   int

	// Held for the whole of a flush, so two of them never send at once.
	sending sync.Mutex

	stop    chan struct{}
	stopped chan struct{}
}

// newSendQueue starts a sendQueue that hands what it is given to send.
func newSendQueue[T any](send func([]T) bool, printError func(string, ...any)) *sendQueue[T] {
	q := &sendQueue[T]{
		send:       send,
		printError: printError,
		limit:      sendQueueLimit,
		stop:       make(chan struct{}),
		stopped:    make(chan struct{}),
	}

	go q.run()

	return q
}

// add puts metrics in line to be sent. It never waits on the network.
func (q *sendQueue[T]) add(items ...T) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.pending = append(q.pending, items...)
	q.trim()
}

// trim drops the oldest metrics once there are more than the limit, and
// counts them so flush can say so. It must be called with mu held.
func (q *sendQueue[T]) trim() {
	if over := len(q.pending) - q.limit; over > 0 {
		q.pending = q.pending[over:]
		q.dropped += over
	}
}

// run sends what is waiting every sendQueueInterval until close is called.
func (q *sendQueue[T]) run() {
	defer close(q.stopped)

	ticker := time.NewTicker(sendQueueInterval)
	defer ticker.Stop()

	for {
		select {
		case <-q.stop:
			return

		case <-ticker.C:
			q.flush()
		}
	}
}

// flush sends everything waiting, a batch at a time, and reports whether all
// of it went out. A batch that has to be tried again goes back to the front
// of the line and ends the flush, so a backend that is down is asked once
// per interval rather than over and over.
func (q *sendQueue[T]) flush() bool {
	q.sending.Lock()
	defer q.sending.Unlock()

	for {
		q.mu.Lock()
		dropped := q.dropped
		q.dropped = 0
		n := min(len(q.pending), sendQueueBatchSize)
		// Capped at n, so putting it back below cannot write over what
		// add appended in the meantime.
		batch := q.pending[:n:n]
		q.pending = q.pending[n:]
		q.mu.Unlock()

		if dropped > 0 {
			q.printError("the buffer of unsent metrics was full, so the %d oldest were dropped", dropped)
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

	q.printError("gave up on %d metrics that could not be sent", len(q.pending)+q.dropped)
}
