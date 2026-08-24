package dns

import (
	"context"
	"net"
	"net/netip"
	"testing"
	"time"
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

func TestSelectResolvedIPv4(t *testing.T) {
	var (
		ip1 = netip.MustParseAddr("172.20.10.238")
		ip2 = netip.MustParseAddr("8.8.8.8")
	)

	t.Run("IPv4 Selection", func(t *testing.T) {
		actual, _ := selectRandomIP([]netip.Addr{ip1, ip2})

		if !actual.IsValid() {
			t.Errorf("Expected an IP but got invalid address")
		}
		if actual != ip1 && actual != ip2 {
			t.Errorf("Expected an IP but got invalid address")
		}
	})
}

func TestSelectResolvedIPv6(t *testing.T) {
	var (
		ip1 = netip.MustParseAddr("2001:0db8:85a3:0000:0000:8a2e:0370:7334")
		ip2 = netip.MustParseAddr("2001:4860:4860::8888")
	)

	t.Run("IPv6 Selection", func(t *testing.T) {
		actual, _ := selectRandomIP([]netip.Addr{ip1, ip2})
		if !actual.IsValid() {
			t.Errorf("Expected an IP but got invalid address")
		}
		if actual != ip1 && actual != ip2 {
			t.Errorf("Expected an IP but got invalid address")
		}
	})
}
