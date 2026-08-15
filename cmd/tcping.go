// tcping.go probes a target using TCP
package main

import (
	"bufio"
	"fmt"
	"os"

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
- Use built-in slice functions for min max avg, etc
- Read the entire code once everything is done for "code smells"
*/

// monitorSummaryRequest checks stdin to see whether the 'Enter' key was pressed
// if so, it prints the statistics
func monitorSummaryRequest(p printers.Printer, s *stats.Statistics) {
	reader := bufio.NewReader(os.Stdin)

	stdinChan := make(chan bool, 1)

	go func() {
		for {
			input, err := reader.ReadString('\n')
			if err != nil {
				continue
			}

			if input == "\n" || input == "\r" || input == "\r\n" {
				stdinChan <- true
			}
		}
	}()

	for pressedEnter := range stdinChan {
		if pressedEnter {
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

	printers.SignalHandler(printer, stats)

	var pinger probers.Pinger

	switch cfg.Protocol {
	case consts.TCP:
		pinger = probers.Tcping{}
	default:
		pinger = probers.Tcping{}
	}

	if !cfg.NonInteractive {
		go monitorSummaryRequest(printer, stats)
	}

	printer.PrintStart(stats)

	var probeCount uint

	for {
		if cfg.ShouldRetryResolve && stats.OngoingUnsuccessfulProbes >= cfg.RetryResolveAfterNFailures {
			stats.RetriedHostnameLookups++
			printer.PrintRetryingToResolve(stats.Hostname)
			if err := cfg.Resolver.RetryResolveHostname(
				stats,
				cfg.UseIPv4,
				cfg.UseIPv6,
			); err != nil {
				printer.PrintError("%s", err.Error())
			}
		}

		pinger.Ping(stats, printer, cfg)

		// probers.Ping(stats, printer, tcping, cfg)

		// -c flag is provided
		if cfg.ProbesBeforeQuit != 0 {
			probeCount++
			if probeCount == cfg.ProbesBeforeQuit {
				printer.Shutdown(stats)
			}
		}
	}
}
