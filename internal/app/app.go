package app

import (
	"context"
	"os"
	"os/signal"
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
