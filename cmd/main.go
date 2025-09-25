// Package main enables tcping to execute as a CLI tool
package main

import (
	"bufio"
	"os"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/dns"
	"github.com/pouriyajamshidi/tcping/v3/internal/options"
	"github.com/pouriyajamshidi/tcping/v3/printers"
	"github.com/pouriyajamshidi/tcping/v3/probes/statistics"
	probes "github.com/pouriyajamshidi/tcping/v3/probes/tcp"
	"github.com/pouriyajamshidi/tcping/v3/types"
)

/* TODO:
- Pass `Prober` instead of tcping to printers, helpers, etc
- Implement functional pattern to chose the prober
- Probably it is better to move SignalHandler to probes.go instead of printers
- I think there are some overlaps in printer success and probe failure conditionals
- The PrintStatistics across printers seems like it has a LOT of duplicates. perhaps it can be refactored out
- Cross-check the printer implementations to see how much they differ
- See what printer methods are not used
- Show how long we were up on failure similar to what we do for success?
- Get DNS timeout as a user input option?
- Display name resolution times?
- Perhaps unexport the Colors in ColorPrinter
- Run modernize
- Read the entire code once everything is done for "code smells"
*/

// monitorStatsRequest checks stdin to see whether the 'Enter' key was pressed
// if so, prints the statistics
func monitorStatsRequest(p printers.Printer, s *statistics.Statistics) {
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
	tcping := &types.Tcping{}
	stats := &statistics.Statistics{}

	printer := options.ProcessUserInput(tcping, stats)

	printer.PrintStart(stats)

	tcping.StartTime = time.Now()

	stats.IP = dns.ResolveHostname(printer, stats, true, false)

	tcping.Ticker = time.NewTicker(tcping.Options.IntervalBetweenProbes)
	defer tcping.Ticker.Stop()

	printers.SignalHandler(printer, stats)

	if !tcping.Options.NonInteractive {
		go monitorStatsRequest(printer, stats)
	}

	var probeCount uint

	for {
		if tcping.Options.ShouldRetryResolve {
			dns.RetryResolveHostname(printer, stats, 300, true, false)
		}

		probes.Ping(stats, printer, tcping)

		if tcping.Options.ProbesBeforeQuit != 0 {
			probeCount++
			if probeCount == tcping.Options.ProbesBeforeQuit {
				printer.Shutdown(stats)
			}
		}
	}
}
