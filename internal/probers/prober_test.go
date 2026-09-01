package probers

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/dns"
	"github.com/pouriyajamshidi/tcping/v3/internal/nic"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
)

// fakePinger returns a scripted result for each call to Ping, driven by
// outcomeFn. Calls are counted so tests can assert how many probes ran.
type fakePinger struct {
	mu        sync.Mutex
	callCount int
	outcomeFn func(call int) (ProbeResult, error)
}

func (f *fakePinger) Ping(ctx context.Context) (ProbeResult, error) {
	f.mu.Lock()
	call := f.callCount
	f.callCount++
	f.mu.Unlock()

	return f.outcomeFn(call)
}

func (f *fakePinger) calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.callCount
}

// alwaysSucceeds is a convenience fakePinger that succeeds on every call.
func alwaysSucceeds() *fakePinger {
	return &fakePinger{outcomeFn: func(int) (ProbeResult, error) {
		return ProbeResult{LocalAddr: &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 12345}}, nil
	}}
}

// alwaysFails is a convenience fakePinger that fails on every call.
func alwaysFails() *fakePinger {
	return &fakePinger{outcomeFn: func(int) (ProbeResult, error) {
		return ProbeResult{}, errConnRefused
	}}
}

var errConnRefused = &net.OpError{Op: "dial", Err: errFake("connection refused")}

type errFake string

func (e errFake) Error() string { return string(e) }

// fakePrinter implements printers.Printer and records how many times each
// method was called, so tests can assert on the printer's involvement
// without depending on any concrete printer implementation.
type fakePrinter struct {
	mu sync.Mutex

	startCalls      int
	successCalls    int
	failureCalls    int
	statsCalls      int
	retryCalls      int
	downtimeCalls   int
	errorCalls      int
	shutdownCalls   int
	lastRetryTarget string
}

func (f *fakePrinter) PrintStart(s *stats.Statistics) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.startCalls++
}

func (f *fakePrinter) PrintProbeSuccess(s *stats.Statistics) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.successCalls++
}

func (f *fakePrinter) PrintProbeFailure(s *stats.Statistics) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failureCalls++
}

func (f *fakePrinter) PrintStatistics(s *stats.Statistics) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.statsCalls++
}

func (f *fakePrinter) PrintRetryingToResolve(hostname string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.retryCalls++
	f.lastRetryTarget = hostname
}

func (f *fakePrinter) PrintDownTimeDuration(s *stats.Statistics) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.downtimeCalls++
}

func (f *fakePrinter) PrintError(format string, args ...any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.errorCalls++
}

func (f *fakePrinter) Shutdown(s *stats.Statistics) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.shutdownCalls++
}

func (f *fakePrinter) snapshot() fakePrinter {
	f.mu.Lock()
	defer f.mu.Unlock()
	return fakePrinter{
		startCalls:    f.startCalls,
		successCalls:  f.successCalls,
		failureCalls:  f.failureCalls,
		statsCalls:    f.statsCalls,
		retryCalls:    f.retryCalls,
		downtimeCalls: f.downtimeCalls,
		errorCalls:    f.errorCalls,
		shutdownCalls: f.shutdownCalls,
	}
}

// newTestProber builds a Prober wired up with the given pinger and a fresh
// fakePrinter/Statistics pair, ready for either the unit-level handle*
// tests or a full Probe/ProbeV2 run.
func newTestProber(pinger Pinger, cfg config.Config) (*Prober, *fakePrinter) {
	printer := &fakePrinter{}
	p := &Prober{
		pinger:     pinger,
		printer:    printer,
		config:     cfg,
		Statistics: &stats.Statistics{},
	}
	return p, printer
}

// --- handleProbeFailure -------------------------------------------------

func TestHandleProbeFailure_FirstFailureRecordsCounters(t *testing.T) {
	p, _ := newTestProber(nil, config.Config{})
	now := time.Now()

	p.handleProbeFailure(now)

	s := p.Statistics
	if s.Failed != 1 || s.TotalUnsuccessfulProbes != 1 || s.OngoingUnsuccessfulProbes != 1 {
		t.Fatalf("got Failed=%d TotalUnsuccessfulProbes=%d OngoingUnsuccessfulProbes=%d, want all 1",
			s.Failed, s.TotalUnsuccessfulProbes, s.OngoingUnsuccessfulProbes)
	}
	if !s.LastProbeHadFailed {
		t.Error("LastProbeHadFailed = false, want true")
	}
	if !s.StartOfDowntime.Equal(now) {
		t.Errorf("StartOfDowntime = %v, want %v", s.StartOfDowntime, now)
	}
	if !s.LastUnsuccessfulProbe.Equal(now) {
		t.Errorf("LastUnsuccessfulProbe = %v, want %v", s.LastUnsuccessfulProbe, now)
	}
}

func TestHandleProbeFailure_EndsOngoingUptimeStreak(t *testing.T) {
	p, _ := newTestProber(nil, config.Config{})
	start := time.Now()
	p.Statistics.StartOfUptime = start
	p.Statistics.OngoingSuccessfulProbes = 5

	failAt := start.Add(100 * time.Millisecond)
	p.handleProbeFailure(failAt)

	s := p.Statistics
	if s.OngoingSuccessfulProbes != 0 {
		t.Errorf("OngoingSuccessfulProbes = %d, want 0", s.OngoingSuccessfulProbes)
	}
	if s.TotalUptime != 100*time.Millisecond {
		t.Errorf("TotalUptime = %v, want 100ms", s.TotalUptime)
	}
	if s.LongestUptime.Duration != 100*time.Millisecond || !s.LongestUptime.Start.Equal(start) {
		t.Errorf("LongestUptime = %+v, want Duration=100ms Start=%v", s.LongestUptime, start)
	}
}

func TestHandleProbeFailure_ConsecutiveFailuresDoNotDoubleCountDowntime(t *testing.T) {
	p, _ := newTestProber(nil, config.Config{})
	first := time.Now()
	second := first.Add(50 * time.Millisecond)

	p.handleProbeFailure(first)
	p.handleProbeFailure(second)

	s := p.Statistics
	if s.OngoingUnsuccessfulProbes != 2 {
		t.Errorf("OngoingUnsuccessfulProbes = %d, want 2", s.OngoingUnsuccessfulProbes)
	}
	// StartOfDowntime should stay pinned to the first failure in the streak.
	if !s.StartOfDowntime.Equal(first) {
		t.Errorf("StartOfDowntime = %v, want %v (unchanged after 2nd failure)", s.StartOfDowntime, first)
	}
}

func TestHandleProbeFailure_UsesConfiguredInterfaceAddress(t *testing.T) {
	sourceAddr := &net.TCPAddr{IP: net.IPv4(10, 0, 0, 5), Port: 0}
	cfg := config.Config{
		NetworkInterface: nic.NetworkInterface{
			Use:    true,
			Dialer: net.Dialer{LocalAddr: sourceAddr},
		},
	}
	p, _ := newTestProber(nil, cfg)

	p.handleProbeFailure(time.Now())

	if p.Statistics.LocalAddr != sourceAddr {
		t.Errorf("LocalAddr = %v, want %v", p.Statistics.LocalAddr, sourceAddr)
	}
}

// --- handleProbeSuccess --------------------------------------------------

func TestHandleProbeSuccess_RecordsRTTAndCounters(t *testing.T) {
	p, _ := newTestProber(nil, config.Config{})
	now := time.Now()
	localAddr := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 4321}

	p.handleProbeSuccess(now, 15*time.Millisecond, ProbeResult{LocalAddr: localAddr})

	s := p.Statistics
	if len(s.RTT) != 1 || s.RTT[0] != 15 {
		t.Errorf("RTT = %v, want [15]", s.RTT)
	}
	if s.LatestRTT != 15 {
		t.Errorf("LatestRTT = %v, want 15", s.LatestRTT)
	}
	if !s.RTTResults.HasResults {
		t.Error("HasResults = false, want true")
	}
	if s.LocalAddr != localAddr {
		t.Errorf("LocalAddr = %v, want %v", s.LocalAddr, localAddr)
	}
	if s.Successful != 1 || s.TotalSuccessfulProbes != 1 || s.OngoingSuccessfulProbes != 1 {
		t.Errorf("got Successful=%d TotalSuccessfulProbes=%d OngoingSuccessfulProbes=%d, want all 1",
			s.Successful, s.TotalSuccessfulProbes, s.OngoingSuccessfulProbes)
	}
	if !s.StartOfUptime.Equal(now) {
		t.Errorf("StartOfUptime = %v, want %v", s.StartOfUptime, now)
	}
}

func TestHandleProbeSuccess_EndsOngoingDowntimeStreak(t *testing.T) {
	p, printer := newTestProber(nil, config.Config{})
	start := time.Now()
	p.Statistics.LastProbeHadFailed = true
	p.Statistics.StartOfDowntime = start
	p.Statistics.OngoingUnsuccessfulProbes = 3

	upAt := start.Add(50 * time.Millisecond)
	p.handleProbeSuccess(upAt, time.Millisecond, ProbeResult{})

	s := p.Statistics
	if s.LastProbeHadFailed {
		t.Error("LastProbeHadFailed = true, want false")
	}
	if s.TotalDowntime != 50*time.Millisecond || s.DownTime != 50*time.Millisecond {
		t.Errorf("TotalDowntime=%v DownTime=%v, want both 50ms", s.TotalDowntime, s.DownTime)
	}
	if s.LongestDown.Duration != 50*time.Millisecond {
		t.Errorf("LongestDown.Duration = %v, want 50ms", s.LongestDown.Duration)
	}
	if !s.StartOfUptime.Equal(upAt) {
		t.Errorf("StartOfUptime = %v, want %v", s.StartOfUptime, upAt)
	}
	if got := printer.snapshot().downtimeCalls; got != 1 {
		t.Errorf("PrintDownTimeDuration called %d times, want 1", got)
	}
}

func TestHandleProbeSuccess_OngoingUptimeDoesNotReprintDowntime(t *testing.T) {
	p, printer := newTestProber(nil, config.Config{})
	start := time.Now()
	p.Statistics.StartOfUptime = start

	p.handleProbeSuccess(start.Add(time.Millisecond), time.Millisecond, ProbeResult{})
	p.handleProbeSuccess(start.Add(2*time.Millisecond), time.Millisecond, ProbeResult{})

	if got := printer.snapshot().downtimeCalls; got != 0 {
		t.Errorf("PrintDownTimeDuration called %d times, want 0 (never went down)", got)
	}
	if p.Statistics.OngoingSuccessfulProbes != 2 {
		t.Errorf("OngoingSuccessfulProbes = %d, want 2", p.Statistics.OngoingSuccessfulProbes)
	}
}

// --- finalizeStatistics ---------------------------------------------------

func TestFinalizeStatistics_WhileDownAccruesRemainingDowntime(t *testing.T) {
	p, _ := newTestProber(nil, config.Config{})
	p.Statistics.LastProbeHadFailed = true
	p.Statistics.StartOfDowntime = time.Now().Add(-50 * time.Millisecond)

	p.finalizeStatistics()

	s := p.Statistics
	if s.EndTime.IsZero() {
		t.Fatal("EndTime was not set")
	}
	if s.TotalDowntime < 50*time.Millisecond {
		t.Errorf("TotalDowntime = %v, want at least 50ms", s.TotalDowntime)
	}
	if s.LongestDown.Duration != s.TotalDowntime {
		t.Errorf("LongestDown.Duration = %v, want %v", s.LongestDown.Duration, s.TotalDowntime)
	}
}

func TestFinalizeStatistics_WhileUpAccruesRemainingUptime(t *testing.T) {
	p, _ := newTestProber(nil, config.Config{})
	p.Statistics.StartOfUptime = time.Now().Add(-50 * time.Millisecond)

	p.finalizeStatistics()

	s := p.Statistics
	if s.TotalUptime < 50*time.Millisecond {
		t.Errorf("TotalUptime = %v, want at least 50ms", s.TotalUptime)
	}
	if s.LongestUp.Duration != s.TotalUptime {
		t.Errorf("LongestUp.Duration = %v, want %v", s.LongestUp.Duration, s.TotalUptime)
	}
}

func TestFinalizeStatistics_NoProbesYetIsANoop(t *testing.T) {
	p, _ := newTestProber(nil, config.Config{})
	p.Statistics.StartTime = time.Now()

	p.finalizeStatistics()

	s := p.Statistics
	if s.TotalUptime != 0 || s.TotalDowntime != 0 {
		t.Errorf("TotalUptime=%v TotalDowntime=%v, want both 0 when no probe ever ran", s.TotalUptime, s.TotalDowntime)
	}
	if s.UpTime < 0 {
		t.Errorf("UpTime = %v, want >= 0", s.UpTime)
	}
}

// --- Probe (v1) -----------------------------------------------------------

func TestProbe_StopsAfterProbesBeforeQuit(t *testing.T) {
	pinger := alwaysSucceeds()
	cfg := config.Config{
		IntervalBetweenProbes: 5 * time.Millisecond,
		ProbesBeforeQuit:      3,
	}
	p, printer := newTestProber(pinger, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := p.Probe(ctx); err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	if got := pinger.calls(); got != 3 {
		t.Errorf("Ping called %d times, want 3", got)
	}
	if p.Statistics.TotalSuccessfulProbes != 3 {
		t.Errorf("TotalSuccessfulProbes = %d, want 3", p.Statistics.TotalSuccessfulProbes)
	}

	snap := printer.snapshot()
	if snap.startCalls != 1 {
		t.Errorf("PrintStart called %d times, want 1", snap.startCalls)
	}
	if snap.successCalls != 3 {
		t.Errorf("PrintProbeSuccess called %d times, want 3", snap.successCalls)
	}
}

func TestProbe_StopsOnContextCancellation(t *testing.T) {
	pinger := alwaysSucceeds()
	cfg := config.Config{
		IntervalBetweenProbes: 5 * time.Millisecond,
		// ProbesBeforeQuit left at 0: run until ctx is done.
	}
	p, _ := newTestProber(pinger, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		p.Probe(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Probe() did not return after its context was cancelled")
	}

	if pinger.calls() == 0 {
		t.Error("Ping was never called before the context expired")
	}
	if !p.Statistics.EndTime.After(p.Statistics.StartTime) {
		t.Error("finalizeStatistics does not appear to have run: EndTime is not after StartTime")
	}
}

func TestProbe_TracksDowntimeThenRecovery(t *testing.T) {
	pinger := &fakePinger{outcomeFn: func(call int) (ProbeResult, error) {
		if call < 2 {
			return ProbeResult{}, errConnRefused
		}
		return ProbeResult{}, nil
	}}
	cfg := config.Config{
		IntervalBetweenProbes: 5 * time.Millisecond,
		ProbesBeforeQuit:      3,
	}
	p, printer := newTestProber(pinger, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := p.Probe(ctx); err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	s := p.Statistics
	if s.Failed != 2 || s.Successful != 1 {
		t.Errorf("Failed=%d Successful=%d, want Failed=2 Successful=1", s.Failed, s.Successful)
	}
	if s.LastProbeHadFailed {
		t.Error("LastProbeHadFailed = true, want false after recovering")
	}
	if s.TotalDowntime <= 0 {
		t.Error("TotalDowntime should be greater than 0 after two failed probes")
	}

	snap := printer.snapshot()
	if snap.failureCalls != 2 {
		t.Errorf("PrintProbeFailure called %d times, want 2", snap.failureCalls)
	}
	if snap.downtimeCalls != 1 {
		t.Errorf("PrintDownTimeDuration called %d times, want 1", snap.downtimeCalls)
	}
}

func TestProbe_RetriesHostnameResolutionAfterNFailures(t *testing.T) {
	pinger := alwaysFails()
	cfg := config.Config{
		IntervalBetweenProbes:      5 * time.Millisecond,
		ProbesBeforeQuit:           2,
		ShouldRetryResolve:         true,
		RetryResolveAfterNFailures: 2,
		// A literal IP resolves without touching the network.
		Resolver: dns.NewResolver("", time.Second, false, false),
	}
	p, printer := newTestProber(pinger, cfg)
	p.Statistics.Hostname = "127.0.0.1"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := p.Probe(ctx); err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	if p.Statistics.RetriedHostnameLookups != 1 {
		t.Errorf("RetriedHostnameLookups = %d, want 1", p.Statistics.RetriedHostnameLookups)
	}

	snap := printer.snapshot()
	if snap.retryCalls != 1 {
		t.Errorf("PrintRetryingToResolve called %d times, want 1", snap.retryCalls)
	}
	if snap.errorCalls != 0 {
		t.Errorf("PrintError called %d times, want 0 (resolving a literal IP should not fail)", snap.errorCalls)
	}
	if printer.lastRetryTarget != "127.0.0.1" {
		t.Errorf("PrintRetryingToResolve hostname = %q, want %q", printer.lastRetryTarget, "127.0.0.1")
	}
}

// --- ProbeV2 ----------------------------------------------------------------

func TestProbeV2_ProbesImmediatelyWithoutWaitingForFirstTick(t *testing.T) {
	pinger := alwaysSucceeds()
	cfg := config.Config{
		// Deliberately much longer than the context timeout below, so a
		// probe can only have happened via the initial, un-ticked call.
		IntervalBetweenProbes: 500 * time.Millisecond,
	}
	p, _ := newTestProber(pinger, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if _, err := p.ProbeV2(ctx); err != nil {
		t.Fatalf("ProbeV2() error = %v", err)
	}

	if got := pinger.calls(); got != 1 {
		t.Errorf("Ping called %d times, want exactly 1 (the immediate, pre-tick probe)", got)
	}
}

// ProbeV2's ProbesBeforeQuit check only runs inside the ticker branch, and
// the immediate initial probe doesn't count toward it. For a threshold of 1
// this means two probes actually run before ProbeV2 returns: one right away,
// and one more on the first tick because the check can't fire any earlier
// than that. This test documents that current, slightly surprising behavior.
func TestProbeV2_ProbesBeforeQuitOfOneRunsTwoProbes(t *testing.T) {
	pinger := alwaysSucceeds()
	cfg := config.Config{
		IntervalBetweenProbes: 5 * time.Millisecond,
		ProbesBeforeQuit:      1,
	}
	p, _ := newTestProber(pinger, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := p.ProbeV2(ctx); err != nil {
		t.Fatalf("ProbeV2() error = %v", err)
	}

	if got := pinger.calls(); got != 2 {
		t.Errorf("Ping called %d times, want 2 (see comment above)", got)
	}
}

func TestProbeV2_ProbesBeforeQuitOfThreeRunsThreeProbes(t *testing.T) {
	pinger := alwaysSucceeds()
	cfg := config.Config{
		IntervalBetweenProbes: 5 * time.Millisecond,
		ProbesBeforeQuit:      3,
	}
	p, _ := newTestProber(pinger, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := p.ProbeV2(ctx); err != nil {
		t.Fatalf("ProbeV2() error = %v", err)
	}

	if got := pinger.calls(); got != 3 {
		t.Errorf("Ping called %d times, want 3", got)
	}
	if p.Statistics.TotalSuccessfulProbes != 3 {
		t.Errorf("TotalSuccessfulProbes = %d, want 3", p.Statistics.TotalSuccessfulProbes)
	}
}

func TestProbeV2_StopsOnContextCancellation(t *testing.T) {
	pinger := alwaysSucceeds()
	cfg := config.Config{IntervalBetweenProbes: 5 * time.Millisecond}
	p, _ := newTestProber(pinger, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		p.ProbeV2(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ProbeV2() did not return after its context was cancelled")
	}

	if pinger.calls() == 0 {
		t.Error("Ping was never called before the context expired")
	}
}
