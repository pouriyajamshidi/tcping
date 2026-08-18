// Package probers handles the general probing logic
package probers

import (
	"context"
	"errors"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/printers"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
	"github.com/pouriyajamshidi/tcping/v3/internal/utils"
)

var (
	ErrTimeout       = errors.New("timed out waiting for ping")
	ErrPingCompleted = errors.New("ping completed")
)

type Prober struct {
	pinger     Pinger
	printer    printers.Printer
	config     config.Config
	Ticker     *time.Ticker
	Timeout    time.Duration
	Interval   time.Duration
	Statistics stats.Statistics
}

type Pinger interface {
	// Ping(s *stats.Statistics, p printers.Printer, cfg config.Config)
	Ping(ctx context.Context) error
}

func (p *Prober) Probe(ctx context.Context) (stats.Statistics, error) {
	p.Ticker = time.NewTicker(p.Interval)
	defer p.Ticker.Stop()

	timeoutTimer := time.NewTimer(p.Timeout)
	defer timeoutTimer.Stop()

	p.Statistics.StartTime = time.Now()
	p.printer.PrintStart(&p.Statistics)

	var probeCount uint

	for {
		select {

		case <-ctx.Done():
			p.finalizeStatistics()
			return p.Statistics, nil

		case <-timeoutTimer.C:
			p.finalizeStatistics()

			// Graceful completion if we got successful results
			if p.Statistics.Successful > 0 {
				return p.Statistics, nil
			}
			return p.Statistics, ErrTimeout

		case <-p.Ticker.C:
			pingTime := time.Now()
			err := p.pinger.Ping(ctx)
			rtt := time.Since(pingTime)
			if err != nil {
				// Handle failure
				p.Statistics.OngoingSuccessfulProbes = 0
				p.Statistics.OngoingUnsuccessfulProbes++
				p.Statistics.Failed++
				p.Statistics.TotalUnsuccessfulProbes++
				p.Statistics.LastUnsuccessfulProbe = pingTime

				// Track downtime periods
				if !p.Statistics.DestWasDown {
					p.Statistics.DestWasDown = true
					p.Statistics.StartOfDowntime = pingTime
				}

				p.printer.PrintProbeFailure(&p.Statistics)

				// Retry hostname resolution if threshold reached
				if p.config.ShouldRetryResolve && p.Statistics.OngoingUnsuccessfulProbes >= p.config.RetryResolveAfterNFailures {
					p.Statistics.RetriedHostnameLookups++
					p.printer.PrintRetryingToResolve(p.Statistics.Hostname)
					if err := p.config.Resolver.RetryResolveHostname(&p.Statistics); err != nil {
						p.printer.PrintError("%s", err.Error())
					}
				}
			} else {
				// Handle success
				rttMs := utils.NanoToMillisecond(rtt.Nanoseconds())
				p.Statistics.RTT = append(p.Statistics.RTT, rttMs)
				p.Statistics.LatestRTT = rttMs
				p.Statistics.HasResults = true
				p.Statistics.Successful++
				p.Statistics.TotalSuccessfulProbes++
				p.Statistics.OngoingSuccessfulProbes++
				p.Statistics.OngoingUnsuccessfulProbes = 0
				p.Statistics.LastSuccessfulProbe = pingTime

				// Track uptime periods
				if p.Statistics.DestWasDown {
					// Transitioning from down to up
					p.Statistics.DestWasDown = false
					downDuration := pingTime.Sub(p.Statistics.StartOfDowntime)
					p.Statistics.TotalDowntime += downDuration
					p.Statistics.DownTime = downDuration
					utils.SetLongestDuration(p.Statistics.StartOfDowntime, downDuration, &p.Statistics.LongestDown)
					p.Statistics.StartOfUptime = pingTime
					p.printer.PrintTotalDownTime(&p.Statistics)
				}

				if p.Statistics.StartOfUptime.IsZero() {
					p.Statistics.StartOfUptime = pingTime
				}

				p.printer.PrintProbeSuccess(&p.Statistics)
			}

			// Check probe count limit
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

func Run(pinger Pinger, printer printers.Printer, stats *stats.Statistics, cfg config.Config) {
	printer.PrintStart(stats)

	stats.StartTime = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	var probeCount uint

	for {
		if cfg.ShouldRetryResolve && stats.OngoingUnsuccessfulProbes >= cfg.RetryResolveAfterNFailures {
			stats.RetriedHostnameLookups++
			printer.PrintRetryingToResolve(stats.Hostname)
			if err := cfg.Resolver.RetryResolveHostname(stats); err != nil {
				printer.PrintError("%s", err.Error())
			}
		}

		pinger.Ping(ctx)

		// -c flag is provided
		if cfg.ProbesBeforeQuit != 0 {
			probeCount++
			if probeCount == cfg.ProbesBeforeQuit {
				printer.Shutdown(stats)
			}
		}
	}
}

func (p *Prober) finalizeStatistics() {
	p.Statistics.EndTime = time.Now()
	p.Statistics.UpTime = p.Statistics.EndTime.Sub(p.Statistics.StartTime)

	if p.Statistics.DestWasDown {
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

	if s.DestWasDown {
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
