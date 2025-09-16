// Package main enables tcping to execute as a CLI tool
package main

import (
	"time"

	"github.com/pouriyajamshidi/tcping/v2/internal/dns"
	"github.com/pouriyajamshidi/tcping/v2/internal/options"
	"github.com/pouriyajamshidi/tcping/v2/internal/utils"
	"github.com/pouriyajamshidi/tcping/v2/printers"
	"github.com/pouriyajamshidi/tcping/v2/probes"
	"github.com/pouriyajamshidi/tcping/v2/types"
)

/* TODO:
- Pass `Prober` instead of tcping to printers, helpers, etc
- I think there are some overlaps in printer success and probe failure conditionals
- Where should we place the Shutdown function? Printers seems a bit off
- Probably it is better to move SignalHandler to probes instead of printers
- Make `Options` of `Tcping` implicit too?
- Move types.NewLongestTime to printer instead?
- Should consts package move to internal?
- Should types package move to internal?
- The PrintStatistics across printers seems like it has a LOT of duplicates. perhaps it can be refactored out
- Show how long we were up on failure similar to what we do for success?
- Implement functional pattern to chose the prober
- Cross-check the printer implementations to see how much they differ
- Only the `NewLongestTime` is in the types.go file while others aren't
- See what printer methods are not used
- Run modernize
- Read the entire code once everything is done for "code smells"
*/

func main() {
	tcping := &types.Tcping{}

	options.ProcessUserInput(tcping)

	tcping.PrintStart(tcping.Options.Hostname, tcping.Options.Port)

	tcping.StartTime = time.Now()

	tcping.Ticker = time.NewTicker(tcping.Options.IntervalBetweenProbes)
	defer tcping.Ticker.Stop()

	printers.SignalHandler(tcping)

	var stdinChan chan bool
	if !tcping.Options.NonInteractive {
		stdinChan = make(chan bool)
		go utils.MonitorSTDIN(stdinChan)
	}

	var probeCount uint

	for {
		if tcping.Options.ShouldRetryResolve {
			dns.RetryResolveHostname(tcping)
		}

		probes.Ping(tcping)

		if stdinChan != nil {
			select {
			case pressedEnter := <-stdinChan:
				if pressedEnter {
					printers.PrintStats(tcping)
				}
			default:
			}
		}

		if tcping.Options.ProbesBeforeQuit != 0 {
			probeCount++
			if probeCount == tcping.Options.ProbesBeforeQuit {
				printers.Shutdown(tcping)
			}
		}
	}
}
