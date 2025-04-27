package main

import (
	"context"
	"log"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/probes"
	"github.com/pouriyajamshidi/tcping/v2/probes/pinger"
	"github.com/pouriyajamshidi/tcping/v2/probes/printer"
)

func main() {
	printer := printer.NewColor()

	pinger := pinger.NewTCPPinger(
		pinger.WithIP("159.89.251.4"),
		pinger.WithPort(80))

	prober := probes.NewProber(pinger,
		probes.WithInterval(1*time.Second),
		probes.WithTimeout(5*time.Second),
		probes.WithPrinter(printer),
	)

	ctx := context.Background()
	statistics, err := prober.Probe(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printer.PrintStatistics(statistics)

}
