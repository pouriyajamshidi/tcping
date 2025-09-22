// Package probes provides the logic to ping a target using
// a particular protocol
package probes

import (
	"net"
	"net/netip"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/utils"
	"github.com/pouriyajamshidi/tcping/v3/printers"
	"github.com/pouriyajamshidi/tcping/v3/probes/statistics"
	"github.com/pouriyajamshidi/tcping/v3/types"
)

// handleConnFailure processes failed probes
func handleConnFailure(s *statistics.Statistics, p printers.Printer, startTime time.Time, elapsed time.Duration) {
	// if the last probe had succeeded
	if !s.DestWasDown {
		s.StartOfDowntime = startTime
		uptimeDuration := s.StartOfDowntime.Sub(s.StartOfUptime)
		// set longest uptime since it is interrupted
		utils.SetLongestDuration(s.StartOfUptime, uptimeDuration, &s.LongestUptime)
		s.StartOfUptime = time.Time{}
		s.DestWasDown = true
	}

	s.TotalDowntime += elapsed
	s.LastUnsuccessfulProbe = startTime
	s.TotalUnsuccessfulProbes++
	s.OngoingUnsuccessfulProbes++

	p.PrintProbeFailure(s)
}

// handleConnSuccess processes successful probes
func handleConnSuccess(s *statistics.Statistics, p printers.Printer, startTime time.Time, elapsed time.Duration, rtt float32, showFailuresOnly bool) {
	if s.DestWasDown {
		s.StartOfUptime = startTime
		downtimeDuration := s.StartOfUptime.Sub(s.StartOfDowntime)
		// set longest downtime since it is interrupted
		utils.SetLongestDuration(s.StartOfDowntime, downtimeDuration, &s.LongestDowntime)
		p.PrintTotalDownTime(s)
		s.StartOfDowntime = time.Time{}
		s.DestWasDown = false
		s.OngoingUnsuccessfulProbes = 0
		s.OngoingSuccessfulProbes = 0
	}

	if s.StartOfUptime.IsZero() {
		s.StartOfUptime = startTime
	}

	s.TotalUptime += elapsed
	s.LastSuccessfulProbe = startTime
	s.TotalSuccessfulProbes++
	s.OngoingSuccessfulProbes++
	s.RTT = append(s.RTT, rtt)
	s.LatestRTT = rtt

	if showFailuresOnly {
		return
	}

	p.PrintProbeSuccess(s)
}

// Ping checks target's availability using TCP
func Ping(s *statistics.Statistics, p printers.Printer, tcping *types.Tcping) {
	var err error
	var conn net.Conn

	connStart := time.Now()

	if tcping.Options.NetworkInterface.Use {
		// The timeout value of this Dialer is set inside the `newNetworkInterface` function
		conn, err = tcping.Options.NetworkInterface.Dialer.Dial("tcp", tcping.Options.NetworkInterface.RemoteAddr.String())
	} else {
		ipAndPort := netip.AddrPortFrom(s.IP, s.Port)
		conn, err = net.DialTimeout("tcp", ipAndPort.String(), tcping.Options.Timeout)
	}

	connDuration := time.Since(connStart)
	elapsed := utils.MaxDuration(connDuration, tcping.Options.IntervalBetweenProbes)

	if err != nil {
		handleConnFailure(s, p, connStart, elapsed)
	} else {
		rtt := utils.NanoToMillisecond(connDuration.Nanoseconds())
		handleConnSuccess(s, p, connStart, elapsed, rtt, false)

		conn.Close()
	}

	<-tcping.Ticker.C
}
