// tcping.go probes a target using TCP
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/pouriyajamshidi/tcping/v3/internal/app"
	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/consts"
	"github.com/pouriyajamshidi/tcping/v3/internal/printers"
	"github.com/pouriyajamshidi/tcping/v3/internal/probers"
	"github.com/pouriyajamshidi/tcping/v3/internal/server"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
)

/* TODO:
Library support, so other codebases can import tcping (in priority order):
1. Move internal/config, internal/dns, internal/nic, internal/printers,
   internal/probers, internal/stats (and internal/consts) out of internal/
   so they're importable by other modules - blocks everything below
2. Make config.ProcessUserInput's CLI-only bits (os.Args, the global
   flag.CommandLine, os.Exit on invalid input) usable programmatically
3. Do not let checkForUpdates make a network call or exit when reachable
   from library code - keep it strictly opt-in CLI behavior
4. Give Statistics a Snapshot() so a caller can read it while a Prober is
   running. Nothing reads it unsynchronized today, but a library caller
   has no safe way to do it either.
5. Design a curated top-level public API/entrypoint instead of requiring
   several packages to be wired together by hand
6. Treat exported types/fields/methods as a public API contract once this
   is importable (semver discipline)
7. Sending over proxy connections
8. add system name to Alloy and Influx printers? so that we can distinguish them when graphing
*/

func main() {
	cfg := config.ProcessUserInput()

	if cfg.UDPServer {
		ctx := app.SetupSignalHandler(context.Background())

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

	probeCtx := app.SetupSignalHandler(context.Background())

	var pinger probers.Pinger

	switch cfg.Protocol {
	case consts.HTTP, consts.HTTPS:
		pinger = probers.NewHTTPing(cfg)
	case consts.UDP:
		pinger = probers.NewUDPing(cfg)
	default:
		pinger = probers.NewTcping(cfg)
	}

	// Only worth watching stdin when there is a user at a keyboard to
	// press 'Enter'.
	var summaryRequests <-chan struct{}
	if app.IsForegroundTerminal() {
		summaryRequests = app.SummaryRequests()
	}

	prober := probers.NewProber(pinger, printer, cfg, stats, summaryRequests)
	if err := prober.Probe(probeCtx); err != nil {
		printer.PrintError("%v", err)
	}

	printer.Shutdown(stats)
}
