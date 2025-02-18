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
- Take care of `startTime time.Time` for printers other than colored and plain
- I think there are some overlaps in printer success and probe failure conditionals
- Rename `PrintProbeFail` to `PrintProbeFailure`
- Should I implement a Done function for CSV and DB printers instead of doing different things in Shutdown()?
- retry resolve hostname is not shown in other printers than color and plain
- `PrintInfo` and possibly some other methods of the printer interface are not used -- remove
- Probably it is better to move SignalHandler to probes instead of printers
- SetLongestTime does not seem to belong to printer package
- Cross-check the printer implementations to see how much they differ
- Implement a non-interactive mode
- Move types.NewLongestTime to printer instead?
- Make `Options` of `Tcping` implicit too?
- Pass Handler instead of tcping to printers helpers, etc
- Where should we place the Shutdown function? Printers seems a bit off
- Separate probe packages. e.g. tcp.Ping, http.Ping
- Show how long we were up on failure similar to what we do for success?
- Should consts package move to internal?
- Should types package move to internal?
- Implement functional pattern to chose the prober
- Read the entire code once everything is done for "code smells"
- Implement non-interactive mode so that we can use it with `disown` and `nohup`
- Possibly use new slice functions instead of the current manual way
- See what printer methods are not used
- The PrintStatistics across printers seems like it has a LOT of duplicates. perhaps it can be refactored out
*/

func main() {
	tcping := &types.Tcping{}

	options.ProcessUserInput(tcping)

	tcping.PrintStart(tcping.Options.Hostname, tcping.Options.Port)

	tcping.StartTime = time.Now()

	tcping.Ticker = time.NewTicker(tcping.Options.IntervalBetweenProbes)
	defer tcping.Ticker.Stop()

	printers.SignalHandler(tcping)

	stdinChan := make(chan bool)
	go utils.MonitorSTDIN(stdinChan)

	var probeCount uint

	for {
		if tcping.Options.ShouldRetryResolve {
			dns.RetryResolveHostname(tcping)
		}

		probes.Ping(tcping)

		select {
		case pressedEnter := <-stdinChan:
			if pressedEnter {
				printers.PrintStats(tcping)
			}
		default:
		}

		if tcping.Options.ProbesBeforeQuit != 0 {
			probeCount++
			if probeCount == tcping.Options.ProbesBeforeQuit {
				printers.Shutdown(tcping)
			}
		}
	}
}
