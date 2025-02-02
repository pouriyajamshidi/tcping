// Package main enables tcping to execute as a CLI tool
package main

import (
	"time"

	"github.com/pouriyajamshidi/tcping/v2/internal/nameresolution"
	"github.com/pouriyajamshidi/tcping/v2/internal/options"
	"github.com/pouriyajamshidi/tcping/v2/internal/utils"
	"github.com/pouriyajamshidi/tcping/v2/printers"
	"github.com/pouriyajamshidi/tcping/v2/probes"
	"github.com/pouriyajamshidi/tcping/v2/types"
)

// TODO:
// 1. Perhaps move HostnameChange from types to nameresolution
// 2. Pass Handler instead of tcping to helpers, etc
// 3. Pass a Printer to `newNetworkInterface`
// 4. Probably it is better to move SignalHandler to probes instead of printers
// 5. Make `Options` of `Tcping` implicit too?

func main() {
	tcping := &types.Tcping{}

	options.ProcessUserInput(tcping)

	tcping.StartTime = time.Now()

	tcping.Ticker = time.NewTicker(tcping.Options.IntervalBetweenProbes)
	defer tcping.Ticker.Stop()

	printers.SignalHandler(tcping)

	tcping.PrintStart(tcping.Options.Hostname, tcping.Options.Port)

	stdinchan := make(chan bool)
	go utils.MonitorSTDIN(stdinchan)

	var probeCount uint
	for {
		if tcping.Options.ShouldRetryResolve {
			nameresolution.RetryResolveHostname(tcping)
		}

		probes.Probe(tcping)

		select {
		case pressedEnter := <-stdinchan:
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
