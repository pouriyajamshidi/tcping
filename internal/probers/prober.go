// Package probers handles the general probing logic
package probers

import (
	"context"
	"net"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/printers"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
)

type Prober struct {
	pinger     Pinger
	printer    printers.Printer
	config     config.Config
	Ticker     *time.Ticker
	Statistics *stats.Statistics
}

func NewProber(pinger Pinger, printer printers.Printer, cfg config.Config, stats *stats.Statistics) *Prober {
	pr := Prober{
		pinger:     pinger,
		printer:    printer,
		config:     cfg,
		Statistics: stats,
	}

	return &pr
}

type ProbeResult struct {
	LocalAddr net.Addr
}

type Pinger interface {
	Ping(ctx context.Context) (ProbeResult, error)
}

func (p *Prober) Probe(ctx context.Context) (*stats.Statistics, error) {
	p.Ticker = time.NewTicker(p.config.IntervalBetweenProbes)
	defer p.Ticker.Stop()

	p.Statistics.StartTime = time.Now()
	p.printer.PrintStart(p.Statistics)

	var probeCount uint

	// runProbe performs a single probe, prints its result, retries hostname
	// resolution if configured to, and reports whether that was the last
	// probe to run (ProbesBeforeQuit reached).
	// we need this to avoid waiting n seconds for the first probe to run.
	runProbe := func() (done bool) {
		pingTime := time.Now()

		probeResult, err := p.pinger.Ping(ctx)
		rtt := time.Since(pingTime)

		if err != nil {
			p.handleProbeFailure(pingTime)
			p.printer.PrintProbeFailure(p.Statistics)
		} else {
			p.handleProbeSuccess(pingTime, rtt, probeResult)
			p.printer.PrintProbeSuccess(p.Statistics)
		}

		if p.config.ShouldRetryResolve &&
			p.Statistics.OngoingUnsuccessfulProbes >= p.config.RetryResolveAfterNFailures {

			p.Statistics.RetriedHostnameLookups++

			p.printer.PrintRetryingToResolve(p.Statistics.Hostname)

			if err := p.config.Resolver.RetryResolveHostname(p.Statistics); err != nil {
				p.printer.PrintError("%s", err.Error())
			}
		}

		if p.config.ProbesBeforeQuit > 0 {
			probeCount++

			if probeCount >= p.config.ProbesBeforeQuit {
				return true
			}
		}

		return false
	}

	// Probe immediately instead of waiting for the ticker's first tick.
	if runProbe() {
		p.finalizeStatistics()
		return p.Statistics, nil
	}

	for {
		select {
		case <-ctx.Done():
			p.finalizeStatistics()
			return p.Statistics, nil

		case <-p.Ticker.C:
			if runProbe() {
				p.finalizeStatistics()
				return p.Statistics, nil
			}
		}
	}
}

func (p *Prober) handleProbeFailure(pingTime time.Time) {
	s := p.Statistics

	s.OngoingSuccessfulProbes = 0
	s.OngoingUnsuccessfulProbes++
	s.TotalUnsuccessfulProbes++
	s.LastUnsuccessfulProbe = pingTime

	if p.config.NetworkInterface.Use {
		s.LocalAddr = p.config.NetworkInterface.Dialer.LocalAddr
	}

	if !s.LastProbeHadFailed {
		// UP -> DOWN
		s.LastProbeHadFailed = true
		s.StartOfDowntime = pingTime

		uptimeDuration := pingTime.Sub(s.StartOfUptime)

		// TODO: what do we do with this?
		s.CurrentUptime = uptimeDuration

		if !s.StartOfUptime.IsZero() {
			s.TotalUptime += uptimeDuration

			stats.SetLongestDuration(
				s.StartOfUptime,
				uptimeDuration,
				&s.LongestUptime,
			)
		}
	}
}

func (p *Prober) handleProbeSuccess(pingTime time.Time, rtt time.Duration, probeResult ProbeResult) {
	s := p.Statistics

	rttMs := NanoToMillisecond(rtt.Nanoseconds())

	s.LatestRTT = rttMs
	s.LocalAddr = probeResult.LocalAddr

	s.TotalSuccessfulProbes++
	s.OngoingSuccessfulProbes++
	s.OngoingUnsuccessfulProbes = 0
	s.LastSuccessfulProbe = pingTime

	s.RTTResults.Update(rttMs, s.TotalSuccessfulProbes)

	if s.LastProbeHadFailed {
		// DOWN -> UP
		s.LastProbeHadFailed = false

		downtimeDuration := pingTime.Sub(s.StartOfDowntime)

		s.TotalDowntime += downtimeDuration
		s.CurrentDowntime = downtimeDuration

		stats.SetLongestDuration(
			s.StartOfDowntime,
			downtimeDuration,
			&s.LongestDowntime,
		)

		p.printer.PrintDownTimeDuration(s)

		s.StartOfUptime = pingTime
	}

	if s.StartOfUptime.IsZero() {
		s.StartOfUptime = pingTime
	}
}

// TODO: this should replace the ShutDown and PrintStats methods
func (p *Prober) finalizeStatistics() {
	p.Statistics.EndTime = time.Now()

	if p.Statistics.LastProbeHadFailed {
		downDuration := p.Statistics.EndTime.Sub(p.Statistics.StartOfDowntime)
		p.Statistics.TotalDowntime += downDuration
		stats.SetLongestDuration(p.Statistics.StartOfDowntime, downDuration, &p.Statistics.LongestDowntime)
		return
	}

	if !p.Statistics.StartOfUptime.IsZero() {
		upDuration := p.Statistics.EndTime.Sub(p.Statistics.StartOfUptime)
		p.Statistics.TotalUptime += upDuration
		stats.SetLongestDuration(p.Statistics.StartOfUptime, upDuration, &p.Statistics.LongestUptime)
	}
}

// NanoToMillisecond returns an amount of milliseconds from nanoseconds.
// Using duration.Milliseconds() is not an option, because it drops
// decimal points, returning an int.
func NanoToMillisecond(nano int64) float32 {
	return float32(nano) / float32(time.Millisecond)
}
