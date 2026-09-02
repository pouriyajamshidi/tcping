package probers

import (
	"context"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/nic"
)

// testServerListen creates a new listener
// on port 12345 and automatically starts it.
//
// Use t.Cleanup with srv.Close() to close it after
// the test, so that other tests are not affected.
//
// It could fail if net.Listen or Accept has failed.
func testServerListen(t *testing.T) net.Listener {
	srv, err := net.Listen("tcp", ":12345")
	if err != nil {
		// Fatal, not an error: carrying on would hand back a nil
		// listener and the accept loop below would panic instead of
		// saying what actually went wrong.
		t.Fatalf("test server: %v", err)
	}

	go func() {
		for {
			c, err := srv.Accept()
			if err != nil {
				return
			}

			c.Close()
		}
	}()

	return srv
}

// reservedButClosedPort asks the OS for a free TCP port and immediately
// releases it without ever accepting a connection on it, so a subsequent
// dial to it reliably gets a "connection refused" instead of a timeout.
func reservedButClosedPort(t *testing.T) uint16 {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to reserve a port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()

	return uint16(port)
}

// --- NewTcping ---------------------------------------------------------

func TestNewTcping_UsesConfigPortAndTimeout(t *testing.T) {
	cfg := config.Config{
		Port:    8080,
		Timeout: 2 * time.Second,
	}

	tp := NewTcping(cfg)

	if tp.port != cfg.Port {
		t.Errorf("port = %v, want %v", tp.port, cfg.Port)
	}
	if tp.timeout != cfg.Timeout {
		t.Errorf("timeout = %v, want %v", tp.timeout, cfg.Timeout)
	}
}

func TestNewTcping_UsesNetworkInterfaceWhenConfigured(t *testing.T) {
	cfg := config.Config{
		Port:    443,
		Timeout: 3 * time.Second,
		NetworkInterface: nic.NetworkInterface{
			Use:        true,
			SourceIPv4: net.IPv4(10, 0, 0, 5),
		},
	}

	tp := NewTcping(cfg)

	if !tp.networkInterface.Use {
		t.Error("networkInterface.Use = false, want true")
	}
	if !tp.networkInterface.SourceIPv4.Equal(net.IPv4(10, 0, 0, 5)) {
		t.Errorf("networkInterface.SourceIPv4 = %v, want 10.0.0.5", tp.networkInterface.SourceIPv4)
	}
}

// --- address -------------------------------------------------------------

func TestAddress(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		port uint16
		want string
	}{
		{"IPv4", "192.168.1.1", 8080, "192.168.1.1:8080"},
		{"IPv6 gets bracketed", "::1", 443, "[::1]:443"},
		{"loopback", "127.0.0.1", 12345, "127.0.0.1:12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := address(netip.MustParseAddr(tt.ip), tt.port); got != tt.want {
				t.Errorf("address() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- Ping ------------------------------------------------------------------

func TestPing_SucceedsAgainstAnOpenPort(t *testing.T) {
	srv := testServerListen(t)
	t.Cleanup(func() { srv.Close() })

	tp := Tcping{timeout: 2 * time.Second, port: 12345}

	result, err := tp.Ping(context.Background(), netip.MustParseAddr("127.0.0.1"))
	if err != nil {
		t.Fatalf("Ping() error = %v, want nil", err)
	}
	if result.LocalAddr == nil {
		t.Error("result.LocalAddr = nil, want the local address used for the connection")
	}
}

func TestPing_FailsAgainstAClosedPort(t *testing.T) {
	port := reservedButClosedPort(t)

	tp := Tcping{timeout: 2 * time.Second, port: port}

	result, err := tp.Ping(context.Background(), netip.MustParseAddr("127.0.0.1"))
	if err == nil {
		t.Fatal("Ping() error = nil, want a connection-refused error")
	}
	if result.LocalAddr != nil {
		t.Errorf("result.LocalAddr = %v, want nil on failure", result.LocalAddr)
	}
}

func TestPing_FailsWhenContextIsAlreadyCancelled(t *testing.T) {
	srv := testServerListen(t)
	t.Cleanup(func() { srv.Close() })

	tp := Tcping{timeout: 2 * time.Second, port: 12345}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := tp.Ping(ctx, netip.MustParseAddr("127.0.0.1"))
	if err == nil {
		t.Fatal("Ping() error = nil, want an error because the context was already cancelled")
	}
}

func TestPing_RespectsAShorterContextDeadlineThanTheDialerTimeout(t *testing.T) {
	// TEST-NET-1 (RFC 5737) is reserved for documentation and is never
	// routable, so dialing it reliably hangs (or is rejected) rather than
	// racily succeeding, regardless of what network this test runs on.
	tp := Tcping{timeout: 10 * time.Second, port: 81} // much longer than the context below

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := tp.Ping(ctx, netip.MustParseAddr("192.0.2.1"))
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Ping() error = nil, want an error since the target is unroutable")
	}
	if elapsed > 2*time.Second {
		t.Errorf("Ping() took %v, want it to give up close to the 50ms context deadline, not the 10s dialer timeout", elapsed)
	}
}

// A network interface with only an IPv4 address must not silently dial out
// from an unrelated interface when asked to reach an IPv6 target - it
// should fail the probe cleanly instead, since the whole point of -I is to
// test reachability from a specific, known source.
func TestPing_FailsCleanlyWhenInterfaceHasNoMatchingFamily(t *testing.T) {
	srv := testServerListen(t)
	t.Cleanup(func() { srv.Close() })

	tp := Tcping{
		timeout: 2 * time.Second,
		port:    12345,
		networkInterface: nic.NetworkInterface{
			Use:        true,
			SourceIPv4: net.ParseIP("127.0.0.1"),
		},
	}

	_, err := tp.Ping(context.Background(), netip.MustParseAddr("::1"))
	if err == nil {
		t.Fatal("Ping() error = nil, want an error since the interface has no IPv6 address")
	}
}

// When the interface has the matching family, the probe should succeed and
// bind from it as expected.
func TestPing_SucceedsWithMatchingInterfaceFamily(t *testing.T) {
	srv := testServerListen(t)
	t.Cleanup(func() { srv.Close() })

	tp := Tcping{
		timeout: 2 * time.Second,
		port:    12345,
		networkInterface: nic.NetworkInterface{
			Use:        true,
			SourceIPv4: net.ParseIP("127.0.0.1"),
		},
	}

	result, err := tp.Ping(context.Background(), netip.MustParseAddr("127.0.0.1"))
	if err != nil {
		t.Fatalf("Ping() error = %v, want nil", err)
	}

	localIP := result.LocalAddr.(*net.TCPAddr).IP
	if !localIP.Equal(net.ParseIP("127.0.0.1")) {
		t.Errorf("LocalAddr IP = %v, want 127.0.0.1", localIP)
	}
}
