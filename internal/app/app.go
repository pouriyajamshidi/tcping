package app

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/printers"
	"github.com/pouriyajamshidi/tcping/v3/internal/probers"
	"github.com/pouriyajamshidi/tcping/v3/internal/server"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
)

// setupSignalHandler catches SIGINT and SIGTERM
func setupSignalHandler(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	return ctx
}

// summaryRequests watches stdin and reports every time the user hits the
// 'Enter' key, so the statistics can be printed by whoever owns them rather
// than from this goroutine. The channel is closed when stdin runs out.
func summaryRequests() <-chan struct{} {
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

func Run() {
	cfg := config.ProcessUserInput()

	if cfg.UDPServer {
		ctx := setupSignalHandler(context.Background())

		listenAddr := net.JoinHostPort(cfg.IP.String(), strconv.Itoa(int(cfg.Port)))
		if err := server.ListenUDP(ctx, listenAddr); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to listen on %s: %s\n", listenAddr, err)
			os.Exit(1)
		}

		return
	}

	stats := stats.NewStatistics(cfg)

	printer, err := printers.NewPrinter(cfg.PrinterConfig)
	if err != nil {
		fmt.Printf("Failed to create printer: %s\n", err.Error())
		os.Exit(1)
	}

	probeCtx := setupSignalHandler(context.Background())

	var pinger probers.Pinger

	switch cfg.Protocol {
	case config.HTTP, config.HTTPS:
		pinger = probers.NewHTTPing(cfg)
	case config.UDP:
		pinger = probers.NewUDPing(cfg)
	default:
		pinger = probers.NewTcping(cfg)
	}

	// Only worth watching stdin when there is a user at a keyboard to
	// press 'Enter'.
	var summaryReqs <-chan struct{}
	if isForegroundTerminal() {
		summaryReqs = summaryRequests()
	}

	prober := probers.NewProber(pinger, printer, cfg, stats, summaryReqs)
	if err := prober.Probe(probeCtx); err != nil {
		printer.PrintError("%v", err)
	}

	printer.Shutdown(stats)
}
