package pingers_test

import (
	"context"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/pingers"
)

func TestNewTCPPinger(t *testing.T) {
	ip := netip.MustParseAddr("192.168.1.1")
	port := uint16(80)

	pinger := pingers.NewTCPPinger(ip, port)

	if pinger == nil {
		t.Fatal("NewTCPPinger() returned nil")
	}

	if pinger.IP().Compare(ip) != 0 {
		t.Errorf("IP() = %q, want %q", pinger.IP(), ip.String())
	}

	if pinger.Port() != port {
		t.Errorf("Port() = %d, want %d", pinger.Port(), port)
	}
}

func TestNewTCPPinger_WithTimeout(t *testing.T) {
	ip := netip.MustParseAddr("10.0.0.1")
	port := uint16(443)
	timeout := 2 * time.Second

	pinger := pingers.NewTCPPinger(ip, port, pingers.WithTimeout(timeout))

	if pinger == nil {
		t.Fatal("NewTCPPinger() returned nil")
	}
}

func TestNewTCPPinger_WithDialer(t *testing.T) {
	ip := netip.MustParseAddr("10.0.0.1")
	port := uint16(443)

	dialer := &net.Dialer{
		Timeout: 3 * time.Second,
	}

	pinger := pingers.NewTCPPinger(ip, port, pingers.WithDialer(dialer))

	if pinger == nil {
		t.Fatal("NewTCPPinger() returned nil")
	}
}

func TestTCPPinger_Ping_Localhost(t *testing.T) {
	// start a test TCP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("start test server: %v", err)
	}
	defer listener.Close()

	// get the port
	addr := listener.Addr().(*net.TCPAddr)
	port := uint16(addr.Port)

	// accept connections in background
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	ip := netip.MustParseAddr("127.0.0.1")
	pinger := pingers.NewTCPPinger(ip, port)

	err = pinger.Ping(t.Context())
	if err != nil {
		t.Errorf("Ping() error = %v, expected nil", err)
	}
}

func TestTCPPinger_Ping_UnreachableHost(t *testing.T) {
	ip := netip.MustParseAddr("127.0.0.1")
	port := uint16(9) // unlikely to be open

	pinger := pingers.NewTCPPinger(ip, port, pingers.WithTimeout(100*time.Millisecond))

	err := pinger.Ping(t.Context())

	if err == nil {
		t.Error("Ping() expected error for unreachable host, got nil")
	}
}

func TestTCPPinger_Ping_ContextCancellation(t *testing.T) {
	ip := netip.MustParseAddr("192.0.2.1") // RFC 5737 documentation prefix
	port := uint16(80)

	pinger := pingers.NewTCPPinger(ip, port, pingers.WithTimeout(5*time.Second))

	ctx, cancel := context.WithCancel(t.Context())
	cancel() // cancel immediately

	err := pinger.Ping(ctx)

	if err == nil {
		t.Error("Ping() expected error for cancelled context, got nil")
	}
}

func TestTCPPinger_Ping_Timeout(t *testing.T) {
	ip := netip.MustParseAddr("192.0.2.1") // non-routable for timeout test
	port := uint16(80)

	pinger := pingers.NewTCPPinger(ip, port, pingers.WithTimeout(10*time.Millisecond))

	start := time.Now()
	err := pinger.Ping(t.Context())
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Ping() expected timeout error, got nil")
	}

	if elapsed > 200*time.Millisecond {
		t.Errorf("Ping() took %v, expected ~10ms timeout", elapsed)
	}
}

func TestTCPPinger_IP(t *testing.T) {
	tests := []struct {
		name string
		ip   netip.Addr
		want netip.Addr
	}{
		{
			name: "ipv4",
			ip:   netip.MustParseAddr("192.168.1.1"),
			want: netip.MustParseAddr("192.168.1.1"),
		},
		{
			name: "ipv6",
			ip:   netip.MustParseAddr("::1"),
			want: netip.MustParseAddr("::1"),
		},
		{
			name: "ipv4-mapped ipv6",
			ip:   netip.MustParseAddr("::ffff:192.168.1.1"),
			want: netip.MustParseAddr("::ffff:192.168.1.1"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pinger := pingers.NewTCPPinger(tt.ip, 80)
			got := pinger.IP()
			if got != tt.want {
				t.Errorf("IP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTCPPinger_Port(t *testing.T) {
	tests := []struct {
		name string
		port uint16
	}{
		{name: "http", port: 80},
		{name: "https", port: 443},
		{name: "custom", port: 8080},
		{name: "high port", port: 65535},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := netip.MustParseAddr("127.0.0.1")
			pinger := pingers.NewTCPPinger(ip, tt.port)
			got := pinger.Port()
			if got != tt.port {
				t.Errorf("Port() = %d, want %d", got, tt.port)
			}
		})
	}
}

func TestTCPPinger_MultipleOptions(t *testing.T) {
	ip := netip.MustParseAddr("10.0.0.1")
	port := uint16(443)

	dialer := &net.Dialer{
		Timeout: 1 * time.Second,
	}

	pinger := pingers.NewTCPPinger(
		ip,
		port,
		pingers.WithDialer(dialer),
		pingers.WithTimeout(2*time.Second),
	)

	if pinger == nil {
		t.Fatal("NewTCPPinger() with multiple options returned nil")
	}

	if pinger.IP() != ip {
		t.Errorf("IP() = %v, want %v", pinger.IP(), ip)
	}

	if pinger.Port() != port {
		t.Errorf("Port() = %d, want %d", pinger.Port(), port)
	}
}
