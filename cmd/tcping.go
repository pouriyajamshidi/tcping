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
- Show how long we were up on failure similar to what we do for success?
- Get DNS timeout as a user input option?
- createDNSResolver to also account for the -I flag
- Display name resolution times?
- Use built-in slice functions for min max avg, etc
- Run modernize
- Maybe make the inclusion of timestamps in CSV filenames an option?
- Shutdown would not be needed once the program flow is solid
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
