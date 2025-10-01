package tcping

import (
	"context"
	"errors"
	"net/netip"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/options"
	"github.com/pouriyajamshidi/tcping/v3/printers"
	"github.com/pouriyajamshidi/tcping/v3/statistics"
)

var (
	ErrTimeout = errors.New("timed out waiting for ping")
)

// Prober orchestrates periodic connectivity testing with configurable timing and output.
type Prober struct {
	pinger     Pinger
	printer    Printer
	Ticker     *time.Ticker
	Timeout    time.Duration
	Interval   time.Duration
	Statistics statistics.Statistics
}

type ProberOption = options.Option[Prober]

// WithInterval configures the interval between probe attempts.
func WithInterval(interval time.Duration) ProberOption {
	return func(p *Prober) {
		p.Interval = interval
	}
}

// WithTimeout configures the timeout duration for probe attempts.
func WithTimeout(timeout time.Duration) ProberOption {
	return func(p *Prober) {
		p.Timeout = timeout + p.Interval
	}
}

// WithPrinter configures the printer for probe output formatting.
func WithPrinter(printer Printer) ProberOption {
	return func(p *Prober) {
		p.printer = printer
	}
}

// NewProber creates a new prober with the given pinger and optional configuration.
func NewProber(p Pinger, opts ...ProberOption) *Prober {
	pr := Prober{
		pinger:   p,
		printer:  printers.NewColorPrinter(),
		Interval: DefaultInterval,
		Timeout:  DefaultTimeout,
	}

	// Initialize statistics with pinger details
	pr.Statistics.Hostname = p.IP()
	pr.Statistics.Port = p.Port()
	pr.Statistics.Protocol = "TCP"

	// Parse and set the IP address
	if ip, err := netip.ParseAddr(p.IP()); err == nil {
		pr.Statistics.IP = ip
		pr.Statistics.DestIsIP = true
	}

	for _, opt := range opts {
		opt(&pr)
	}
	return &pr
}

const (
	DefaultInterval = 1 * time.Second
	DefaultTimeout  = 5 * time.Second
)

func (p *Prober) Probe(ctx context.Context) (statistics.Statistics, error) {
	p.Ticker = time.NewTicker(p.Interval)
	defer p.Ticker.Stop()

	timeoutTimer := time.NewTimer(p.Timeout)
	defer timeoutTimer.Stop()

	p.Statistics.StartTime = time.Now()

	for {
		select {

		case <-ctx.Done():
			p.Statistics.EndTime = time.Now()
			p.Statistics.UpTime = p.Statistics.EndTime.Sub(p.Statistics.StartTime)

			// Finalize uptime/downtime tracking
			if p.Statistics.DestWasDown {
				downDuration := p.Statistics.EndTime.Sub(p.Statistics.StartOfDowntime)
				p.Statistics.TotalDowntime += downDuration
				statistics.SetLongestDuration(p.Statistics.StartOfDowntime, downDuration, &p.Statistics.LongestDown)
			} else if !p.Statistics.StartOfUptime.IsZero() {
				upDuration := p.Statistics.EndTime.Sub(p.Statistics.StartOfUptime)
				p.Statistics.TotalUptime += upDuration
				statistics.SetLongestDuration(p.Statistics.StartOfUptime, upDuration, &p.Statistics.LongestUp)
			}

			return p.Statistics, nil

		case <-timeoutTimer.C:
			p.Statistics.EndTime = time.Now()
			p.Statistics.UpTime = p.Statistics.EndTime.Sub(p.Statistics.StartTime)

			// Finalize uptime/downtime tracking
			if p.Statistics.DestWasDown {
				downDuration := p.Statistics.EndTime.Sub(p.Statistics.StartOfDowntime)
				p.Statistics.TotalDowntime += downDuration
				statistics.SetLongestDuration(p.Statistics.StartOfDowntime, downDuration, &p.Statistics.LongestDown)
			} else if !p.Statistics.StartOfUptime.IsZero() {
				upDuration := p.Statistics.EndTime.Sub(p.Statistics.StartOfUptime)
				p.Statistics.TotalUptime += upDuration
				statistics.SetLongestDuration(p.Statistics.StartOfUptime, upDuration, &p.Statistics.LongestUp)
			}

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
				continue
			}

			// Handle success
			rttMs := statistics.NanoToMillisecond(rtt.Nanoseconds())
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
				statistics.SetLongestDuration(p.Statistics.StartOfDowntime, downDuration, &p.Statistics.LongestDowntime)
				p.Statistics.StartOfUptime = pingTime
			}

			if p.Statistics.StartOfUptime.IsZero() {
				p.Statistics.StartOfUptime = pingTime
			}

			p.printer.PrintProbeSuccess(&p.Statistics)
		}
	}
}
