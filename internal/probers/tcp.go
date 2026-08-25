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
	if !s.LastProbeHadFailed {
		s.StartOfDowntime = startTime
		uptimeDuration := s.StartOfDowntime.Sub(s.StartOfUptime)
		// set longest uptime since it is interrupted
		stats.SetLongestDuration(s.StartOfUptime, uptimeDuration, &s.LongestUptime)
		s.StartOfUptime = time.Time{} // TODO: why are we doing this?
		s.LastProbeHadFailed = true
	}

	s.TotalDowntime += elapsed
	s.LastUnsuccessfulProbe = startTime
	s.TotalUnsuccessfulProbes++
	s.OngoingUnsuccessfulProbes++

	p.PrintProbeFailure(s)
}

// handleConnSuccess processes successful probes
func handleConnSuccess(s *stats.Statistics, p printers.Printer, startTime time.Time, elapsed time.Duration, rtt float32, showFailuresOnly bool) {
	if s.LastProbeHadFailed {
		s.StartOfUptime = startTime
		downtimeDuration := s.StartOfUptime.Sub(s.StartOfDowntime)
		// set longest downtime since it is interrupted
		stats.SetLongestDuration(s.StartOfDowntime, downtimeDuration, &s.LongestDowntime)
		p.PrintTotalDownTime(s)
		s.StartOfDowntime = time.Time{} // TODO: why are we doing this?
		s.LastProbeHadFailed = false
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
