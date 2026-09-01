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
- Display name resolution times?
- Run modernize
- Maybe make the inclusion of timestamps in CSV filenames an option?
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
