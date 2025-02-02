// Package probes provides the logic to ping a target using
// a particular protocol
package probes

import (
	"net"
	"net/netip"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/internal/utils"
	"github.com/pouriyajamshidi/tcping/v2/printers"
	"github.com/pouriyajamshidi/tcping/v2/types"
)

// handleConnError processes failed probes
func handleConnError(t *types.Tcping, connTime time.Time, elapsed time.Duration) {
	if !t.DestWasDown {
		t.StartOfDowntime = connTime
		uptime := t.StartOfDowntime.Sub(t.StartOfUptime)
		printers.CalcLongestUptime(t, uptime)
		t.StartOfUptime = time.Time{}
		t.DestWasDown = true
	}

	t.TotalDowntime += elapsed
	t.LastUnsuccessfulProbe = connTime
	t.TotalUnsuccessfulProbes++
	t.OngoingUnsuccessfulProbes++

	t.PrintProbeFail(
		t.Options,
		t.OngoingUnsuccessfulProbes,
	)
}

// handleConnSuccess processes successful probes
func handleConnSuccess(t *types.Tcping, sourceAddr string, rtt float32, connTime time.Time, elapsed time.Duration) {
	if t.DestWasDown {
		t.StartOfUptime = connTime
		downtime := t.StartOfUptime.Sub(t.StartOfDowntime)
		printers.CalcLongestDowntime(t, downtime)
		t.PrintTotalDownTime(downtime)
		t.StartOfDowntime = time.Time{}
		t.DestWasDown = false
		t.OngoingUnsuccessfulProbes = 0
		t.OngoingSuccessfulProbes = 0
	}

	if t.StartOfUptime.IsZero() {
		t.StartOfUptime = connTime
	}

	t.TotalUptime += elapsed
	t.LastSuccessfulProbe = connTime
	t.TotalSuccessfulProbes++
	t.OngoingSuccessfulProbes++
	t.Rtt = append(t.Rtt, rtt)

	if !t.Options.ShowFailuresOnly {
		t.PrintProbeSuccess(
			sourceAddr,
			t.Options,
			t.OngoingSuccessfulProbes,
			rtt,
		)
	}
}

// Probe pings a host using TCP
func Probe(tcping *types.Tcping) {
	var err error
	var conn net.Conn
	connStart := time.Now()

	if tcping.Options.NetworkInterface.Use {
		// dialer already contains the timeout value
		conn, err = tcping.Options.NetworkInterface.Dialer.Dial("tcp", tcping.Options.NetworkInterface.RemoteAddr.String())
	} else {
		ipAndPort := netip.AddrPortFrom(tcping.Options.IP, tcping.Options.Port)
		conn, err = net.DialTimeout("tcp", ipAndPort.String(), tcping.Options.Timeout)
	}

	connDuration := time.Since(connStart)
	rtt := utils.NanoToMillisecond(connDuration.Nanoseconds())

	elapsed := utils.MaxDuration(connDuration, tcping.Options.IntervalBetweenProbes)

	if err != nil {
		handleConnError(tcping, connStart, elapsed)
	} else {
		handleConnSuccess(tcping, conn.LocalAddr().String(), rtt, connStart, elapsed)
		conn.Close()
	}
	<-tcping.Ticker.C
}
