package probe

import (
	"context"
	"net"
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/config"
	"github.com/pouriyajamshidi/tcping/v3/dns"
	"github.com/pouriyajamshidi/tcping/v3/nic"
	"github.com/pouriyajamshidi/tcping/v3/stats"
)

// fakePinger returns a scripted result for each call to Ping, driven by
// outcomeFn. Calls are counted, and every IP it was called with is
// recorded, so tests can assert how many probes ran and what they targeted.
type fakePinger struct {
	mu        sync.Mutex
	callCount int
	ipsCalled []netip.Addr
	outcomeFn func(call int) (ProbeResult, error)
}

func (f *fakePinger) Ping(ctx context.Context, ip netip.Addr) (ProbeResult, error) {
	f.mu.Lock()
	call := f.callCount
	f.callCount++
	f.ipsCalled = append(f.ipsCalled, ip)
	f.mu.Unlock()

	return f.outcomeFn(call)
}

func (f *fakePinger) calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.callCount
}

func (f *fakePinger) ips() []netip.Addr {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]netip.Addr(nil), f.ipsCalled...)
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

	startCalls          int
	nameResolutionCalls int
	successCalls        int
	failureCalls        int
	statsCalls          int
	retryCalls          int
	downtimeCalls       int
	uptimeCalls         int
	errorCalls          int
	shutdownCalls       int
	lastRetryTarget     string
}

func (f *fakePrinter) PrintStart(s *stats.Statistics) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.startCalls++
}

func (f *fakePrinter) PrintNameResolutionDuration(s *stats.Statistics) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nameResolutionCalls++
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

func (f *fakePrinter) PrintUpTimeDuration(s *stats.Statistics) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.uptimeCalls++
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
		startCalls:          f.startCalls,
		nameResolutionCalls: f.nameResolutionCalls,
		successCalls:        f.successCalls,
		failureCalls:        f.failureCalls,
		statsCalls:          f.statsCalls,
		retryCalls:          f.retryCalls,
		downtimeCalls:       f.downtimeCalls,
		uptimeCalls:         f.uptimeCalls,
		errorCalls:          f.errorCalls,
		shutdownCalls:       f.shutdownCalls,
	}
}

// newTestProber builds a Prober wired up with the given pinger and a fresh
// fakePrinter/Statistics pair, ready for either the unit-level handle*
// tests or a full Probe run.
func newTestProber(pinger Pinger, cfg config.Config) (*Prober, *fakePrinter) {
	printer := &fakePrinter{}
	p := &Prober{
		pinger:     pinger,
		printer:    printer,
		config:     cfg,
		statistics: &stats.Statistics{},
	}
	return p, printer
}

// --- handleProbeFailure -------------------------------------------------

func TestHandleProbeFailure_FirstFailureRecordsCounters(t *testing.T) {
	p, _ := newTestProber(nil, config.Config{})
	now := time.Now()

	p.handleProbeFailure(now, ProbeResult{})

	s := p.statistics
	if s.TotalUnsuccessfulProbes != 1 || s.OngoingUnsuccessfulProbes != 1 {
		t.Fatalf("got TotalUnsuccessfulProbes=%d OngoingUnsuccessfulProbes=%d, want all 1",
			s.TotalUnsuccessfulProbes, s.OngoingUnsuccessfulProbes)
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
	p, printer := newTestProber(nil, config.Config{})
	start := time.Now()
	p.statistics.StartOfUptime = start
	p.statistics.OngoingSuccessfulProbes = 5

	failAt := start.Add(100 * time.Millisecond)
	p.handleProbeFailure(failAt, ProbeResult{})

	s := p.statistics
	if s.OngoingSuccessfulProbes != 0 {
		t.Errorf("OngoingSuccessfulProbes = %d, want 0", s.OngoingSuccessfulProbes)
	}
	if s.TotalUptime != 100*time.Millisecond || s.CurrentUptime != 100*time.Millisecond {
		t.Errorf("TotalUptime=%v CurrentUptime=%v, want both 100ms", s.TotalUptime, s.CurrentUptime)
	}
	if s.LongestUptime.Duration != 100*time.Millisecond || !s.LongestUptime.Start.Equal(start) {
		t.Errorf("LongestUptime = %+v, want Duration=100ms Start=%v", s.LongestUptime, start)
	}
	if got := printer.snapshot().uptimeCalls; got != 1 {
		t.Errorf("PrintUpTimeDuration called %d times, want 1", got)
	}
}

// The very first probe ever has no prior uptime streak to report (the
// target's status before this point was simply unknown), so failing on it
// must not print a bogus "up for 0s" (or worse, garbage) message.
func TestHandleProbeFailure_FirstEverFailureDoesNotPrintUptime(t *testing.T) {
	p, printer := newTestProber(nil, config.Config{})

	p.handleProbeFailure(time.Now(), ProbeResult{})

	if got := printer.snapshot().uptimeCalls; got != 0 {
		t.Errorf("PrintUpTimeDuration called %d times, want 0 (no uptime streak ever started)", got)
	}
}

// Consecutive failures are the same, single downtime streak - the "was up
// for X" message should only ever fire once, at the moment uptime ends,
// not be repeated on every subsequent failed probe.
func TestHandleProbeFailure_ConsecutiveFailuresDoNotReprintUptime(t *testing.T) {
	p, printer := newTestProber(nil, config.Config{})
	start := time.Now()
	p.statistics.StartOfUptime = start

	p.handleProbeFailure(start.Add(50*time.Millisecond), ProbeResult{})
	p.handleProbeFailure(start.Add(100*time.Millisecond), ProbeResult{})

	if got := printer.snapshot().uptimeCalls; got != 1 {
		t.Errorf("PrintUpTimeDuration called %d times, want 1", got)
	}
}

func TestHandleProbeFailure_ConsecutiveFailuresDoNotDoubleCountDowntime(t *testing.T) {
	p, _ := newTestProber(nil, config.Config{})
	first := time.Now()
	second := first.Add(50 * time.Millisecond)

	p.handleProbeFailure(first, ProbeResult{})
	p.handleProbeFailure(second, ProbeResult{})

	s := p.statistics
	if s.OngoingUnsuccessfulProbes != 2 {
		t.Errorf("OngoingUnsuccessfulProbes = %d, want 2", s.OngoingUnsuccessfulProbes)
	}
	// StartOfDowntime should stay pinned to the first failure in the streak.
	if !s.StartOfDowntime.Equal(first) {
		t.Errorf("StartOfDowntime = %v, want %v (unchanged after 2nd failure)", s.StartOfDowntime, first)
	}
}

func TestHandleProbeFailure_UsesConfiguredInterfaceAddress(t *testing.T) {
	sourceIP := net.IPv4(10, 0, 0, 5)
	cfg := config.Config{
		NetworkInterface: nic.NetworkInterface{
			Use:        true,
			SourceIPv4: sourceIP,
		},
	}
	p, _ := newTestProber(nil, cfg)
	p.statistics.IP = netip.MustParseAddr("93.184.216.34") // an IPv4 target

	p.handleProbeFailure(time.Now(), ProbeResult{})

	wantAddr := &net.TCPAddr{IP: sourceIP}
	gotAddr, ok := p.statistics.LocalAddr.(*net.TCPAddr)
	if !ok || !gotAddr.IP.Equal(wantAddr.IP) {
		t.Errorf("LocalAddr = %v, want %v", p.statistics.LocalAddr, wantAddr)
	}
}

// --- handleProbeSuccess --------------------------------------------------

func TestHandleProbeSuccess_RecordsRTTAndCounters(t *testing.T) {
	p, _ := newTestProber(nil, config.Config{})
	now := time.Now()
	localAddr := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 4321}

	p.handleProbeSuccess(now, 15*time.Millisecond, ProbeResult{LocalAddr: localAddr})

	s := p.statistics
	if s.LatestRTT != 15 {
		t.Errorf("LatestRTT = %v, want 15", s.LatestRTT)
	}
	if s.TotalSuccessfulProbes == 0 {
		t.Error("TotalSuccessfulProbes = 0, want at least 1")
	}
	if s.RTTResults.Min != 15 || s.RTTResults.Max != 15 || s.RTTResults.Average != 15 {
		t.Errorf("RTTResults = %+v, want Min=Max=Average=15", s.RTTResults)
	}
	if s.LocalAddr != localAddr {
		t.Errorf("LocalAddr = %v, want %v", s.LocalAddr, localAddr)
	}
	if s.TotalSuccessfulProbes != 1 || s.OngoingSuccessfulProbes != 1 {
		t.Errorf("got TotalSuccessfulProbes=%d OngoingSuccessfulProbes=%d, want all 1",
			s.TotalSuccessfulProbes, s.OngoingSuccessfulProbes)
	}
	if !s.StartOfUptime.Equal(now) {
		t.Errorf("StartOfUptime = %v, want %v", s.StartOfUptime, now)
	}
}

func TestHandleProbeSuccess_EndsOngoingDowntimeStreak(t *testing.T) {
	p, printer := newTestProber(nil, config.Config{})
	start := time.Now()
	p.statistics.LastProbeHadFailed = true
	p.statistics.StartOfDowntime = start
	p.statistics.OngoingUnsuccessfulProbes = 3

	upAt := start.Add(50 * time.Millisecond)
	p.handleProbeSuccess(upAt, time.Millisecond, ProbeResult{})

	s := p.statistics
	if s.LastProbeHadFailed {
		t.Error("LastProbeHadFailed = true, want false")
	}
	if s.TotalDowntime != 50*time.Millisecond || s.CurrentDowntime != 50*time.Millisecond {
		t.Errorf("TotalDowntime=%v DownTime=%v, want both 50ms", s.TotalDowntime, s.CurrentDowntime)
	}
	if s.LongestDowntime.Duration != 50*time.Millisecond {
		t.Errorf("LongestDowntime.Duration = %v, want 50ms", s.LongestDowntime.Duration)
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
	p.statistics.StartOfUptime = start

	p.handleProbeSuccess(start.Add(time.Millisecond), time.Millisecond, ProbeResult{})
	p.handleProbeSuccess(start.Add(2*time.Millisecond), time.Millisecond, ProbeResult{})

	if got := printer.snapshot().downtimeCalls; got != 0 {
		t.Errorf("PrintDownTimeDuration called %d times, want 0 (never went down)", got)
	}
	if p.statistics.OngoingSuccessfulProbes != 2 {
		t.Errorf("OngoingSuccessfulProbes = %d, want 2", p.statistics.OngoingSuccessfulProbes)
	}
}

// --- finalizeStatistics ---------------------------------------------------

func TestFinalizeStatistics_WhileDownAccruesRemainingDowntime(t *testing.T) {
	p, _ := newTestProber(nil, config.Config{})
	p.statistics.LastProbeHadFailed = true
	p.statistics.StartOfDowntime = time.Now().Add(-50 * time.Millisecond)

	p.finalizeStatistics()

	s := p.statistics
	if s.EndTime.IsZero() {
		t.Fatal("EndTime was not set")
	}
	if s.TotalDowntime < 50*time.Millisecond {
		t.Errorf("TotalDowntime = %v, want at least 50ms", s.TotalDowntime)
	}
	if s.LongestDowntime.Duration != s.TotalDowntime {
		t.Errorf("LongestDowntime.Duration = %v, want %v", s.LongestDowntime.Duration, s.TotalDowntime)
	}
}

func TestFinalizeStatistics_WhileUpAccruesRemainingUptime(t *testing.T) {
	p, _ := newTestProber(nil, config.Config{})
	p.statistics.StartOfUptime = time.Now().Add(-50 * time.Millisecond)

	p.finalizeStatistics()

	s := p.statistics
	if s.TotalUptime < 50*time.Millisecond {
		t.Errorf("TotalUptime = %v, want at least 50ms", s.TotalUptime)
	}
	if s.LongestUptime.Duration != s.TotalUptime {
		t.Errorf("LongestUptime.Duration = %v, want %v", s.LongestUptime.Duration, s.TotalUptime)
	}
}

func TestFinalizeStatistics_NoProbesYetIsANoop(t *testing.T) {
	p, _ := newTestProber(nil, config.Config{})
	p.statistics.StartTime = time.Now()

	p.finalizeStatistics()

	s := p.statistics
	if s.TotalUptime != 0 || s.TotalDowntime != 0 {
		t.Errorf("TotalUptime=%v TotalDowntime=%v, want both 0 when no probe ever ran", s.TotalUptime, s.TotalDowntime)
	}
	if s.EndTime.Before(s.StartTime) {
		t.Errorf("EndTime = %v, want it not to be before StartTime = %v", s.EndTime, s.StartTime)
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

	if err := p.Probe(ctx); err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	if got := pinger.calls(); got != 3 {
		t.Errorf("Ping called %d times, want 3", got)
	}
	if p.statistics.TotalSuccessfulProbes != 3 {
		t.Errorf("TotalSuccessfulProbes = %d, want 3", p.statistics.TotalSuccessfulProbes)
	}

	snap := printer.snapshot()
	if snap.startCalls != 1 {
		t.Errorf("PrintStart called %d times, want 1", snap.startCalls)
	}
	if snap.nameResolutionCalls != 0 {
		t.Errorf("PrintNameResolutionDuration called %d times, want 0 (no retry-resolve happened)", snap.nameResolutionCalls)
	}
	if snap.successCalls != 3 {
		t.Errorf("PrintProbeSuccess called %d times, want 3", snap.successCalls)
	}
}

// PrintNameResolutionDuration is now only about retry-resolve (the initial
// resolution time is folded into PrintStart's own line instead), so a
// successful retry must print it and update Statistics.NameResolutionDuration.
func TestProbe_PrintsNameResolutionDurationOnSuccessfulRetryResolve(t *testing.T) {
	pinger := alwaysFails()
	cfg := config.Config{
		IntervalBetweenProbes:      5 * time.Millisecond,
		ProbesBeforeQuit:           1,
		ShouldRetryResolve:         true,
		RetryResolveAfterNFailures: 1,
		// A literal IP resolves without touching the network.
		Resolver: dns.NewResolver("", time.Second, false, false, nic.NetworkInterface{}),
	}
	p, printer := newTestProber(pinger, cfg)
	p.statistics.Hostname = "127.0.0.1"

	if err := p.Probe(context.Background()); err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	if got := printer.snapshot().nameResolutionCalls; got != 1 {
		t.Errorf("PrintNameResolutionDuration called %d times, want 1", got)
	}
	if p.statistics.NameResolutionDuration < 0 {
		t.Errorf("NameResolutionDuration = %v, want >= 0", p.statistics.NameResolutionDuration)
	}
}

// A failed retry-resolve has no successful resolution to report a duration
// for - PrintError already covers it.
func TestProbe_DoesNotPrintNameResolutionDurationOnFailedRetryResolve(t *testing.T) {
	// A UDP socket that reads and discards every packet without ever
	// replying, so a lookup against it reliably times out instead of
	// racily depending on real network/DNS behavior.
	blackhole, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatalf("failed to start UDP listener: %v", err)
	}
	defer blackhole.Close()
	go func() {
		buf := make([]byte, 512)
		for {
			if _, _, err := blackhole.ReadFrom(buf); err != nil {
				return
			}
		}
	}()

	pinger := alwaysFails()
	cfg := config.Config{
		IntervalBetweenProbes:      5 * time.Millisecond,
		ProbesBeforeQuit:           1,
		ShouldRetryResolve:         true,
		RetryResolveAfterNFailures: 1,
		Resolver:                   dns.NewResolver(blackhole.LocalAddr().String(), 200*time.Millisecond, false, false, nic.NetworkInterface{}),
	}
	p, printer := newTestProber(pinger, cfg)
	p.statistics.Hostname = "example.com"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := p.Probe(ctx); err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	snap := printer.snapshot()
	if snap.nameResolutionCalls != 0 {
		t.Errorf("PrintNameResolutionDuration called %d times, want 0 (retry-resolve failed)", snap.nameResolutionCalls)
	}
	if snap.errorCalls != 1 {
		t.Errorf("PrintError called %d times, want 1", snap.errorCalls)
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
	if !p.statistics.EndTime.After(p.statistics.StartTime) {
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
	if err := p.Probe(ctx); err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	s := p.statistics
	if s.TotalUnsuccessfulProbes != 2 || s.TotalSuccessfulProbes != 1 {
		t.Errorf("TotalUnsuccessfulProbes=%d TotalSuccessfulProbes=%d, want 2 and 1", s.TotalUnsuccessfulProbes, s.TotalSuccessfulProbes)
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
		Resolver: dns.NewResolver("", time.Second, false, false, nic.NetworkInterface{}),
	}
	p, printer := newTestProber(pinger, cfg)
	p.statistics.Hostname = "127.0.0.1"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := p.Probe(ctx); err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	if p.statistics.RetriedHostnameLookups != 1 {
		t.Errorf("RetriedHostnameLookups = %d, want 1", p.statistics.RetriedHostnameLookups)
	}
	if p.statistics.ResolvedThisProbe {
		t.Error("ResolvedThisProbe = true, want false (failure-triggered retry, not ResolveEveryProbe)")
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

// --- ResolveEveryProbe ------------------------------------------------

func TestProbe_ResolveEveryProbe_ResolvesBeforeEachProbe(t *testing.T) {
	pinger := alwaysSucceeds()
	cfg := config.Config{
		IntervalBetweenProbes: 5 * time.Millisecond,
		ProbesBeforeQuit:      3,
		ResolveEveryProbe:     true,
		// A literal IP resolves without touching the network.
		Resolver: dns.NewResolver("", time.Second, false, false, nic.NetworkInterface{}),
	}
	p, printer := newTestProber(pinger, cfg)
	p.statistics.Hostname = "127.0.0.1"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := p.Probe(ctx); err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	if p.statistics.RetriedHostnameLookups != 3 {
		t.Errorf("RetriedHostnameLookups = %d, want 3 (one resolve per probe)", p.statistics.RetriedHostnameLookups)
	}
	if !p.statistics.ResolvedThisProbe {
		t.Error("ResolvedThisProbe = false, want true (so color/plain fold the duration into their own probe line)")
	}

	snap := printer.snapshot()
	if snap.nameResolutionCalls != 3 {
		t.Errorf("PrintNameResolutionDuration called %d times, want 3", snap.nameResolutionCalls)
	}
	if snap.retryCalls != 0 {
		t.Errorf("PrintRetryingToResolve called %d times, want 0 (this isn't a failure-triggered retry)", snap.retryCalls)
	}
}

// Resolving a literal IP target is a meaningless no-op, so it must be
// skipped entirely rather than printing a stream of "resolved in 0.000 ms".
func TestProbe_ResolveEveryProbe_SkippedForLiteralIPTarget(t *testing.T) {
	pinger := alwaysSucceeds()
	cfg := config.Config{
		IntervalBetweenProbes: 5 * time.Millisecond,
		ProbesBeforeQuit:      2,
		ResolveEveryProbe:     true,
		Resolver:              dns.NewResolver("", time.Second, false, false, nic.NetworkInterface{}),
	}
	p, printer := newTestProber(pinger, cfg)
	p.statistics.DestIsIP = true

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := p.Probe(ctx); err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	if p.statistics.RetriedHostnameLookups != 0 {
		t.Errorf("RetriedHostnameLookups = %d, want 0 (target is a literal IP)", p.statistics.RetriedHostnameLookups)
	}
	if got := printer.snapshot().nameResolutionCalls; got != 0 {
		t.Errorf("PrintNameResolutionDuration called %d times, want 0", got)
	}
}

// ResolveEveryProbe supersedes the failure-threshold retry entirely - it
// must not also trigger the old ShouldRetryResolve path.
func TestProbe_ResolveEveryProbe_TakesPrecedenceOverShouldRetryResolve(t *testing.T) {
	pinger := alwaysFails()
	cfg := config.Config{
		IntervalBetweenProbes:      5 * time.Millisecond,
		ProbesBeforeQuit:           2,
		ResolveEveryProbe:          true,
		ShouldRetryResolve:         true,
		RetryResolveAfterNFailures: 1,
		Resolver:                   dns.NewResolver("", time.Second, false, false, nic.NetworkInterface{}),
	}
	p, printer := newTestProber(pinger, cfg)
	p.statistics.Hostname = "127.0.0.1"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := p.Probe(ctx); err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	if p.statistics.RetriedHostnameLookups != 2 {
		t.Errorf("RetriedHostnameLookups = %d, want 2 (once per probe, not double-counted)", p.statistics.RetriedHostnameLookups)
	}
	if got := printer.snapshot().retryCalls; got != 0 {
		t.Errorf("PrintRetryingToResolve called %d times, want 0 (ResolveEveryProbe takes precedence)", got)
	}
}

// A retry-resolve that changes the target IP (including to a different
// address family) must actually change what gets dialed on the next probe,
// not just what gets displayed - otherwise every probe after the first
// keeps hitting the original, possibly now-stale, address forever.
func TestProbe_RetryResolveChangesTheActualDialTarget(t *testing.T) {
	pinger := alwaysFails()
	cfg := config.Config{
		IntervalBetweenProbes:      5 * time.Millisecond,
		ProbesBeforeQuit:           3,
		ShouldRetryResolve:         true,
		RetryResolveAfterNFailures: 1,
		// Literal IPs resolve without touching the network.
		Resolver: dns.NewResolver("", time.Second, false, false, nic.NetworkInterface{}),
	}
	p, _ := newTestProber(pinger, cfg)
	p.statistics.IP = netip.MustParseAddr("192.0.2.1")
	p.statistics.Hostname = "192.0.2.2" // what retry-resolve will "resolve" to

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := p.Probe(ctx); err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	ips := pinger.ips()
	if len(ips) != 3 {
		t.Fatalf("Ping was called %d times, want 3", len(ips))
	}

	want := netip.MustParseAddr("192.0.2.1")
	if ips[0] != want {
		t.Errorf("ips[0] = %v, want %v (the original target, before any retry)", ips[0], want)
	}

	want = netip.MustParseAddr("192.0.2.2")
	for i, ip := range ips[1:] {
		if ip != want {
			t.Errorf("ips[%d] = %v, want %v (the retry-resolved target)", i+1, ip, want)
		}
	}
}

// --- Probe: immediate first probe --------------------------------------

func TestProbe_ProbesImmediatelyWithoutWaitingForFirstTick(t *testing.T) {
	pinger := alwaysSucceeds()
	cfg := config.Config{
		// Deliberately much longer than the context timeout below, so a
		// probe can only have happened via the initial, un-ticked call.
		IntervalBetweenProbes: 500 * time.Millisecond,
	}
	p, _ := newTestProber(pinger, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := p.Probe(ctx); err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	if got := pinger.calls(); got != 1 {
		t.Errorf("Ping called %d times, want exactly 1 (the immediate, pre-tick probe)", got)
	}
}

// The immediate initial probe reuses the exact same runProbe body as every
// ticked probe, so it counts toward ProbesBeforeQuit like any other: a
// threshold of 1 now runs exactly one probe.
func TestProbe_ProbesBeforeQuitOfOneRunsOneProbe(t *testing.T) {
	pinger := alwaysSucceeds()
	cfg := config.Config{
		IntervalBetweenProbes: 5 * time.Millisecond,
		ProbesBeforeQuit:      1,
	}
	p, _ := newTestProber(pinger, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := p.Probe(ctx); err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	if got := pinger.calls(); got != 1 {
		t.Errorf("Ping called %d times, want exactly 1", got)
	}
}

// --- cancelled probes ---------------------------------------------------

// TestProbe_CancelledProbeIsNotAFailure covers hitting Ctrl+C while a probe
// is in flight. The probe returns an error because we cancelled it, but it
// must not be counted as a failure, or the run reports packet loss and a
// downtime that never happened.
func TestProbe_CancelledProbeIsNotAFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	pinger := &fakePinger{outcomeFn: func(call int) (ProbeResult, error) {
		if call == 0 {
			return ProbeResult{LocalAddr: &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)}}, nil
		}

		// The second probe is the one the user interrupts.
		cancel()
		return ProbeResult{}, ctx.Err()
	}}

	p, printer := newTestProber(pinger, config.Config{
		IntervalBetweenProbes: time.Millisecond,
	})

	if err := p.Probe(ctx); err != nil {
		t.Fatalf("Probe() error = %v, want no error", err)
	}

	s := p.statistics

	if s.TotalUnsuccessfulProbes != 0 {
		t.Errorf("TotalUnsuccessfulProbes = %d, want 0: a cancelled probe is not a failure", s.TotalUnsuccessfulProbes)
	}

	if s.TotalSuccessfulProbes != 1 {
		t.Errorf("TotalSuccessfulProbes = %d, want 1", s.TotalSuccessfulProbes)
	}

	if got := s.PacketLoss(); got != 0 {
		t.Errorf("PacketLoss() = %v, want 0", got)
	}

	if s.LongestDowntime.Duration != 0 {
		t.Errorf("LongestDowntime = %v, want 0: no downtime ever happened", s.LongestDowntime.Duration)
	}

	if got := printer.snapshot().failureCalls; got != 0 {
		t.Errorf("PrintProbeFailure was called %d times, want 0", got)
	}
}

// TestProbe_RealFailureStillCounts makes sure the cancellation check above
// only skips probes we cancelled, not genuine failures.
func TestProbe_RealFailureStillCounts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p, printer := newTestProber(alwaysFails(), config.Config{
		IntervalBetweenProbes: time.Millisecond,
		ProbesBeforeQuit:      2,
	})

	if err := p.Probe(ctx); err != nil {
		t.Fatalf("Probe() error = %v, want no error", err)
	}

	s := p.statistics

	if s.TotalUnsuccessfulProbes != 2 {
		t.Errorf("TotalUnsuccessfulProbes = %d, want 2", s.TotalUnsuccessfulProbes)
	}

	if got := printer.snapshot().failureCalls; got != 2 {
		t.Errorf("PrintProbeFailure was called %d times, want 2", got)
	}
}

// -show-failures-only hides successful probes from the printer, but the
// probe itself must still be counted, otherwise the summary would be wrong.
func TestProbe_ShowFailuresOnlyHidesSuccessesButStillCountsThem(t *testing.T) {
	pinger := alwaysSucceeds()
	cfg := config.Config{
		IntervalBetweenProbes: 5 * time.Millisecond,
		ProbesBeforeQuit:      3,
		ShowFailuresOnly:      true,
	}
	p, printer := newTestProber(pinger, cfg)

	if err := p.Probe(context.Background()); err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	if got := printer.snapshot().successCalls; got != 0 {
		t.Errorf("PrintProbeSuccess called %d times, want 0", got)
	}
	if p.statistics.TotalSuccessfulProbes != 3 {
		t.Errorf("TotalSuccessfulProbes = %d, want 3", p.statistics.TotalSuccessfulProbes)
	}
}

// Failures are the whole point of -show-failures-only, so they keep printing.
func TestProbe_ShowFailuresOnlyStillPrintsFailures(t *testing.T) {
	pinger := alwaysFails()
	cfg := config.Config{
		IntervalBetweenProbes: 5 * time.Millisecond,
		ProbesBeforeQuit:      2,
		ShowFailuresOnly:      true,
	}
	p, printer := newTestProber(pinger, cfg)

	if err := p.Probe(context.Background()); err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	if got := printer.snapshot().failureCalls; got != 2 {
		t.Errorf("PrintProbeFailure called %d times, want 2", got)
	}
}

// resolveHostname must record how long the lookup took, both on Statistics
// and on the HostnameChange entry it appends when the address changes.
func TestResolveHostname_RecordsDuration(t *testing.T) {
	cfg := config.Config{
		// A literal IP resolves without touching the network.
		Resolver: dns.NewResolver("", time.Second, false, false, nic.NetworkInterface{}),
	}
	p, _ := newTestProber(nil, cfg)
	p.statistics.Hostname = "127.0.0.1"

	if !p.resolveHostname(false) {
		t.Fatal("resolveHostname() = false, want true")
	}

	s := p.statistics
	if s.NameResolutionDuration < 0 {
		t.Errorf("NameResolutionDuration = %v, want >= 0", s.NameResolutionDuration)
	}
	if len(s.HostnameChanges) != 1 {
		t.Fatalf("len(HostnameChanges) = %d, want 1", len(s.HostnameChanges))
	}
	if s.HostnameChanges[0].Duration != s.NameResolutionDuration {
		t.Errorf("HostnameChanges[0].Duration = %v, want %v (same resolution)",
			s.HostnameChanges[0].Duration, s.NameResolutionDuration)
	}
}

// A resolution that does not change the address must still update
// NameResolutionDuration, without appending a redundant HostnameChange.
func TestResolveHostname_SameAddressUpdatesDurationOnly(t *testing.T) {
	cfg := config.Config{
		Resolver: dns.NewResolver("", time.Second, false, false, nic.NetworkInterface{}),
	}
	p, _ := newTestProber(nil, cfg)
	p.statistics.Hostname = "127.0.0.1"
	p.statistics.HostnameChanges = []stats.HostnameChange{
		{Addr: netip.MustParseAddr("127.0.0.1")},
	}

	if !p.resolveHostname(false) {
		t.Fatal("resolveHostname() = false, want true")
	}

	s := p.statistics
	if len(s.HostnameChanges) != 1 {
		t.Errorf("len(HostnameChanges) = %d, want 1 (no new entry for an unchanged address)", len(s.HostnameChanges))
	}
	if s.NameResolutionDuration < 0 {
		t.Errorf("NameResolutionDuration = %v, want >= 0", s.NameResolutionDuration)
	}
}
