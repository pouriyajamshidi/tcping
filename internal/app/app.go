package app

import (
	"bufio"
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// SetupSignalHandler catches SIGINT and SIGTERM
func SetupSignalHandler(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	return ctx
}

// SummaryRequests watches stdin and reports every time the user hits the
// 'Enter' key, so the statistics can be printed by whoever owns them rather
// than from this goroutine. The channel is closed when stdin runs out.
func SummaryRequests() <-chan struct{} {
	// One buffered slot: if a probe is in flight when 'Enter' is pressed we
	// keep the request and hand it over as soon as the prober asks, and
	// holding on to more than one is pointless since they all print the
	// same thing.
	requests := make(chan struct{}, 1)

	// Bind to stdin here rather than inside the goroutine, so the caller
	// decides what is being read.
	scanner := bufio.NewScanner(os.Stdin)

	go func() {
		defer close(requests)

		for scanner.Scan() {
			if scanner.Err() != nil {
				continue
			}

			if strings.TrimSpace(scanner.Text()) != "" {
				continue
			}

			select {
			case requests <- struct{}{}:
			default:
			}
		}
	}()

	return requests
}
