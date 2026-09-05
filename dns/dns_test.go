package dns

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"slices"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/nic"
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
	resolver := createDNSResolver("", 2*time.Second, nic.NetworkInterface{})
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

	resolver := createDNSResolver(ln.Addr().String(), 2*time.Second, nic.NetworkInterface{})

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

	resolver := createDNSResolver("", 2*time.Second, nic.NetworkInterface{})

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

// When a network interface is given, DNS lookups must be dialed from its
// matching-family address, so resolution honors the -I flag the same way
// probes do.
func TestCreateDNSResolver_BindsToSourceIP(t *testing.T) {
	sourceIP := net.ParseIP("127.0.0.1")
	networkInterface := nic.NetworkInterface{Use: true, SourceIPv4: sourceIP}

	t.Run("tcp", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to start listener: %v", err)
		}
		defer ln.Close()

		resolver := createDNSResolver("", 2*time.Second, networkInterface)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		conn, err := resolver.Dial(ctx, "tcp", ln.Addr().String())
		if err != nil {
			t.Fatalf("Dial failed: %v", err)
		}
		defer conn.Close()

		localIP := conn.LocalAddr().(*net.TCPAddr).IP
		if !localIP.Equal(sourceIP) {
			t.Errorf("LocalAddr IP = %v, want %v", localIP, sourceIP)
		}
	})

	t.Run("udp", func(t *testing.T) {
		ln, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
		if err != nil {
			t.Fatalf("failed to start UDP listener: %v", err)
		}
		defer ln.Close()

		resolver := createDNSResolver("", 2*time.Second, networkInterface)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		conn, err := resolver.Dial(ctx, "udp", ln.LocalAddr().String())
		if err != nil {
			t.Fatalf("Dial failed: %v", err)
		}
		defer conn.Close()

		localIP := conn.LocalAddr().(*net.UDPAddr).IP
		if !localIP.Equal(sourceIP) {
			t.Errorf("LocalAddr IP = %v, want %v", localIP, sourceIP)
		}
	})
}

// A source IP of one address family must never be forced onto a dial to a
// server of the other family: net.Dialer requires LocalAddr and the remote
// address to match families, so doing so would fail the lookup outright
// instead of falling back to the default route.
func TestCreateDNSResolver_IgnoresMismatchedSourceIPFamily(t *testing.T) {
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start IPv4 listener: %v", err)
	}
	defer ln.Close()

	ipv6SourceIP := net.ParseIP("::1")
	networkInterface := nic.NetworkInterface{Use: true, SourceIPv6: ipv6SourceIP}
	resolver := createDNSResolver("", 2*time.Second, networkInterface)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := resolver.Dial(ctx, "tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("Dial failed (mismatched-family source IP should have been ignored, not forced): %v", err)
	}
	defer conn.Close()

	localIP := conn.LocalAddr().(*net.TCPAddr).IP
	if localIP.Equal(ipv6SourceIP) {
		t.Errorf("LocalAddr IP = %v, want anything but the mismatched-family source IP", localIP)
	}
}

// When the interface has addresses for both families, DNS lookups should
// bind to whichever one matches the server actually being dialed.
func TestCreateDNSResolver_PicksMatchingFamilyFromDualStackInterface(t *testing.T) {
	networkInterface := nic.NetworkInterface{
		Use:        true,
		SourceIPv4: net.ParseIP("127.0.0.1"),
		SourceIPv6: net.ParseIP("::1"),
	}

	ln, err := net.Listen("tcp6", "[::1]:0")
	if err != nil {
		t.Fatalf("failed to start IPv6 listener: %v", err)
	}
	defer ln.Close()

	resolver := createDNSResolver("", 2*time.Second, networkInterface)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := resolver.Dial(ctx, "tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	localIP := conn.LocalAddr().(*net.TCPAddr).IP
	if !localIP.Equal(net.ParseIP("::1")) {
		t.Errorf("LocalAddr IP = %v, want ::1 (the IPv6 source, matching the IPv6 server)", localIP)
	}
}

// blackholeListener returns the address of a UDP socket that reads and
// discards every packet it receives without ever replying, so a lookup
// dialed at it (DNS resolution tries UDP first) reliably hangs until its
// own timeout gives up. A TCP-only listener wouldn't do this: dialing UDP
// at a port nothing is bound to gets an almost-instant ICMP port
// unreachable, which is the opposite of what a hang test needs.
func blackholeListener(t *testing.T) net.Addr {
	t.Helper()

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatalf("failed to start UDP listener: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	go func() {
		buf := make([]byte, 512)
		for {
			if _, _, err := conn.ReadFrom(buf); err != nil {
				return
			}
			// received and discarded; never reply
		}
	}()

	return conn.LocalAddr()
}

// NewResolver must actually use the timeout it's given, not silently fall
// back to DefaultTimeout regardless of what's passed in.
func TestNewResolver_TimeoutIsConfigurable(t *testing.T) {
	addr := blackholeListener(t)

	r := NewResolver(addr.String(), 150*time.Millisecond, false, false, nic.NetworkInterface{})

	start := time.Now()
	_, err := r.ResolveHostname("example.com")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("ResolveHostname() error = nil, want a timeout error")
	}
	if elapsed > 1500*time.Millisecond {
		t.Errorf("ResolveHostname() took %v, want it to give up close to the configured 150ms timeout, not the 2s default", elapsed)
	}
}

// A timeout of 0 must mean "no deadline", matching net.Dialer's own
// zero-value semantics and the -t flag's documented convention - not an
// already-expired context that fails every lookup instantly.
func TestNewResolver_ZeroTimeoutMeansNoDeadline(t *testing.T) {
	addr := blackholeListener(t)

	r := NewResolver(addr.String(), 0, false, false, nic.NetworkInterface{})

	done := make(chan struct{})
	go func() {
		r.ResolveHostname("example.com")
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("ResolveHostname() returned almost immediately; a 0 timeout must not create an already-expired context")
	case <-time.After(300 * time.Millisecond):
		// Still running after 300ms with nothing responding - correctly
		// has no deadline instead of failing instantly.
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

// addrs turns a list of textual addresses into the slice the filters take,
// so the test cases below stay readable.
func addrs(t *testing.T, in ...string) []netip.Addr {
	t.Helper()

	out := make([]netip.Addr, len(in))
	for i, s := range in {
		addr, err := netip.ParseAddr(s)
		if err != nil {
			t.Fatalf("parsing %q: %v", s, err)
		}
		out[i] = addr
	}
	return out
}

func TestFilterIPv4(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "keeps only the IPv4 addresses",
			in:   []string{"1.2.3.4", "2001:db8::1", "5.6.7.8"},
			want: []string{"1.2.3.4", "5.6.7.8"},
		},
		{
			name: "an IPv4-mapped address counts as IPv4 and is unmapped",
			in:   []string{"::ffff:1.2.3.4"},
			want: []string{"1.2.3.4"},
		},
		{
			name: "nothing to keep",
			in:   []string{"2001:db8::1", "fe80::1"},
			want: nil,
		},
		{
			name: "empty input",
			in:   nil,
			want: nil,
		},
		{
			name: "the order of what is kept does not change",
			in:   []string{"5.6.7.8", "2001:db8::1", "1.2.3.4"},
			want: []string{"5.6.7.8", "1.2.3.4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterIPv4(addrs(t, tt.in...))
			want := addrs(t, tt.want...)

			if !slices.Equal(got, want) {
				t.Errorf("filterIPv4(%v) = %v, want %v", tt.in, got, want)
			}
		})
	}
}

func TestFilterIPv6(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "keeps only the IPv6 addresses",
			in:   []string{"1.2.3.4", "2001:db8::1", "fe80::1"},
			want: []string{"2001:db8::1", "fe80::1"},
		},
		{
			name: "an IPv4-mapped address is IPv4, so it is left out",
			in:   []string{"::ffff:1.2.3.4", "2001:db8::1"},
			want: []string{"2001:db8::1"},
		},
		{
			name: "nothing to keep",
			in:   []string{"1.2.3.4", "5.6.7.8"},
			want: nil,
		},
		{
			name: "empty input",
			in:   nil,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterIPv6(addrs(t, tt.in...))
			want := addrs(t, tt.want...)

			if !slices.Equal(got, want) {
				t.Errorf("filterIPv6(%v) = %v, want %v", tt.in, got, want)
			}
		})
	}
}

func TestUnmapAddresses(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "an IPv4-mapped address becomes a plain IPv4 one",
			in:   []string{"::ffff:1.2.3.4"},
			want: []string{"1.2.3.4"},
		},
		{
			name: "a real IPv6 address is left alone",
			in:   []string{"2001:db8::1"},
			want: []string{"2001:db8::1"},
		},
		{
			name: "a plain IPv4 address is left alone",
			in:   []string{"1.2.3.4"},
			want: []string{"1.2.3.4"},
		},
		{
			name: "a mixed list keeps its order",
			in:   []string{"2001:db8::1", "::ffff:5.6.7.8", "1.2.3.4"},
			want: []string{"2001:db8::1", "5.6.7.8", "1.2.3.4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unmapAddresses(addrs(t, tt.in...))
			want := addrs(t, tt.want...)

			if !slices.Equal(got, want) {
				t.Errorf("unmapAddresses(%v) = %v, want %v", tt.in, got, want)
			}
		})
	}
}

func TestSelectRandomIP(t *testing.T) {
	t.Run("a single address is the one that comes back", func(t *testing.T) {
		want := addrs(t, "1.2.3.4")

		got, err := selectRandomIP(want)
		if err != nil {
			t.Fatalf("selectRandomIP returned an unexpected error: %v", err)
		}

		if got != want[0] {
			t.Errorf("selectRandomIP = %v, want %v", got, want[0])
		}
	})

	t.Run("the choice is always one of the addresses given", func(t *testing.T) {
		in := addrs(t, "1.2.3.4", "5.6.7.8", "2001:db8::1")

		for range 50 {
			got, err := selectRandomIP(in)
			if err != nil {
				t.Fatalf("selectRandomIP returned an unexpected error: %v", err)
			}

			if !slices.Contains(in, got) {
				t.Fatalf("selectRandomIP = %v, which is not one of %v", got, in)
			}
		}
	})

	t.Run("an empty list is an error, not a zero address", func(t *testing.T) {
		if _, err := selectRandomIP(nil); !errors.Is(err, ErrNoIPAddresses) {
			t.Errorf("selectRandomIP(nil) error = %v, want %v", err, ErrNoIPAddresses)
		}
	})
}
