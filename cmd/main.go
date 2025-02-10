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
- Probably, it is better to pass the userInput to NewXYZPrinter rather than sub functions
- Should I nest the JSON output for example the latency stuff or let it lay flat?
- Pass Handler instead of tcping to printers helpers, etc
- Probably it is better to move SignalHandler to probes instead of printers
- SetLongestTime does not seem to belong to printer package
- Cross-check the printer implementations to see how much they differ
	 for instance JSONPrinter's PrintProbeFail lacks timestamp implementation
- Move types.NewLongestTime to printer instead?
- Make `Options` of `Tcping` implicit too?
- Where should we place the Shutdown function? Printers seems a bit off
- Separate probe packages. e.g. tcp.Ping, http.Ping
- Show how long we were up on failure similar to what we do for success?
- Should consts package move to internal?
- Should types package move to internal?
- Implement functional pattern to chose the prober
- Read the entire code once everything is done for "code smells"
- Implement non-interactive mode so that we can use it with `disown` and `nohup`
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
