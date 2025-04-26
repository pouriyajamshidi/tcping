package probes

import (
	"context"
	"errors"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/types"
)

var (
	ErrTimeout = errors.New("timed out waiting for ping")
)

type Prober struct {
	pinger     Pinger
	Ticker     *time.Ticker
	Timeout    time.Duration
	Interval   time.Duration
	Done       chan struct{}
	Statistics Statistics
}

type Printer interface {
	Print(statistics Statistics)
}

type Pinger interface {
	Ping(ctx context.Context) error
}

type Statistics struct {
	StartTime   time.Time
	EndTime     time.Time
	UpTime      time.Duration
	DownTime    time.Duration
	Successful  int
	Failed      int
	LongestUp   types.LongestTime
	LongestDown types.LongestTime
	RTT         []time.Duration
	HostChanges []types.HostnameChange
	HasResults  bool
}

func NewProber(p Pinger) *Prober {
	return &Prober{
		pinger: p,
	}
}

func (p *Prober) Probe(ctx context.Context) (Statistics, error) {
	if p.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.Timeout)
		defer cancel()
	}
	p.Ticker = time.NewTicker(p.Interval)
	defer p.Ticker.Stop()
	p.Done = make(chan struct{})
	defer close(p.Done)
	p.Statistics.StartTime = time.Now()
	defer func() {
		p.Statistics.EndTime = time.Now()
		p.Statistics.UpTime = p.Statistics.EndTime.Sub(p.Statistics.StartTime)
	}()
	pinging := true
	for pinging {
		select {
		case <-ctx.Done():
			return p.Statistics, ctx.Err()
		case <-p.Ticker.C:
			pingTime := time.Now()
			err := p.pinger.Ping(ctx)
			rtt := time.Since(pingTime)
			p.Statistics.RTT = append(p.Statistics.RTT, rtt)
			p.Statistics.HasResults = true
			if err != nil {
				p.Statistics.Failed++
				p.Statistics.LongestDown = types.NewLongestTime(pingTime, rtt)
			}
			p.Statistics.Successful++
			p.Statistics.LongestUp = types.NewLongestTime(pingTime, rtt)
		case <-time.After(p.Timeout):
			return p.Statistics, ErrTimeout
		case <-p.Done:
			pinging = false
		}
	}
	return p.Statistics, nil
}
