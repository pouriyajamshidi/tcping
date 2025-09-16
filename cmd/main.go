// Package main enables tcping to execute as a CLI tool
package main

import (
	"time"

	"github.com/pouriyajamshidi/tcping/v2/internal/dns"
	"github.com/pouriyajamshidi/tcping/v2/internal/options"
	"github.com/pouriyajamshidi/tcping/v2/internal/utils"
	"github.com/pouriyajamshidi/tcping/v2/printers"
	probes "github.com/pouriyajamshidi/tcping/v2/probes/tcp"
	"github.com/pouriyajamshidi/tcping/v2/types"
)

/* TODO:
- Implement functional pattern to chose the prober
- Pass `Prober` instead of tcping to printers, helpers, etc
- Probably it is better to move SignalHandler to probes.go instead of printers
- I think there are some overlaps in printer success and probe failure conditionals
- The PrintStatistics across printers seems like it has a LOT of duplicates. perhaps it can be refactored out
- Cross-check the printer implementations to see how much they differ
- See what printer methods are not used
- Show how long we were up on failure similar to what we do for success?
- Get DNS timeout as a user input option?
- Display name resolution times?
- Run modernize
- Read the entire code once everything is done for "code smells"
*/

func monitorStatsRequest(stdinChan chan bool, tcping *types.Tcping) {
	for pressedEnter := range stdinChan {
		if pressedEnter {
			printers.PrintStats(tcping)
		}
	}
}

func main() {
	tcping := &types.Tcping{}

	options.ProcessUserInput(tcping)

	tcping.PrintStart(tcping.Options.Hostname, tcping.Options.Port)

	tcping.StartTime = time.Now()

	tcping.Ticker = time.NewTicker(tcping.Options.IntervalBetweenProbes)
	defer tcping.Ticker.Stop()

	printers.SignalHandler(tcping)

	if !tcping.Options.NonInteractive {
		stdinChan := make(chan bool, 1)
		go utils.MonitorSTDIN(stdinChan)
		go monitorStatsRequest(stdinChan, tcping)
	}

	var probeCount uint

	for {
		if tcping.Options.ShouldRetryResolve {
			dns.RetryResolveHostname(tcping)
		}

		probes.Ping(tcping)

		if tcping.Options.ProbesBeforeQuit != 0 {
			probeCount++
			if probeCount == tcping.Options.ProbesBeforeQuit {
				printers.Shutdown(tcping)
			}
		}
	}
}
