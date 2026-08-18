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
	Statistics *stats.Statistics
}

const (
	DefaultInterval = 1 * time.Second
	DefaultTimeout  = 5 * time.Second
)

func NewProber(p Pinger, cfg config.Config) *Prober {
	pr := Prober{
		pinger:     p,
		printer:    printers.NewColorPrinter(),
		Interval:   DefaultInterval,
		Timeout:    DefaultTimeout,
		Statistics: stats.NewStatistics(cfg),
	}

	return &pr
}

type Pinger interface {
	Ping(ctx context.Context) error
}

func (p *Prober) Probe(ctx context.Context) (*stats.Statistics, error) {
	p.Ticker = time.NewTicker(p.Interval)
	defer p.Ticker.Stop()

	timeoutTimer := time.NewTimer(p.Timeout)
	defer timeoutTimer.Stop()

	p.Statistics.StartTime = time.Now()
	p.printer.PrintStart(p.Statistics)

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

				p.printer.PrintProbeFailure(p.Statistics)

				// Retry hostname resolution if threshold reached
				if p.config.ShouldRetryResolve && p.Statistics.OngoingUnsuccessfulProbes >= p.config.RetryResolveAfterNFailures {
					p.Statistics.RetriedHostnameLookups++
					p.printer.PrintRetryingToResolve(p.Statistics.Hostname)
					if err := p.config.Resolver.RetryResolveHostname(p.Statistics); err != nil {
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
					p.printer.PrintTotalDownTime(p.Statistics)
				}

				if p.Statistics.StartOfUptime.IsZero() {
					p.Statistics.StartOfUptime = pingTime
				}

				p.printer.PrintProbeSuccess(p.Statistics)
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

func Run(ctx context.Context, pinger Pinger, printer printers.Printer, stats *stats.Statistics, cfg config.Config) (*stats.Statistics, error) {
	probeTicker := time.NewTicker(cfg.IntervalBetweenProbes)
	defer probeTicker.Stop()

	timeoutTimer := time.NewTimer(cfg.Timeout)
	defer timeoutTimer.Stop()

	stats.StartTime = time.Now()
	printer.PrintStart(stats)

	var probeCount uint

	for {
		select {

		case <-ctx.Done():
			finalizeStatistics(stats)
			return stats, nil

		case <-timeoutTimer.C:
			finalizeStatistics(stats)

			// Graceful completion if we got successful results
			if stats.Successful > 0 {
				return stats, nil
			}
			return stats, ErrTimeout

		case <-probeTicker.C:
			if cfg.ShouldRetryResolve && stats.OngoingUnsuccessfulProbes >= cfg.RetryResolveAfterNFailures {
				stats.RetriedHostnameLookups++
				printer.PrintRetryingToResolve(stats.Hostname)
				if err := cfg.Resolver.RetryResolveHostname(stats); err != nil {
					printer.PrintError("%s", err.Error())
				}
			}

			pingTime := time.Now()
			err := pinger.Ping(ctx)
			rtt := time.Since(pingTime)
			if err == nil {
				// if the last probe had succeeded
				if !stats.DestWasDown {
					stats.StartOfDowntime = pingTime
					uptimeDuration := stats.StartOfDowntime.Sub(stats.StartOfUptime)
					// set longest uptime since it is interrupted
					utils.SetLongestDuration(stats.StartOfUptime, uptimeDuration, &stats.LongestUptime)
					stats.StartOfUptime = time.Time{} // TODO: why are we doing this?
					stats.DestWasDown = true
				}

				stats.TotalDowntime += rtt
				stats.LastUnsuccessfulProbe = pingTime
				stats.TotalUnsuccessfulProbes++
				stats.OngoingUnsuccessfulProbes++

				printer.PrintProbeSuccess(stats)
			} else {
				if stats.DestWasDown {
					stats.StartOfUptime = pingTime
					downtimeDuration := stats.StartOfUptime.Sub(stats.StartOfDowntime)
					// set longest downtime since it is interrupted
					utils.SetLongestDuration(stats.StartOfDowntime, downtimeDuration, &stats.LongestDowntime)
					printer.PrintTotalDownTime(stats)
					stats.StartOfDowntime = time.Time{} // TODO: why are we doing this?
					stats.DestWasDown = false
					stats.OngoingUnsuccessfulProbes = 0
					stats.OngoingSuccessfulProbes = 0
				}

				if stats.StartOfUptime.IsZero() {
					stats.StartOfUptime = pingTime
				}

				stats.TotalUptime += rtt
				stats.LastSuccessfulProbe = pingTime
				stats.TotalSuccessfulProbes++
				stats.OngoingSuccessfulProbes++
				rttMs := utils.NanoToMillisecond(rtt.Nanoseconds())
				stats.RTT = append(stats.RTT, rttMs)
				stats.LatestRTT = rttMs

				if cfg.ShowFailuresOnly {
					continue
				}

				printer.PrintProbeSuccess(stats)
			}

			// -c flag is provided
			if cfg.ProbesBeforeQuit != 0 {
				probeCount++
				if probeCount == cfg.ProbesBeforeQuit {
					printer.Shutdown(stats)
				}
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
