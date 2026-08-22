// Package probers handles the general probing logic
package probers

import (
	"context"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/printers"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
	"github.com/pouriyajamshidi/tcping/v3/internal/utils"
)

type Prober struct {
	pinger     Pinger
	printer    printers.Printer
	config     config.Config
	Ticker     *time.Ticker
	Statistics *stats.Statistics
}

func NewProber(p Pinger, cfg config.Config) *Prober {
	pr := Prober{
		pinger:     p,
		printer:    printers.NewColorPrinter(),
		config:     cfg,
		Statistics: stats.NewStatistics(cfg),
	}

	return &pr
}

type Pinger interface {
	Ping(ctx context.Context) error
}

func (p *Prober) Probe(ctx context.Context) (*stats.Statistics, error) {
	p.Ticker = time.NewTicker(p.config.IntervalBetweenProbes)
	defer p.Ticker.Stop()

	p.Statistics.StartTime = time.Now()
	p.printer.PrintStart(p.Statistics)

	var probeCount uint

	for {
		select {
		case <-ctx.Done():
			p.finalizeStatistics()
			return p.Statistics, nil

		case <-p.Ticker.C:
			pingTime := time.Now()

			err := p.pinger.Ping(ctx)
			rtt := time.Since(pingTime)

			if err != nil {
				p.handleProbeFailure(pingTime)
				p.printer.PrintProbeFailure(p.Statistics)
			} else {
				p.handleProbeSuccess(pingTime, rtt)
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
					p.finalizeStatistics()
					return p.Statistics, nil
				}
			}
		}
	}
}

func (p *Prober) ProbeV2(ctx context.Context) (*stats.Statistics, error) {
	var probeCount uint

	probe := func() {
		probeCount++

		pingTime := time.Now()

		err := p.pinger.Ping(ctx)
		rtt := time.Since(pingTime)

		if err != nil {
			p.handleProbeFailure(pingTime)
		} else {
			p.handleProbeSuccess(pingTime, rtt)
		}
	}

	p.Ticker = time.NewTicker(p.config.IntervalBetweenProbes)
	defer p.Ticker.Stop()

	p.Statistics.StartTime = time.Now()
	p.printer.PrintStart(p.Statistics)

	// so we do not wait the p.Ticker.C to then start probing
	probe()

	for {
		select {
		case <-ctx.Done():
			p.finalizeStatistics()
			return p.Statistics, nil

		case <-p.Ticker.C:
			probe()

			if p.config.ProbesBeforeQuit > 0 {
				if probeCount >= p.config.ProbesBeforeQuit {
					p.finalizeStatistics()
					return p.Statistics, nil
				}
			}
		}
	}
}

func (p *Prober) handleProbeFailure(pingTime time.Time) {
	s := p.Statistics

	s.OngoingSuccessfulProbes = 0
	s.OngoingUnsuccessfulProbes++
	s.Failed++
	s.TotalUnsuccessfulProbes++
	s.LastUnsuccessfulProbe = pingTime

	if !s.LastProbeHadFailed {
		// UP -> DOWN
		s.LastProbeHadFailed = true
		s.StartOfDowntime = pingTime

		if !s.StartOfUptime.IsZero() {
			uptimeDuration := pingTime.Sub(s.StartOfUptime)
			s.TotalUptime += uptimeDuration

			utils.SetLongestDuration(
				s.StartOfUptime,
				uptimeDuration,
				&s.LongestUptime,
			)
		}
	}
}

func (p *Prober) handleProbeSuccess(pingTime time.Time, rtt time.Duration) {
	s := p.Statistics

	rttMs := utils.NanoToMillisecond(rtt.Nanoseconds())

	s.RTT = append(s.RTT, rttMs)
	s.LatestRTT = rttMs
	s.HasResults = true

	s.Successful++
	s.TotalSuccessfulProbes++
	s.OngoingSuccessfulProbes++
	s.OngoingUnsuccessfulProbes = 0
	s.LastSuccessfulProbe = pingTime

	if s.LastProbeHadFailed {
		// DOWN -> UP
		s.LastProbeHadFailed = false

		downtimeDuration := pingTime.Sub(s.StartOfDowntime)

		s.TotalDowntime += downtimeDuration
		s.DownTime = downtimeDuration

		utils.SetLongestDuration(
			s.StartOfDowntime,
			downtimeDuration,
			&s.LongestDown,
		)

		p.printer.PrintTotalDownTime(s)

		s.StartOfUptime = pingTime
	}

	if s.StartOfUptime.IsZero() {
		s.StartOfUptime = pingTime
	}
}

func (p *Prober) finalizeStatistics() {
	p.Statistics.EndTime = time.Now()
	p.Statistics.UpTime = p.Statistics.EndTime.Sub(p.Statistics.StartTime)

	if p.Statistics.LastProbeHadFailed {
		downDuration := p.Statistics.EndTime.Sub(p.Statistics.StartOfDowntime)
		p.Statistics.TotalDowntime += downDuration
		utils.SetLongestDuration(p.Statistics.StartOfDowntime, downDuration, &p.Statistics.LongestDown)
		return
	}

	if !p.Statistics.StartOfUptime.IsZero() {
		upDuration := p.Statistics.EndTime.Sub(p.Statistics.StartOfUptime)
		p.Statistics.TotalUptime += upDuration
		utils.SetLongestDuration(p.Statistics.StartOfUptime, upDuration, &p.Statistics.LongestUp)
	}
}

func finalizeStatistics(s *stats.Statistics) {
	s.EndTime = time.Now()
	s.UpTime = s.EndTime.Sub(s.StartTime)

	if s.LastProbeHadFailed {
		downDuration := s.EndTime.Sub(s.StartOfDowntime)
		s.TotalDowntime += downDuration
		utils.SetLongestDuration(s.StartOfDowntime, downDuration, &s.LongestDown)
		return
	}

	if !s.StartOfUptime.IsZero() {
		upDuration := s.EndTime.Sub(s.StartOfUptime)
		s.TotalUptime += upDuration
		utils.SetLongestDuration(s.StartOfUptime, upDuration, &s.LongestUp)
	}
}
