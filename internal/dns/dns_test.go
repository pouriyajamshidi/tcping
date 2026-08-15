package dns

import (
	"context"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/probers"
)

func TestDNSDialAddress(t *testing.T) {
	tests := []struct {
		name      string
		dnsServer string
		want      string
	}{
		{"empty input", "", ""},
		{"ipv4 with port", "8.8.8.8:53", "8.8.8.8:53"},
		{"ipv4 without port uses default", "8.8.8.8", "8.8.8.8:" + DefaultPort},
		{"ipv6 without port uses default", "2001:4860:4860::8888", "[2001:4860:4860::8888]:" + DefaultPort},
		{"ipv6 with port", "[2001:4860:4860::8888]:53", "[2001:4860:4860::8888]:53"},
		{"hostname is ignored (not an IP)", "dns.google:53", ""},
		{"garbage input is ignored", "not-an-address", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDialAddress(tt.dnsServer)
			if got != tt.want {
				t.Errorf("dnsDialAddress(%q) = %q, want %q", tt.dnsServer, got, tt.want)
			}
		})
	}
}

func TestCreateDNSResolver_Defaults(t *testing.T) {
	resolver := createDNSResolver("")
	if !resolver.PreferGo {
		t.Error("expected PreferGo to be true")
	}
	if resolver.Dial == nil {
		t.Fatal("expected Dial to be set")
	}
}

// Spins up a local TCP listener as a stand-in DNS server and checks that
// Dial connects to it regardless of the address it's called with.
func TestCreateDNSResolver_OverridesAddress(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer ln.Close()

	accepted := make(chan struct{}, 1)
	go func() {
		if conn, err := ln.Accept(); err == nil {
			accepted <- struct{}{}
			conn.Close()
		}
	}()

	resolver := createDNSResolver(ln.Addr().String())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := resolver.Dial(ctx, "tcp", "this-address-should-be-ignored:53")
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	conn.Close()

	select {
	case <-accepted:
	case <-time.After(2 * time.Second):
		t.Fatal("listener never received a connection: override didn't happen")
	}
}

// When no DNS server is configured, Dial should fall through to whatever
// address it's given.
func TestCreateDNSResolver_NoOverride(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer ln.Close()

	accepted := make(chan struct{}, 1)
	go func() {
		if conn, err := ln.Accept(); err == nil {
			accepted <- struct{}{}
			conn.Close()
		}
	}()

	resolver := createDNSResolver("")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := resolver.Dial(ctx, "tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	conn.Close()

	select {
	case <-accepted:
	case <-time.After(2 * time.Second):
		t.Fatal("listener never received a connection")
	}
}

// dummyPrinter is a fake test implementation
// of a printer that does nothing.
type dummyPrinter struct{}

func (fp *dummyPrinter) PrintStart(_ string, _ uint16) {}
func (fp *dummyPrinter) PrintProbeSuccess(_ time.Time, _ string, _ config.Config, _ uint, _ string) {
}
func (fp *dummyPrinter) PrintProbeFailure(_ time.Time, _ config.Config, _ uint) {}
func (fp *dummyPrinter) PrintRetryingToResolve(_ string)                        {}
func (fp *dummyPrinter) PrintTotalDownTime(_ time.Duration)                     {}
func (fp *dummyPrinter) PrintStatistics(_ probers.Tcping)                       {}
func (fp *dummyPrinter) PrintError(_ string, _ ...any)                          {}

// createTestStats should be used to create new stats structs.
// it uses "127.0.0.1:12345" as default values, because
// [testServerListen] use the same values.
// It'll call t.Errorf if netip.ParseAddr has failed.
func createTestStats(t *testing.T) *probers.Tcping {
	_, err := netip.ParseAddr("127.0.0.1")
	s := probers.Tcping{
		Ticker: time.NewTicker(time.Second),
	}
	if err != nil {
		t.Errorf("ip parse: %v", err)
	}

	return &s
}

func TestSelectResolvedIPv4(t *testing.T) {
	userInputV4 := config.Config{
		UseIPv4: true,
	}

	stats := createTestStats(t)
	stats.Options = userInputV4

	var (
		ip1 = netip.MustParseAddr("172.20.10.238")
		ip2 = netip.MustParseAddr("8.8.8.8")
	)

	t.Run("IPv4 Selection", func(t *testing.T) {
		actual, _ := randomlySelectResolvedIP([]netip.Addr{ip1, ip2}, true, false)

		if !actual.IsValid() {
			t.Errorf("Expected an IP but got invalid address")
		}
		if actual != ip1 && actual != ip2 {
			t.Errorf("Expected an IP but got invalid address")
		}
	})
}

func TestSelectResolvedIPv6(t *testing.T) {
	userInputV6 := config.Config{
		UseIPv6: true,
	}

	stats := createTestStats(t)
	stats.Options = userInputV6

	var (
		ip1 = netip.MustParseAddr("2001:0db8:85a3:0000:0000:8a2e:0370:7334")
		ip2 = netip.MustParseAddr("2001:4860:4860::8888")
	)

	t.Run("IPv6 Selection", func(t *testing.T) {
		actual, _ := randomlySelectResolvedIP([]netip.Addr{ip1, ip2}, false, true)
		if !actual.IsValid() {
			t.Errorf("Expected an IP but got invalid address")
		}
		if actual != ip1 && actual != ip2 {
			t.Errorf("Expected an IP but got invalid address")
		}
	})
}
