// Package probers handles the general probing logic
package probers

import (
	"context"
	"errors"
	"time"

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
	Ticker     *time.Ticker
	Timeout    time.Duration
	Interval   time.Duration
	Statistics stats.Statistics
}

type Pinger interface {
	Ping(ctx context.Context) error
	IP() string
	Port() uint16
}

type ProberOption func(*Prober)

func WithInterval(interval time.Duration) ProberOption {
	return func(p *Prober) {
		p.Interval = interval
	}
}

func WithTimeout(timeout time.Duration) ProberOption {
	return func(p *Prober) {
		p.Timeout = timeout + p.Interval
	}
}

func WithPrinter(printer printers.Printer) ProberOption {
	return func(p *Prober) {
		p.printer = printer
	}
}

func NewProber(p Pinger, opts ...ProberOption) *Prober {
	pr := Prober{
		pinger:   p,
		printer:  printers.NewColorPrinter(),
		Interval: DefaultInterval,
		Timeout:  DefaultTimeout,
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

func (p *Prober) Probe(ctx context.Context) (stats.Statistics, error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, p.Timeout)
	defer cancel()

	p.Ticker = time.NewTicker(p.Interval)
	defer p.Ticker.Stop()

	p.Statistics.StartTime = time.Now()

	for {
		select {

		case <-ctx.Done():
			p.Statistics.EndTime = time.Now()
			p.Statistics.UpTime = p.Statistics.EndTime.Sub(p.Statistics.StartTime)
			return p.Statistics, nil

		case <-p.Ticker.C:
			pingTime := time.Now()
			err := p.pinger.Ping(ctx)
			rtt := time.Since(pingTime)
			if err != nil {
				if errors.Is(err, ErrPingCompleted) {
					return p.Statistics, nil
				}
				p.printer.PrintProbeFailure(&p.Statistics)
				p.Statistics.Failed++
				p.Statistics.LongestDown = stats.NewLongestTime(pingTime, rtt)
				continue
			}

			p.Statistics.RTT = append(p.Statistics.RTT, utils.NanoToMillisecond(rtt.Nanoseconds()))
			p.Statistics.HasResults = true
			p.Statistics.Successful++
			p.Statistics.LongestUp = stats.NewLongestTime(pingTime, rtt)
			p.printer.PrintProbeSuccess(&p.Statistics)

		case <-time.After(p.Timeout):
			p.Statistics.EndTime = time.Now()
			p.Statistics.UpTime = p.Statistics.EndTime.Sub(p.Statistics.StartTime)
			return p.Statistics, ErrTimeout
		}
	}
}
