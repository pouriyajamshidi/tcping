package main

import (
	"context"
	"net/netip"

	"github.com/pouriyajamshidi/tcping/v3"
	"github.com/pouriyajamshidi/tcping/v3/pingers"
	"github.com/pouriyajamshidi/tcping/v3/printers"
)

func main() {
	ip, err := netip.ParseAddr("161.35.175.61")
	if err != nil {
		tcping.HandleExit(err)
	}
	port := uint16(80)
	prober := tcping.NewProber(pingers.NewTCPPinger(ip, port))
	stats, err := prober.Probe(context.Background())
	if err != nil {
		tcping.HandleExit(err)
	}
	printer := printers.NewColorPrinter()
	printer.PrintStatistics(&stats)
}

// tcping := &tcping.Result{}
// stats := &statistics.Statistics{}

// printer := input.ProcessUserInput(tcping, stats)

// printer.PrintStart(stats)

// tcping.StartTime = time.Now()

// stats.IP = dns.ResolveHostname(printer, stats, true, false)

// tcping.Ticker = time.NewTicker(tcping.Settings.IntervalBetweenProbes)
// defer tcping.Ticker.Stop()

// printers.SignalHandler(printer, stats)

// if !tcping.Settings.NonInteractive {
// 	go monitorStatsRequest(printer, stats)
// }

// var probeCount uint

// for {
// 	if tcping.Settings.ShouldRetryResolve {
// 		dns.RetryResolveHostname(printer, stats, 300, true, false)
// 	}

// 	probes.Ping(stats, printer, tcping)

// 	if tcping.Settings.ProbesBeforeQuit != 0 {
// 		probeCount++
// 		if probeCount == tcping.Settings.ProbesBeforeQuit {
// 			printer.Shutdown(stats)
// 		}
// 	}
// }
