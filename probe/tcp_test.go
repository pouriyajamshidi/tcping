package probe

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/config"
	"github.com/pouriyajamshidi/tcping/v3/nic"
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

			_ = c.Close()
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
	_ = l.Close()

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
	t.Cleanup(func() { _ = srv.Close() })

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
	t.Cleanup(func() { _ = srv.Close() })

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
	t.Cleanup(func() { _ = srv.Close() })

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
	t.Cleanup(func() { _ = srv.Close() })

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

// --- Ping through a proxy ----------------------------------------------

// fakeHTTPProxy answers every CONNECT request with status and hands the
// request over, so a test can see what the probe asked the proxy for.
func fakeHTTPProxy(t *testing.T, status string) (netip.AddrPort, <-chan *http.Request) {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("fake proxy: %v", err)
	}
	t.Cleanup(func() { _ = l.Close() })

	requests := make(chan *http.Request, 1)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}

			req, err := http.ReadRequest(bufio.NewReader(conn))
			if err == nil {
				requests <- req
				_, _ = conn.Write([]byte("HTTP/1.1 " + status + "\r\n\r\n"))
			}
			_ = conn.Close()
		}
	}()

	return netip.MustParseAddrPort(l.Addr().String()), requests
}

// silentProxy accepts connections and never says a word, like a proxy that
// hangs while it tries to reach the target.
func silentProxy(t *testing.T) netip.AddrPort {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("silent proxy: %v", err)
	}
	t.Cleanup(func() { _ = l.Close() })

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}

			// Read until the probe hangs up, then hang up too.
			go func() {
				_, _ = io.Copy(io.Discard, conn)
				_ = conn.Close()
			}()
		}
	}()

	return netip.MustParseAddrPort(l.Addr().String())
}

// The probe asks the proxy for the hostname, not an IP, and dials the proxy
// address looked up at startup: neither the target nor the proxy's hostname
// resolves, so that is the only way it can succeed.
func TestPing_ThroughHTTPProxy(t *testing.T) {
	proxyAddr, requests := fakeHTTPProxy(t, "200 Connection established")

	tp := Tcping{
		timeout:   2 * time.Second,
		hostname:  "does-not-resolve.invalid",
		port:      22,
		proxy:     &url.URL{Scheme: "http", User: url.UserPassword("user", "secret"), Host: "proxy.does-not-resolve.invalid:3128"},
		proxyAddr: proxyAddr,
	}

	result, err := tp.Ping(context.Background(), netip.Addr{})
	if err != nil {
		t.Fatalf("Ping() error = %v, want the probe to go through the proxy", err)
	}

	req := <-requests
	if req.Method != http.MethodConnect || req.Host != "does-not-resolve.invalid:22" {
		t.Errorf("the proxy was asked for %s %s, want CONNECT does-not-resolve.invalid:22", req.Method, req.Host)
	}

	wantLogin := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:secret"))
	if got := req.Header.Get("Proxy-Authorization"); got != wantLogin {
		t.Errorf("Proxy-Authorization = %q, want %q", got, wantLogin)
	}

	if result.LocalAddr == nil {
		t.Error("LocalAddr is nil, want the address the connection to the proxy came from")
	}
}

// A proxy that will not connect is a failed probe, whatever the reason it
// gives.
func TestPing_ThroughHTTPProxyThatRefuses(t *testing.T) {
	proxyAddr, _ := fakeHTTPProxy(t, "403 Forbidden")

	tp := Tcping{
		timeout:   2 * time.Second,
		hostname:  "example.com",
		port:      22,
		proxy:     &url.URL{Scheme: "http", Host: proxyAddr.String()},
		proxyAddr: proxyAddr,
	}

	_, err := tp.Ping(context.Background(), netip.Addr{})
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Errorf("Ping() error = %v, want the proxy's 403", err)
	}
}

// An https:// proxy is spoken to over TLS, and its certificate is checked
// like any other: this one is self-signed, so the probe must not get through.
func TestPing_ThroughHTTPSProxyChecksItsCertificate(t *testing.T) {
	proxy := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	t.Cleanup(proxy.Close)

	proxyAddr := netip.MustParseAddrPort(proxy.Listener.Addr().String())

	tp := Tcping{
		timeout:   2 * time.Second,
		hostname:  "example.com",
		port:      22,
		proxy:     &url.URL{Scheme: "https", Host: proxyAddr.String()},
		proxyAddr: proxyAddr,
	}

	_, err := tp.Ping(context.Background(), netip.Addr{})

	var certErr *tls.CertificateVerificationError
	if !errors.As(err, &certErr) {
		t.Errorf("Ping() error = %v, want the proxy's certificate to be refused", err)
	}
}

// The timeout covers the whole talk with the proxy, not only reaching it, so
// a proxy that never answers cannot hold a probe up for longer.
func TestPing_ThroughProxyThatNeverAnswersTimesOut(t *testing.T) {
	proxyAddr := silentProxy(t)

	tp := Tcping{
		timeout:   200 * time.Millisecond,
		hostname:  "example.com",
		port:      22,
		proxy:     &url.URL{Scheme: "http", Host: proxyAddr.String()},
		proxyAddr: proxyAddr,
	}

	start := time.Now()
	_, err := tp.Ping(context.Background(), netip.Addr{})

	if err == nil {
		t.Fatal("Ping() error = nil, want a timeout")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("Ping() took %v, want it to give up after the 200ms timeout", elapsed)
	}
}

// Ctrl+C has to stop a probe stuck on a silent proxy right away, even with
// no timeout at all.
func TestPing_ThroughProxyStopsWhenCancelled(t *testing.T) {
	proxyAddr := silentProxy(t)

	tp := Tcping{
		hostname:  "example.com",
		port:      22,
		proxy:     &url.URL{Scheme: "http", Host: proxyAddr.String()},
		proxyAddr: proxyAddr,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := tp.Ping(ctx, netip.Addr{})

	if err == nil {
		t.Fatal("Ping() error = nil, want the cancelled probe to fail")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("Ping() took %v, want it to stop when the context was cancelled", elapsed)
	}
}

// socks5Request is what the probe asked a SOCKS5 proxy for.
type socks5Request struct {
	user, password string
	addressType    byte // 1 for IPv4, 4 for IPv6, 3 for a hostname
	host           string
	port           uint16
}

// serveSOCKS5 reads one SOCKS5 handshake off conn. A short read leaves the
// rest of the request empty, which the test then reports.
func serveSOCKS5(conn net.Conn, replyCode byte) socks5Request {
	read := func(n int) []byte {
		b := make([]byte, n)
		_, _ = io.ReadFull(conn, b)
		return b
	}

	var req socks5Request

	// The version, the number of login methods, and the one the probe offers.
	method := read(3)[2]
	_, _ = conn.Write([]byte{5, method})

	if method == 2 {
		// The login version, then the username and the password, each with
		// its length in front.
		req.user = string(read(int(read(2)[1])))
		req.password = string(read(int(read(1)[0])))
		_, _ = conn.Write([]byte{1, 0})
	}

	// The version, CONNECT, a reserved byte and the address type.
	req.addressType = read(4)[3]
	switch req.addressType {
	case 1:
		req.host = netip.AddrFrom4([4]byte(read(4))).String()
	case 4:
		req.host = netip.AddrFrom16([16]byte(read(16))).String()
	default: // 3, a hostname
		req.host = string(read(int(read(1)[0])))
	}
	req.port = binary.BigEndian.Uint16(read(2))

	_, _ = conn.Write([]byte{5, replyCode, 0, 1, 0, 0, 0, 0, 0, 0})

	return req
}

// fakeSOCKS5Proxy plays the proxy's side of every SOCKS5 handshake, answers
// the request with replyCode, and hands the request over, so a test can see
// what the probe asked for. A reply code of 0 means the proxy connected.
func fakeSOCKS5Proxy(t *testing.T, replyCode byte) (netip.AddrPort, <-chan socks5Request) {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("fake SOCKS5 proxy: %v", err)
	}
	t.Cleanup(func() { _ = l.Close() })

	requests := make(chan socks5Request, 1)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}

			requests <- serveSOCKS5(conn, replyCode)
			_ = conn.Close()
		}
	}()

	return netip.MustParseAddrPort(l.Addr().String()), requests
}

// The proxy gets the hostname as it is and looks it up on its own: neither
// the target nor the proxy's hostname resolves here, so the probe can only
// succeed by dialing the proxy address looked up at startup.
func TestPing_ThroughSOCKS5Proxy(t *testing.T) {
	for _, scheme := range []string{"socks5", "socks5h"} {
		t.Run(scheme, func(t *testing.T) {
			proxyAddr, requests := fakeSOCKS5Proxy(t, 0)

			tp := Tcping{
				timeout:   2 * time.Second,
				hostname:  "does-not-resolve.invalid",
				port:      22,
				proxy:     &url.URL{Scheme: scheme, Host: "proxy.does-not-resolve.invalid:1080"},
				proxyAddr: proxyAddr,
			}

			result, err := tp.Ping(context.Background(), netip.Addr{})
			if err != nil {
				t.Fatalf("Ping() error = %v, want the probe to go through the proxy", err)
			}

			want := socks5Request{addressType: 3, host: "does-not-resolve.invalid", port: 22}
			if got := <-requests; got != want {
				t.Errorf("the proxy was asked for %+v, want %+v", got, want)
			}

			if result.LocalAddr == nil {
				t.Error("LocalAddr is nil, want the address the connection to the proxy came from")
			}
		})
	}
}

func TestPing_ThroughSOCKS5ProxyWithALogin(t *testing.T) {
	proxyAddr, requests := fakeSOCKS5Proxy(t, 0)

	tp := Tcping{
		timeout:   2 * time.Second,
		hostname:  "example.com",
		port:      22,
		proxy:     &url.URL{Scheme: "socks5", User: url.UserPassword("user", "secret"), Host: proxyAddr.String()},
		proxyAddr: proxyAddr,
	}

	if _, err := tp.Ping(context.Background(), netip.Addr{}); err != nil {
		t.Fatalf("Ping() error = %v, want the probe to log in and go through", err)
	}

	if got := <-requests; got.user != "user" || got.password != "secret" {
		t.Errorf("the proxy got the login %q:%q, want user:secret", got.user, got.password)
	}
}

// An IP target goes to the proxy as an IP, not as a name the proxy would
// have to look up.
func TestPing_ThroughSOCKS5ProxyGivesAnIPAsAnIP(t *testing.T) {
	tests := []struct {
		target      string
		addressType byte
	}{
		{target: "192.0.2.1", addressType: 1},
		{target: "2001:db8::1", addressType: 4},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			proxyAddr, requests := fakeSOCKS5Proxy(t, 0)

			tp := Tcping{
				timeout:   2 * time.Second,
				hostname:  tt.target,
				port:      22,
				proxy:     &url.URL{Scheme: "socks5", Host: proxyAddr.String()},
				proxyAddr: proxyAddr,
			}

			if _, err := tp.Ping(context.Background(), netip.Addr{}); err != nil {
				t.Fatalf("Ping() error = %v, want the probe to go through the proxy", err)
			}

			want := socks5Request{addressType: tt.addressType, host: tt.target, port: 22}
			if got := <-requests; got != want {
				t.Errorf("the proxy was asked for %+v, want %+v", got, want)
			}
		})
	}
}

// A proxy that could not connect is a failed probe, whatever the reason it
// gives. Reply code 5 is "connection refused".
func TestPing_ThroughSOCKS5ProxyThatRefuses(t *testing.T) {
	proxyAddr, _ := fakeSOCKS5Proxy(t, 5)

	tp := Tcping{
		timeout:   2 * time.Second,
		hostname:  "example.com",
		port:      22,
		proxy:     &url.URL{Scheme: "socks5", Host: proxyAddr.String()},
		proxyAddr: proxyAddr,
	}

	if _, err := tp.Ping(context.Background(), netip.Addr{}); err == nil {
		t.Error("Ping() error = nil, want the refused connection to fail the probe")
	}
}
