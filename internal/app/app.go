package app

import (
	"bufio"
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/pouriyajamshidi/tcping/v3/internal/printers"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
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

// MonitorSummaryRequest checks stdin to see whether the 'Enter' key was pressed
// if so, it prints the statistics
func MonitorSummaryRequest(p printers.Printer, s *stats.Statistics) {
	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			continue
		}

		if strings.TrimSpace(input) == "" {
			p.PrintStatistics(s)
		}
	}
}
