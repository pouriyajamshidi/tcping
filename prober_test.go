package tcping_test

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3"
	"github.com/pouriyajamshidi/tcping/v3/statistics"
)

// mockPinger implements Pinger interface for testing
type mockPinger struct {
	ip       netip.Addr
	port     uint16
	pingErr  error
	calls    int
	pingFunc func(ctx context.Context) error
}

func (m *mockPinger) Ping(ctx context.Context) error {
	m.calls++
	if m.pingFunc != nil {
		return m.pingFunc(ctx)
	}
	return m.pingErr
}

func (m *mockPinger) IP() netip.Addr {
	return m.ip
}

func (m *mockPinger) Port() uint16 {
	return m.port
}

func (m *mockPinger) SetIP(ip netip.Addr) {
	m.ip = ip
}

// mockPrinter implements Printer interface for testing
type mockPrinter struct {
	startCalls           int
	successCalls         int
	failureCalls         int
	statisticsCalls      int
	retryResolveCalls    int
	totalDownTimeCalls   int
	errorCalls           int
	shutdownCalls        int
	lastStats            *statistics.Statistics
}

func (m *mockPrinter) PrintStart(s *statistics.Statistics) {
	m.startCalls++
	m.lastStats = s
}

func (m *mockPrinter) PrintProbeSuccess(s *statistics.Statistics) {
	m.successCalls++
	m.lastStats = s
}

func (m *mockPrinter) PrintProbeFailure(s *statistics.Statistics) {
	m.failureCalls++
	m.lastStats = s
}

func (m *mockPrinter) PrintRetryingToResolve(s *statistics.Statistics) {
	m.retryResolveCalls++
}

func (m *mockPrinter) PrintTotalDownTime(s *statistics.Statistics) {
	m.totalDownTimeCalls++
}

func (m *mockPrinter) PrintStatistics(s *statistics.Statistics) {
	m.statisticsCalls++
	m.lastStats = s
}

func (m *mockPrinter) PrintError(format string, args ...any) {
	m.errorCalls++
}

func (m *mockPrinter) Shutdown(s *statistics.Statistics) {
	m.shutdownCalls++
}

func TestNewProber(t *testing.T) {
	ip := netip.MustParseAddr("192.168.1.1")
	pinger := &mockPinger{ip: ip, port: 80}

	prober := tcping.NewProber(pinger)

	if prober == nil {
		t.Fatal("NewProber() returned nil")
	}

	if prober.Interval != tcping.DefaultInterval {
		t.Errorf("Interval = %v, want %v", prober.Interval, tcping.DefaultInterval)
	}

	if prober.Timeout != tcping.DefaultTimeout {
		t.Errorf("Timeout = %v, want %v", prober.Timeout, tcping.DefaultTimeout)
	}

	if prober.Statistics.IP != ip {
		t.Errorf("Statistics.IP = %v, want %v", prober.Statistics.IP, ip)
	}

	if prober.Statistics.Port != 80 {
		t.Errorf("Statistics.Port = %v, want 80", prober.Statistics.Port)
	}
}

func TestNewProberWithOptions(t *testing.T) {
	ip := netip.MustParseAddr("10.0.0.1")
	pinger := &mockPinger{ip: ip, port: 443}
	printer := &mockPrinter{}

	prober := tcping.NewProber(
		pinger,
		tcping.WithPrinter(printer),
		tcping.WithInterval(500*time.Millisecond),
		tcping.WithTimeout(2*time.Second),
		tcping.WithProbeCount(5),
		tcping.WithHostname("example.com"),
		tcping.WithShowFailuresOnly(true),
	)

	if prober.Interval != 500*time.Millisecond {
		t.Errorf("Interval = %v, want 500ms", prober.Interval)
	}

	if prober.ProbeCountLimit != 5 {
		t.Errorf("ProbeCountLimit = %v, want 5", prober.ProbeCountLimit)
	}

	if prober.Statistics.Hostname != "example.com" {
		t.Errorf("Statistics.Hostname = %q, want %q", prober.Statistics.Hostname, "example.com")
	}

	if prober.Statistics.DestIsIP {
		t.Error("DestIsIP should be false when hostname is set")
	}

	if !prober.Statistics.ShowFailuresOnly {
		t.Error("ShowFailuresOnly should be true")
	}
}

func TestProber_ProbeSuccess(t *testing.T) {
	ip := netip.MustParseAddr("192.168.1.1")
	pinger := &mockPinger{ip: ip, port: 80}
	printer := &mockPrinter{}

		prober := tcping.NewProber(
			pinger,
			tcping.WithPrinter(printer),
			tcping.WithInterval(100*time.Millisecond),
			tcping.WithTimeout(1*time.Second),
			tcping.WithProbeCount(3),
		)

		stats, err := prober.Probe(t.Context())

		if err != nil {
			t.Fatalf("Probe() error = %v", err)
		}

		if stats.Successful != 3 {
			t.Errorf("Successful probes = %d, want 3", stats.Successful)
		}

		if stats.Failed != 0 {
			t.Errorf("Failed probes = %d, want 0", stats.Failed)
		}

		if pinger.calls != 3 {
			t.Errorf("Pinger called %d times, want 3", pinger.calls)
		}

		if printer.startCalls != 1 {
			t.Errorf("PrintStart called %d times, want 1", printer.startCalls)
		}

		if printer.successCalls != 3 {
			t.Errorf("PrintProbeSuccess called %d times, want 3", printer.successCalls)
		}

	if len(stats.RTT) != 3 {
		t.Errorf("RTT array length = %d, want 3", len(stats.RTT))
	}
}

func TestProber_ProbeFailure(t *testing.T) {
	ip := netip.MustParseAddr("192.168.1.1")
	pinger := &mockPinger{
			ip:      ip,
			port:    80,
			pingErr: errors.New("connection refused"),
		}
		printer := &mockPrinter{}

		prober := tcping.NewProber(
			pinger,
			tcping.WithPrinter(printer),
			tcping.WithInterval(100*time.Millisecond),
			tcping.WithTimeout(1*time.Second),
			tcping.WithProbeCount(3),
		)

		stats, err := prober.Probe(t.Context())

		if err != nil {
			t.Fatalf("Probe() error = %v", err)
		}

		if stats.Successful != 0 {
			t.Errorf("Successful probes = %d, want 0", stats.Successful)
		}

		if stats.Failed != 3 {
			t.Errorf("Failed probes = %d, want 3", stats.Failed)
		}

		if printer.failureCalls != 3 {
			t.Errorf("PrintProbeFailure called %d times, want 3", printer.failureCalls)
		}

	if stats.DestWasDown != true {
		t.Error("DestWasDown should be true after failures")
	}
}

func TestProber_MixedResults(t *testing.T) {
	ip := netip.MustParseAddr("192.168.1.1")
	printer := &mockPrinter{}

	// fail first 2, succeed last 3
	callCount := 0
	pinger := &mockPinger{
		ip:   ip,
		port: 80,
		pingFunc: func(ctx context.Context) error {
			callCount++
			if callCount <= 2 {
				return errors.New("connection refused")
			}
			return nil
		},
	}

		prober := tcping.NewProber(
			pinger,
			tcping.WithPrinter(printer),
			tcping.WithInterval(100*time.Millisecond),
			tcping.WithTimeout(2*time.Second),
			tcping.WithProbeCount(5),
		)

		stats, err := prober.Probe(t.Context())

		if err != nil {
			t.Fatalf("Probe() error = %v", err)
		}

		if stats.Successful != 3 {
			t.Errorf("Successful probes = %d, want 3", stats.Successful)
		}

		if stats.Failed != 2 {
			t.Errorf("Failed probes = %d, want 2", stats.Failed)
		}

		if printer.successCalls != 3 {
			t.Errorf("PrintProbeSuccess called %d times, want 3", printer.successCalls)
		}

		if printer.failureCalls != 2 {
			t.Errorf("PrintProbeFailure called %d times, want 2", printer.failureCalls)
		}

	if printer.totalDownTimeCalls != 1 {
		t.Errorf("PrintTotalDownTime called %d times, want 1", printer.totalDownTimeCalls)
	}
}

func TestProber_ContextCancellation(t *testing.T) {
	ip := netip.MustParseAddr("192.168.1.1")
	ctx, cancel := context.WithCancel(t.Context())

	pinger := &mockPinger{
		ip:   ip,
		port: 80,
	}
	pinger.pingFunc = func(ctx context.Context) error {
		if pinger.calls == 1 {
			cancel()
		}
		return nil
	}
	printer := &mockPrinter{}

	prober := tcping.NewProber(
		pinger,
		tcping.WithPrinter(printer),
		tcping.WithInterval(100*time.Millisecond),
		tcping.WithTimeout(10*time.Second), // long timeout
	)

	stats, err := prober.Probe(ctx)

	if err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	if stats.Successful == 0 {
		t.Error("Expected at least 1 successful probe before cancellation")
	}

	if !stats.EndTime.After(stats.StartTime) {
		t.Error("EndTime should be after StartTime")
	}
}

func TestProber_Timeout(t *testing.T) {
	ip := netip.MustParseAddr("192.168.1.1")
	pinger := &mockPinger{
		ip:      ip,
		port:    80,
		pingErr: errors.New("connection timeout"),
	}
	printer := &mockPrinter{}

	prober := tcping.NewProber(
		pinger,
		tcping.WithPrinter(printer),
		tcping.WithInterval(100*time.Millisecond),
		tcping.WithTimeout(500*time.Millisecond), // short timeout
	)

	_, err := prober.Probe(t.Context())

	if !errors.Is(err, tcping.ErrTimeout) {
		t.Errorf("Expected ErrTimeout, got %v", err)
	}
}

func TestProber_NoProbeCountLimit(t *testing.T) {
	ip := netip.MustParseAddr("192.168.1.1")
	pinger := &mockPinger{ip: ip, port: 80}
	printer := &mockPrinter{}

	prober := tcping.NewProber(
		pinger,
		tcping.WithPrinter(printer),
		tcping.WithInterval(50*time.Millisecond),
		tcping.WithTimeout(500*time.Millisecond),
		// no probe count - should run until timeout
	)

	stats, err := prober.Probe(t.Context())

	if err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	if stats.Successful < 5 {
		t.Errorf("Expected at least 5 probes within timeout, got %d", stats.Successful)
	}
}

func TestProber_Statistics(t *testing.T) {
	ip := netip.MustParseAddr("10.0.0.1")
	pinger := &mockPinger{ip: ip, port: 443}
	printer := &mockPrinter{}

	prober := tcping.NewProber(
		pinger,
		tcping.WithPrinter(printer),
		tcping.WithInterval(100*time.Millisecond),
		tcping.WithTimeout(2*time.Second),
		tcping.WithProbeCount(5),
		tcping.WithHostname("example.com"),
	)

	stats, err := prober.Probe(t.Context())

	if err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	if stats.IP != ip {
		t.Errorf("IP = %v, want %v", stats.IP, ip)
	}

	if stats.Port != 443 {
		t.Errorf("Port = %v, want 443", stats.Port)
	}

	if stats.Hostname != "example.com" {
		t.Errorf("Hostname = %q, want %q", stats.Hostname, "example.com")
	}

	if stats.Protocol != "TCP" {
		t.Errorf("Protocol = %q, want %q", stats.Protocol, "TCP")
	}

	if stats.TotalSuccessfulProbes != 5 {
		t.Errorf("TotalSuccessfulProbes = %d, want 5", stats.TotalSuccessfulProbes)
	}

	if stats.StartTime.IsZero() {
		t.Error("StartTime should be set")
	}

	if stats.EndTime.IsZero() {
		t.Error("EndTime should be set")
	}

	if stats.UpTime == 0 {
		t.Error("UpTime should be non-zero")
	}

	if !stats.HasResults {
		t.Error("HasResults should be true")
	}
}

func TestProber_OngoingStreaks(t *testing.T) {
	ip := netip.MustParseAddr("192.168.1.1")
	pinger := &mockPinger{ip: ip, port: 80}
	printer := &mockPrinter{}

	prober := tcping.NewProber(
		pinger,
		tcping.WithPrinter(printer),
		tcping.WithInterval(100*time.Millisecond),
		tcping.WithTimeout(2*time.Second),
		tcping.WithProbeCount(3),
	)

	stats, err := prober.Probe(t.Context())

	if err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	if stats.OngoingSuccessfulProbes != 3 {
		t.Errorf("OngoingSuccessfulProbes = %d, want 3", stats.OngoingSuccessfulProbes)
	}

	if stats.OngoingUnsuccessfulProbes != 0 {
		t.Errorf("OngoingUnsuccessfulProbes = %d, want 0", stats.OngoingUnsuccessfulProbes)
	}
}
