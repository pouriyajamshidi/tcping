package probes

import (
	"context"
	"errors"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/probes/printer"
	"github.com/pouriyajamshidi/tcping/v2/probes/statistics"
	"github.com/pouriyajamshidi/tcping/v2/types"
)

var (
	ErrTimeout = errors.New("timed out waiting for ping")
)

type Prober struct {
	pinger     Pinger
	printer    Printer
	Ticker     *time.Ticker
	Timeout    time.Duration
	Interval   time.Duration
	Statistics statistics.Statistics
}

type Printer interface {
	Print(s string)
	PrintStatistics(statistics statistics.Statistics)
	PrintError(err error)
}

type Pinger interface {
	Ping(ctx context.Context) error
}

type ProberOption func(*Prober)

func WithInterval(interval time.Duration) ProberOption {
	return func(p *Prober) {
		p.Interval = interval
	}
}

func WithTimeout(timeout time.Duration) ProberOption {
	return func(p *Prober) {
		p.Timeout = timeout
	}
}

func WithPrinter(printer Printer) ProberOption {
	return func(p *Prober) {
		p.printer = printer
	}
}

func NewProber(p Pinger, opts ...ProberOption) *Prober {
	pr := Prober{
		pinger:   p,
		printer:  printer.NewColor(),
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

func (p *Prober) Probe(ctx context.Context) (statistics.Statistics, error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, p.Timeout)
	defer cancel()

	p.Ticker = time.NewTicker(p.Interval)
	defer p.Ticker.Stop()
	p.Statistics.StartTime = time.Now()
	defer func() {
		p.Statistics.EndTime = time.Now()
		p.Statistics.UpTime = p.Statistics.EndTime.Sub(p.Statistics.StartTime)
	}()
	for {
		select {
		case <-ctx.Done():
			return p.Statistics, nil
		case <-p.Ticker.C:
			pingTime := time.Now()
			err := p.pinger.Ping(ctx)
			rtt := time.Since(pingTime)
			if err != nil {
				p.Statistics.Failed++
				p.Statistics.LongestDown = types.NewLongestTime(pingTime, rtt)
				p.printer.PrintError(err)
				p.printer.PrintStatistics(p.Statistics)
				continue
			}
			p.Statistics.RTT = append(p.Statistics.RTT, rtt)
			p.Statistics.HasResults = true
			p.Statistics.Successful++
			p.Statistics.LongestUp = types.NewLongestTime(pingTime, rtt)
			p.printer.PrintStatistics(p.Statistics)
		case <-time.After(p.Timeout):
			return p.Statistics, ErrTimeout
		}
	}
}
