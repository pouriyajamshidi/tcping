package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/probes"
	"github.com/pouriyajamshidi/tcping/v2/probes/pinger"
	"github.com/pouriyajamshidi/tcping/v2/probes/printer"
)

func main() {
	printer := printer.NewColor()
	hostname := "159.89.250.4"
	port := uint16(80)
	printer.Print(fmt.Sprintf("TCPinging %s on port %d\n", hostname, port))
	pinger := pinger.NewTCPPinger(
		pinger.WithIP(hostname),
		pinger.WithPort(port))

	prober := probes.NewProber(pinger,
		probes.WithInterval(1*time.Second),
		probes.WithTimeout(10*time.Second),
		probes.WithPrinter(printer),
	)

	ctx := context.Background()
	statistics, err := prober.Probe(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printer.PrintStatistics(statistics)
}
