// Package probe runs the probe loop. It holds the TCP, HTTP and UDP probes
// and the Prober that repeats one of them, records what came back in the
// statistics and hands each result to a Printer.
package probe

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/config"
	"github.com/pouriyajamshidi/tcping/v3/nic"
	"github.com/pouriyajamshidi/tcping/v3/stats"
)

// address is the "ip:port" a probe dials.
func address(ip netip.Addr, port uint16) string {
	return net.JoinHostPort(ip.String(), strconv.Itoa(int(port)))
}

// dialer builds the net.Dialer for one probe on the given network ("tcp" or
// "udp"). When a network interface is configured (-I), the connection is
// sourced from it, matching ip's address family. If the interface has no
// address of that family the probe fails here, so it does not dial out of
// some other interface.
func dialer(network string, networkInterface nic.NetworkInterface, timeout time.Duration, ip netip.Addr) (net.Dialer, error) {
	d := net.Dialer{Timeout: timeout}

	if !networkInterface.Use {
		return d, nil
	}

	localIP := networkInterface.LocalIPFor(ip)
	if localIP == nil {
		return net.Dialer{}, fmt.Errorf("network interface has no source address to reach %s", ip)
	}

	// net.Dialer wants the local address to be of the same kind as the
	// network it dials, otherwise the dial fails outright.
	if network == udp {
		d.LocalAddr = &net.UDPAddr{IP: localIP}
	} else {
		d.LocalAddr = &net.TCPAddr{IP: localIP}
	}

	return d, nil
}

type Prober struct {
	pinger     Pinger
	printer    Printer
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
func NewProber(pinger Pinger, printer Printer, cfg config.Config, stats *stats.Statistics, summaryRequests <-chan struct{}) *Prober {
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

	// Filled in by HTTP probes only, left zero by the others. Shared with
	// Statistics so the result can be handed to the printers as one value.
	stats.HTTPInfo

	// Filled in by UDP probes only, left zero by the others.
	stats.UDPInfo
}

type Pinger interface {
	Ping(ctx context.Context, ip netip.Addr) (ProbeResult, error)
}

func (p *Prober) Probe(ctx context.Context) error {
	ticker := time.NewTicker(p.config.IntervalBetweenProbes)
	defer ticker.Stop()

	p.statistics.StartTime = time.Now()
	p.printer.PrintStart(p.statistics)

	var probeCount uint

	// runProbe performs a single probe, prints its result, retries hostname
	// resolution if configured to, and reports whether that was the last
	// probe to run (ProbesBeforeQuit reached).
	runProbe := func() (done bool) {
		p.statistics.ResolvedThisProbe = false

		// Only the probe that ends an uptime or a downtime reports it, so
		// clear them here and let this probe set its own.
		p.statistics.EndedUptime = 0
		p.statistics.EndedDowntime = 0

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
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			p.finalizeStatistics()
			return nil

		case _, ok := <-p.summaryRequests:
			if !ok {
				// stdin is gone, so stop waiting on it. A nil channel
				// never fires again, unlike a closed one.
				p.summaryRequests = nil
				continue
			}

			// The run is still going, so the uptime or downtime it is in
			// the middle of has not been added to the totals yet.
			p.printer.PrintStatistics(p.statistics.SummaryNow())

		case <-ticker.C:
			if runProbe() {
				p.finalizeStatistics()
				return nil
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

// handleProbeFailure records a failed probe. When it is the one that took the
// target from up to down, it fills in Statistics.EndedUptime so the printers
// can report the uptime that just ended along with the probe.
func (p *Prober) handleProbeFailure(pingTime time.Time, probeResult ProbeResult) {
	s := p.statistics

	// A 4xx or 5xx is a failed probe that still came with a response, so
	// keep it around for the printers. It is zero when nothing answered.
	s.HTTP = probeResult.HTTPInfo

	// A failed UDP probe still tells us whether we were refused or simply
	// never answered, which is the whole difference for UDP.
	s.UDP = probeResult.UDPInfo

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

	if s.LastProbeHadFailed {
		return
	}

	// UP -> DOWN
	s.LastProbeHadFailed = true
	s.StartOfDowntime = pingTime

	// Nothing to report on the very first probe: the target was never up.
	if s.StartOfUptime.IsZero() {
		return
	}

	uptimeDuration := pingTime.Sub(s.StartOfUptime)
	s.EndedUptime = uptimeDuration
	s.TotalUptime += uptimeDuration

	stats.SetLongestDuration(
		s.StartOfUptime,
		uptimeDuration,
		&s.LongestUptime,
	)
}

// handleProbeSuccess records a successful probe. When it is the one that
// brought the target back up, it fills in Statistics.EndedDowntime so the
// printers can report the outage that just ended along with the probe.
func (p *Prober) handleProbeSuccess(pingTime time.Time, rtt time.Duration, probeResult ProbeResult) {
	s := p.statistics

	rttMs := stats.DurationToMilliseconds(rtt)

	s.LatestRTT = rttMs
	s.LocalAddr = probeResult.LocalAddr
	s.HTTP = probeResult.HTTPInfo
	s.UDP = probeResult.UDPInfo

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
		s.EndedDowntime = downtimeDuration

		stats.SetLongestDuration(
			s.StartOfDowntime,
			downtimeDuration,
			&s.LongestDowntime,
		)

		s.StartOfUptime = pingTime
	}

	if s.StartOfUptime.IsZero() {
		s.StartOfUptime = pingTime
	}
}

func (p *Prober) finalizeStatistics() {
	p.statistics.EndTime = time.Now()
	p.statistics.CloseOpenPeriod(p.statistics.EndTime)
}
