package probers

import (
	"context"
	"net"
	"net/netip"
	"strconv"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/printers"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
	"github.com/pouriyajamshidi/tcping/v3/internal/utils"
)

const tcp = "tcp"

type Tcping struct {
	dialer *net.Dialer
	ip     netip.Addr
	port   uint16
}

func NewTcping(cfg config.Config) Tcping {
	if cfg.NetworkInterface.Use {
		cfg.NetworkInterface.Dialer.Timeout = cfg.Timeout
		return Tcping{dialer: &cfg.NetworkInterface.Dialer, ip: cfg.IP, port: cfg.Port}
	}

	return Tcping{dialer: &net.Dialer{Timeout: cfg.Timeout}, ip: cfg.IP, port: cfg.Port}
}

func (t *Tcping) address() string {
	return net.JoinHostPort(t.ip.String(), strconv.Itoa(int(t.port)))
}

// handleConnFailure processes failed probes
func handleConnFailure(s *stats.Statistics, p printers.Printer, startTime time.Time, elapsed time.Duration) {
	// if the last probe had succeeded
	if !s.DestWasDown {
		s.StartOfDowntime = startTime
		uptimeDuration := s.StartOfDowntime.Sub(s.StartOfUptime)
		// set longest uptime since it is interrupted
		utils.SetLongestDuration(s.StartOfUptime, uptimeDuration, &s.LongestUptime)
		s.StartOfUptime = time.Time{} // TODO: why are we doing this?
		s.DestWasDown = true
	}

	s.TotalDowntime += elapsed
	s.LastUnsuccessfulProbe = startTime
	s.TotalUnsuccessfulProbes++
	s.OngoingUnsuccessfulProbes++

	p.PrintProbeFailure(s)
}

// handleConnSuccess processes successful probes
func handleConnSuccess(s *stats.Statistics, p printers.Printer, startTime time.Time, elapsed time.Duration, rtt float32, showFailuresOnly bool) {
	if s.DestWasDown {
		s.StartOfUptime = startTime
		downtimeDuration := s.StartOfUptime.Sub(s.StartOfDowntime)
		// set longest downtime since it is interrupted
		utils.SetLongestDuration(s.StartOfDowntime, downtimeDuration, &s.LongestDowntime)
		p.PrintTotalDownTime(s)
		s.StartOfDowntime = time.Time{} // TODO: why are we doing this?
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

func (t Tcping) Ping(ctx context.Context) error {
	conn, err := t.dialer.DialContext(ctx, tcp, t.address())
	if err != nil {
		return err
	}

	defer conn.Close()

	return nil
}

// Ping checks target's availability using TCP
// func (t Tcping) Ping(s *stats.Statistics, p printers.Printer, cfg config.Config) {
// 	t.Ticker = time.NewTicker(cfg.IntervalBetweenProbes)
// 	defer t.Ticker.Stop()

// 	var err error
// 	var conn net.Conn

// 	connStart := time.Now()

// 	if cfg.NetworkInterface.Use {
// 		// The timeout value of this Dialer is set inside the `newNetworkInterface` function
// 		conn, err = cfg.NetworkInterface.Dialer.Dial("tcp", cfg.NetworkInterface.RemoteAddr.String())
// 	} else {
// 		ipAndPort := netip.AddrPortFrom(cfg.IP, cfg.Port)
// 		conn, err = net.DialTimeout("tcp", ipAndPort.String(), cfg.Timeout)
// 	}

// 	connDuration := time.Since(connStart)
// 	elapsed := utils.MaxDuration(connDuration, cfg.IntervalBetweenProbes)

// 	if err != nil {
// 		handleConnFailure(s, p, connStart, elapsed)
// 	} else {
// 		rtt := utils.NanoToMillisecond(connDuration.Nanoseconds())
// 		handleConnSuccess(s, p, connStart, elapsed, rtt, cfg.ShowFailuresOnly)

// 		conn.Close()
// 	}

// 	<-t.Ticker.C

// 	// TODO: Possibly we can drop using Ticker. Need to think more...
// 	// if wait := cfg.IntervalBetweenProbes - time.Since(start); wait > 0 {
// 	// 	time.Sleep(wait)
// 	// }
// }
