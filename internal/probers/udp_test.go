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

// udpEchoServer starts a UDP listener on a free port that replies to every
// datagram with reply, which is what an answering port looks like to a
// probe. Passing a nil reply echoes the datagram back unchanged, the way
// `tcping --udp-server` does. It is closed when the test ends.
func udpEchoServer(t *testing.T, reply []byte) uint16 {
	t.Helper()

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatalf("failed to start the UDP echo server: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	go func() {
		buf := make([]byte, 1500)
		for {
			n, peer, err := conn.ReadFrom(buf)
			if err != nil {
				return
			}

			answer := reply
			if answer == nil {
				answer = buf[:n]
			}

			conn.WriteTo(answer, peer)
		}
	}()

	return uint16(conn.LocalAddr().(*net.UDPAddr).Port)
}

// udpBlackholePort starts a UDP listener that reads every datagram and never
// answers, which is the case a probe cannot draw a conclusion from.
func udpBlackholePort(t *testing.T) uint16 {
	t.Helper()

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatalf("failed to start the UDP listener: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	go func() {
		buf := make([]byte, 1500)
		for {
			if _, _, err := conn.ReadFrom(buf); err != nil {
				return
			}
		}
	}()

	return uint16(conn.LocalAddr().(*net.UDPAddr).Port)
}

// closedUDPPort asks the OS for a free UDP port and immediately releases it,
// so a probe sent to it gets an ICMP port unreachable back.
func closedUDPPort(t *testing.T) uint16 {
	t.Helper()

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatalf("failed to reserve a UDP port: %v", err)
	}
	port := conn.LocalAddr().(*net.UDPAddr).Port
	conn.Close()

	return uint16(port)
}

// --- NewUDPing -------------------------------------------------------------

func TestNewUDPing_UsesConfigPortAndTimeout(t *testing.T) {
	cfg := config.Config{
		Port:    53,
		Timeout: 2 * time.Second,
	}

	up := NewUDPing(cfg)

	if up.port != cfg.Port {
		t.Errorf("port = %v, want %v", up.port, cfg.Port)
	}
	if up.timeout != cfg.Timeout {
		t.Errorf("timeout = %v, want %v", up.timeout, cfg.Timeout)
	}
}

// --- Ping ------------------------------------------------------------------

func TestUDPPing_SucceedsAgainstAnEchoServer(t *testing.T) {
	up := UDPing{timeout: 2 * time.Second, port: udpEchoServer(t, nil)}

	result, err := up.Ping(context.Background(), netip.MustParseAddr("127.0.0.1"))
	if err != nil {
		t.Fatalf("Ping() error = %v, want nil", err)
	}
	if !result.Echoed {
		t.Error("result.Echoed = false, want true since the reply carried our payload back")
	}
	if result.ReplySize == 0 {
		t.Error("result.ReplySize = 0, want the size of the reply")
	}
	if result.LocalAddr == nil {
		t.Error("result.LocalAddr = nil, want the local address used for the probe")
	}
	if result.ProbeNumber != 1 {
		t.Errorf("result.ProbeNumber = %d, want 1", result.ProbeNumber)
	}
}

// Anything that answers proves the port is open, even when it is not an echo
// server and the reply has nothing to do with what we sent.
func TestUDPPing_SucceedsWithoutEchoWhenSomethingElseAnswers(t *testing.T) {
	up := UDPing{timeout: 2 * time.Second, port: udpEchoServer(t, []byte("something else"))}

	result, err := up.Ping(context.Background(), netip.MustParseAddr("127.0.0.1"))
	if err != nil {
		t.Fatalf("Ping() error = %v, want nil", err)
	}
	if result.Echoed {
		t.Error("result.Echoed = true, want false since the reply was not our payload")
	}
	if result.ReplySize != len("something else") {
		t.Errorf("result.ReplySize = %d, want %d", result.ReplySize, len("something else"))
	}
}

func TestUDPPing_ReportsRejectedAgainstAClosedPort(t *testing.T) {
	up := UDPing{timeout: 2 * time.Second, port: closedUDPPort(t)}

	result, err := up.Ping(context.Background(), netip.MustParseAddr("127.0.0.1"))
	if err == nil {
		t.Fatal("Ping() error = nil, want a port-unreachable error")
	}
	if !result.Rejected {
		t.Error("result.Rejected = false, want true since the target refused us")
	}
}

// Silence is not a refusal: the port may well be open and simply not
// answering, so the probe must not claim it was rejected.
func TestUDPPing_ReportsNoReplyFromASilentListener(t *testing.T) {
	up := UDPing{timeout: 200 * time.Millisecond, port: udpBlackholePort(t)}

	result, err := up.Ping(context.Background(), netip.MustParseAddr("127.0.0.1"))
	if err == nil {
		t.Fatal("Ping() error = nil, want a timeout error since nothing answered")
	}
	if result.Rejected {
		t.Error("result.Rejected = true, want false since nothing answered at all")
	}
	if result.ReplySize != 0 {
		t.Errorf("result.ReplySize = %d, want 0 since nothing answered", result.ReplySize)
	}
	if result.ProbeNumber != 1 {
		t.Errorf("result.ProbeNumber = %d, want 1 so the output can say which probe was lost", result.ProbeNumber)
	}
}

// A cancelled context has to cut the wait short, otherwise Ctrl+C would only
// take effect after the timeout expired.
func TestUDPPing_GivesUpWhenTheContextIsCancelled(t *testing.T) {
	up := UDPing{timeout: 10 * time.Second, port: udpBlackholePort(t)}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := up.Ping(ctx, netip.MustParseAddr("127.0.0.1"))
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Ping() error = nil, want an error since the context was cancelled")
	}
	if elapsed > 2*time.Second {
		t.Errorf("Ping() took %v, want it to give up close to the 50ms context deadline, not the 10s timeout", elapsed)
	}
}

// Every probe carries its own number, so a reply can be matched to the probe
// that asked for it.
func TestUDPPing_NumbersEveryProbe(t *testing.T) {
	up := NewUDPing(config.Config{Port: udpEchoServer(t, nil), Timeout: 2 * time.Second})

	for i := uint64(1); i <= 3; i++ {
		result, err := up.Ping(context.Background(), netip.MustParseAddr("127.0.0.1"))
		if err != nil {
			t.Fatalf("Ping() error = %v, want nil", err)
		}
		if up.probeNumber != i {
			t.Errorf("probeNumber = %d, want %d", up.probeNumber, i)
		}
		if result.ProbeNumber != i {
			t.Errorf("result.ProbeNumber = %d, want %d", result.ProbeNumber, i)
		}
	}
}

// A network interface with only an IPv4 address must not silently probe out
// of an unrelated interface when asked to reach an IPv6 target.
func TestUDPPing_FailsCleanlyWhenInterfaceHasNoMatchingFamily(t *testing.T) {
	up := UDPing{
		timeout: 2 * time.Second,
		port:    udpEchoServer(t, nil),
		networkInterface: nic.NetworkInterface{
			Use:        true,
			SourceIPv4: net.ParseIP("127.0.0.1"),
		},
	}

	_, err := up.Ping(context.Background(), netip.MustParseAddr("::1"))
	if err == nil {
		t.Fatal("Ping() error = nil, want an error since the interface has no IPv6 address")
	}
}

// The local address a UDP probe binds to has to be a *net.UDPAddr, otherwise
// net.Dialer rejects the dial outright.
func TestUDPPing_SucceedsWithMatchingInterfaceFamily(t *testing.T) {
	up := UDPing{
		timeout: 2 * time.Second,
		port:    udpEchoServer(t, nil),
		networkInterface: nic.NetworkInterface{
			Use:        true,
			SourceIPv4: net.ParseIP("127.0.0.1"),
		},
	}

	result, err := up.Ping(context.Background(), netip.MustParseAddr("127.0.0.1"))
	if err != nil {
		t.Fatalf("Ping() error = %v, want nil", err)
	}

	localIP := result.LocalAddr.(*net.UDPAddr).IP
	if !localIP.Equal(net.ParseIP("127.0.0.1")) {
		t.Errorf("LocalAddr IP = %v, want 127.0.0.1", localIP)
	}
}
