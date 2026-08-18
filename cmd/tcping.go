// tcping.go probes a target using TCP
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pouriyajamshidi/tcping/v3/internal/app"
	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/consts"
	"github.com/pouriyajamshidi/tcping/v3/internal/printers"
	"github.com/pouriyajamshidi/tcping/v3/internal/probers"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
)

/* TODO:
- Pass `Prober` instead of tcping to printers, helpers, etc
- Implement functional pattern to chose the prober
- Probably it is better to move SignalHandler to probes.go instead of printers
- The PrintStatistics across printers seems like it has a LOT of duplicates. perhaps it can be refactored out
- Cross-check the printer implementations to see how much they differ
- See what printer methods are not used
- Show how long we were up on failure similar to what we do for success?
- Get DNS timeout as a user input option?
- Display name resolution times?
- Perhaps unexport the Colors in ColorPrinter
- Run modernize
- Remove the non-interactive flag? since we check the TTY now
- Make tcping background aware?
- Use built-in slice functions for min max avg, etc
- Read the entire code once everything is done for "code smells"
*/

// monitorSummaryRequest checks stdin to see whether the 'Enter' key was pressed
// if so, it prints the statistics
func monitorSummaryRequest(p printers.Printer, s *stats.Statistics) {
	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			continue
		}

		if strings.TrimSpace(input) == "" {
			printers.PrintStats(p, s)
		}
	}
}

func main() {
	cfg := config.ProcessUserInput()

	stats := stats.NewStatistics(cfg)

	printer, err := printers.NewPrinter(cfg.PrinterConfig)
	if err != nil {
		fmt.Printf("Failed to create printer: %s\n", err)
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

	if isForegroundTerminal() {
		go monitorSummaryRequest(printer, stats)
	}

	prober := probers.NewProber(pinger, cfg)

	stats, err = prober.Probe(probeCtx)
	// stats, err = probers.Run(probeCtx, pinger, printer, stats, cfg)
	if err != nil {
		printer.PrintError("%v", err)
	}
	printer.PrintStatistics(stats)
}
