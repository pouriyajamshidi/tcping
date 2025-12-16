package dns_test

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/dns"
)

func TestResolver_ResolveHostname_IPAddress(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		want     netip.Addr
	}{
		{
			name:     "ipv4 address",
			hostname: "192.168.1.1",
			want:     netip.MustParseAddr("192.168.1.1"),
		},
		{
			name:     "ipv6 address",
			hostname: "::1",
			want:     netip.MustParseAddr("::1"),
		},
		{
			name:     "ipv4 loopback",
			hostname: "127.0.0.1",
			want:     netip.MustParseAddr("127.0.0.1"),
		},
	}

	resolver := dns.NewResolver()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolver.ResolveHostname(t.Context(), tt.hostname)
			if err != nil {
				t.Fatalf("ResolveHostname() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("ResolveHostname() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolver_ResolveHostname_Localhost(t *testing.T) {
	resolver := dns.NewResolver()
	got, err := resolver.ResolveHostname(t.Context(), "localhost")
	if err != nil {
		t.Fatalf("ResolveHostname(localhost) error = %v", err)
	}

	if !got.IsLoopback() {
		t.Errorf("ResolveHostname(localhost) = %v, want loopback address", got)
	}
}

func TestResolver_WithIPv4Only(t *testing.T) {
	resolver := dns.NewResolver(dns.WithIPv4Only())
	got, err := resolver.ResolveHostname(t.Context(), "localhost")
	if err != nil {
		t.Fatalf("ResolveHostname() error = %v", err)
	}

	if !got.Is4() {
		t.Errorf("ResolveHostname() with IPv4Only = %v, want IPv4 address", got)
	}
}

func TestResolver_ResolveHostname_InvalidHostname(t *testing.T) {
	resolver := dns.NewResolver()
	_, err := resolver.ResolveHostname(t.Context(), "this-hostname-definitely-does-not-exist-12345.invalid")
	if err == nil {
		t.Error("ResolveHostname() expected error for invalid hostname")
	}

	if !errors.Is(err, dns.ErrResolve) {
		t.Errorf("ResolveHostname() error = %v, want ErrResolve", err)
	}
}

func TestResolver_ResolveHostname_ContextCancellation(t *testing.T) {
	resolver := dns.NewResolver()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := resolver.ResolveHostname(ctx, "example.com")
	if err == nil {
		t.Error("ResolveHostname() expected error for cancelled context")
	}
}

func TestResolver_ResolveHostname_Timeout(t *testing.T) {
	resolver := dns.NewResolver()
	ctx, cancel := context.WithDeadline(t.Context(), time.Now().Add(-1*time.Second))
	defer cancel()

	_, err := resolver.ResolveHostname(ctx, "example.com")
	if err == nil {
		t.Error("ResolveHostname() expected error for timed out context")
	}
}

func TestResolver_WithTimeout(t *testing.T) {
	resolver := dns.NewResolver(dns.WithTimeout(1 * time.Nanosecond))

	_, err := resolver.ResolveHostname(t.Context(), "example.com")
	if err == nil {
		t.Error("ResolveHostname() expected timeout error")
	}
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "ErrNoIPv4Address", err: dns.ErrNoIPv4Address},
		{name: "ErrNoIPv6Address", err: dns.ErrNoIPv6Address},
		{name: "ErrNoIPAddresses", err: dns.ErrNoIPAddresses},
		{name: "ErrResolve", err: dns.ErrResolve},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s is nil", tt.name)
			}
			if tt.err.Error() == "" {
				t.Errorf("%s has empty error message", tt.name)
			}
		})
	}
}
