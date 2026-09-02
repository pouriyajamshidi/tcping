// tcping.go probes a target using TCP
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pouriyajamshidi/tcping/v3/internal/app"
	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/consts"
	"github.com/pouriyajamshidi/tcping/v3/internal/printers"
	"github.com/pouriyajamshidi/tcping/v3/internal/probers"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
)

/* TODO:
Library support, so other codebases can import tcping (in priority order):
1. Move internal/config, internal/dns, internal/nic, internal/printers,
   internal/probers, internal/stats (and internal/consts) out of internal/
   so they're importable by other modules - blocks everything below
2. Make config.ProcessUserInput's CLI-only bits (os.Args, the global
   flag.CommandLine, os.Exit on invalid input) usable programmatically
3. Stop printers' Shutdown() from calling os.Exit directly - a library
   caller's process must not be killed by a printer
4. Do not let checkForUpdates make a network call or exit when reachable
   from library code - keep it strictly opt-in CLI behavior
5. Add a mutex/Snapshot() to Statistics so it is safe to read concurrently
   with an active Prober (app.MonitorSummaryRequest already does this
   unsynchronized today)
6. Keep signal handling (app.SetupSignalHandler) and stdin monitoring
   (app.MonitorSummaryRequest) CLI-only, out of the library's core path
7. Design a curated top-level public API/entrypoint instead of requiring
   several packages to be wired together by hand
8. Treat exported types/fields/methods as a public API contract once this
   is importable (semver discipline)

- Read the entire code once everything is done for "code smells"
*/

func main() {
	cfg := config.ProcessUserInput()

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
	case consts.TCP:
		pinger = probers.NewTcping(cfg)
	default:
		pinger = probers.NewTcping(cfg)
	}

	if app.IsForegroundTerminal() {
		go app.MonitorSummaryRequest(printer, stats)
	}

	prober := probers.NewProber(pinger, printer, cfg, stats)
	stats, err = prober.Probe(probeCtx)
	if err != nil {
		printer.PrintError("%v", err)
	}

	printer.Shutdown(stats)
}
