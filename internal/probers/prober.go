// Package probers handles the general probing logic
package probers

import (
	"context"
	"net"
	"net/netip"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/printers"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
)

type Prober struct {
	pinger     Pinger
	printer    printers.Printer
	config     config.Config
	statistics *stats.Statistics

	// Asks for the statistics to be printed mid-run. Reading them here,
	// between probes, is what keeps them from being read while a probe is
	// busy updating them. Nil when nobody can make the request.
	summaryRequests <-chan struct{}
}

// NewProber wires up a prober. summaryRequests may be nil; when it is not,
// the statistics are printed every time it yields a value (see
// app.SummaryRequests).
func NewProber(pinger Pinger, printer printers.Printer, cfg config.Config, stats *stats.Statistics, summaryRequests <-chan struct{}) *Prober {
	pr := Prober{
		pinger:          pinger,
		printer:         printer,
		config:          cfg,
		statistics:      stats,
		summaryRequests: summaryRequests,
	}

	return &pr
}

type ProbeResult struct {
	LocalAddr net.Addr

	// Filled in by HTTP probes only, left zero by TCP ones. Shared with
	// Statistics so the result can be handed to the printers as one value.
	stats.HTTPInfo
}

type Pinger interface {
	Ping(ctx context.Context, ip netip.Addr) (ProbeResult, error)
}

func (p *Prober) Probe(ctx context.Context) (*stats.Statistics, error) {
	ticker := time.NewTicker(p.config.IntervalBetweenProbes)
	defer ticker.Stop()

	p.statistics.StartTime = time.Now()
	p.printer.PrintStart(p.statistics)

	var probeCount uint

	// runProbe performs a single probe, prints its result, retries hostname
	// resolution if configured to, and reports whether that was the last
	// probe to run (ProbesBeforeQuit reached).
	// we need this to avoid waiting n seconds for the first probe to run.
	runProbe := func() (done bool) {
		p.statistics.ResolvedThisProbe = false

		// Resolve the hostname fresh before this probe when requested,
		// so it always dials whatever the hostname currently points to
		// (e.g. DNS round-robin or a frequently-changing record) rather
		// than waiting for a failure streak to trigger a retry.
		if p.config.ResolveEveryProbe && !p.statistics.DestIsIP {
			p.statistics.RetriedHostnameLookups++
			p.resolveHostname(true)
		}

		pingTime := time.Now()

		// Read the target IP fresh on every probe (not just once at
		// startup), so a hostname-retry-resolve that changes it - possibly
		// even to a different address family - actually takes effect on
		// the next probe instead of being silently ignored.
		probeResult, err := p.pinger.Ping(ctx, p.statistics.IP)
		rtt := time.Since(pingTime)

		// A probe that only failed because we cancelled it ourselves, when
		// the user hits Ctrl+C while it is in flight, is not a failed probe.
		// Counting it would report a failure, a packet loss and a downtime
		// that never happened.
		if err != nil && ctx.Err() != nil {
			return true
		}

		if err != nil {
			p.handleProbeFailure(pingTime, probeResult)
			p.printer.PrintProbeFailure(p.statistics)
		} else {
			p.handleProbeSuccess(pingTime, rtt, probeResult)

			// The probe is still counted, we just do not report it.
			if !p.config.ShowFailuresOnly {
				p.printer.PrintProbeSuccess(p.statistics)
			}
		}

		if !p.config.ResolveEveryProbe && p.config.ShouldRetryResolve &&
			p.statistics.OngoingUnsuccessfulProbes >= p.config.RetryResolveAfterNFailures {

			p.statistics.RetriedHostnameLookups++

			p.printer.PrintRetryingToResolve(p.statistics.Hostname)

			p.resolveHostname(false)
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
		return p.statistics, nil
	}

	for {
		select {
		case <-ctx.Done():
			p.finalizeStatistics()
			return p.statistics, nil

		case _, ok := <-p.summaryRequests:
			if !ok {
				// stdin is gone, so stop waiting on it. A nil channel
				// never fires again, unlike a closed one.
				p.summaryRequests = nil
				continue
			}

			p.printer.PrintStatistics(p.statistics)

		case <-ticker.C:
			if runProbe() {
				p.finalizeStatistics()
				return p.statistics, nil
			}
		}
	}
}

// resolveHostname resolves the target's hostname again, recording how long
// it took and appending a HostnameChange when the address actually changed.
// It prints the resulting duration on success or the error on failure. When
// markResolvedThisProbe is true and resolution succeeds, it sets
// Statistics.ResolvedThisProbe before printing, so PrintProbeSuccess/
// PrintProbeFailure fold the duration into their own line instead of a
// separate one. It reports whether the resolution succeeded.
func (p *Prober) resolveHostname(markResolvedThisProbe bool) bool {
	s := p.statistics

	start := time.Now()
	newIP, err := p.config.Resolver.ResolveHostname(s.Hostname)
	duration := time.Since(start)
	if err != nil {
		p.printer.PrintError("%s", err.Error())
		return false
	}

	s.IP = newIP
	s.NameResolutionDuration = duration

	if len(s.HostnameChanges) == 0 || s.HostnameChanges[len(s.HostnameChanges)-1].Addr != newIP {
		s.HostnameChanges = append(s.HostnameChanges, stats.HostnameChange{
			Addr:     newIP,
			When:     time.Now(),
			Duration: duration,
		})
	}

	if markResolvedThisProbe {
		s.ResolvedThisProbe = true
	}
	p.printer.PrintNameResolutionDuration(s)

	return true
}

func (p *Prober) handleProbeFailure(pingTime time.Time, probeResult ProbeResult) {
	s := p.statistics

	// A 4xx or 5xx is a failed probe that still came with a response, so
	// keep it around for the printers. It is zero when nothing answered.
	s.HTTP = probeResult.HTTPInfo

	s.OngoingSuccessfulProbes = 0
	s.OngoingUnsuccessfulProbes++
	s.TotalUnsuccessfulProbes++
	s.LastUnsuccessfulProbe = pingTime

	if p.config.NetworkInterface.Use {
		localIP := p.config.NetworkInterface.LocalIPFor(s.IP)
		if localIP != nil {
			s.LocalAddr = &net.TCPAddr{IP: localIP}
		} else {
			s.LocalAddr = nil
		}
	}

	if !s.LastProbeHadFailed {
		// UP -> DOWN
		s.LastProbeHadFailed = true
		s.StartOfDowntime = pingTime

		uptimeDuration := pingTime.Sub(s.StartOfUptime)
		s.CurrentUptime = uptimeDuration

		if !s.StartOfUptime.IsZero() {
			s.TotalUptime += uptimeDuration

			stats.SetLongestDuration(
				s.StartOfUptime,
				uptimeDuration,
				&s.LongestUptime,
			)

			p.printer.PrintUpTimeDuration(s)
		}
	}
}

func (p *Prober) handleProbeSuccess(pingTime time.Time, rtt time.Duration, probeResult ProbeResult) {
	s := p.statistics

	rttMs := stats.DurationToMilliseconds(rtt)

	s.LatestRTT = rttMs
	s.LocalAddr = probeResult.LocalAddr
	s.HTTP = probeResult.HTTPInfo

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

func (p *Prober) finalizeStatistics() {
	p.statistics.EndTime = time.Now()

	if p.statistics.LastProbeHadFailed {
		downDuration := p.statistics.EndTime.Sub(p.statistics.StartOfDowntime)
		p.statistics.TotalDowntime += downDuration
		stats.SetLongestDuration(p.statistics.StartOfDowntime, downDuration, &p.statistics.LongestDowntime)
		return
	}

	if !p.statistics.StartOfUptime.IsZero() {
		upDuration := p.statistics.EndTime.Sub(p.statistics.StartOfUptime)
		p.statistics.TotalUptime += upDuration
		stats.SetLongestDuration(p.statistics.StartOfUptime, upDuration, &p.statistics.LongestUptime)
	}
}
