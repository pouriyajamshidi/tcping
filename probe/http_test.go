package probe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/config"
	"github.com/pouriyajamshidi/tcping/v3/nic"
)

// httpingFor builds an HTTPing aimed at srv, the way config would for the
// URL the server is listening on.
func httpingFor(t *testing.T, srv *httptest.Server, path string) (HTTPing, netip.Addr) {
	t.Helper()

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("could not parse the test server URL %q: %v", srv.URL, err)
	}

	addrPort, err := netip.ParseAddrPort(u.Host)
	if err != nil {
		t.Fatalf("could not parse the test server address %q: %v", u.Host, err)
	}

	return HTTPing{
		timeout:       5 * time.Second,
		port:          addrPort.Port(),
		url:           srv.URL + path,
		hostname:      u.Hostname(),
		skipTLSVerify: true,
	}, addrPort.Addr()
}

func TestNewHTTPing_UsesConfig(t *testing.T) {
	cfg := config.Config{
		Hostname:      "example.com",
		URL:           "https://example.com/health",
		Port:          8443,
		Timeout:       3 * time.Second,
		SkipTLSVerify: true,
		NetworkInterface: nic.NetworkInterface{
			Use: true,
		},
	}

	h := NewHTTPing(cfg)

	if h.url != cfg.URL || h.hostname != cfg.Hostname || h.port != cfg.Port ||
		h.timeout != cfg.Timeout || !h.skipTLSVerify || !h.networkInterface.Use {
		t.Errorf("NewHTTPing(%+v) did not carry the config over: %+v", cfg, h)
	}
}

func TestHTTPing_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}))
	t.Cleanup(srv.Close)

	h, ip := httpingFor(t, srv, "/")

	result, err := h.Ping(context.Background(), ip)
	if err != nil {
		t.Fatalf("Ping() error = %v, want no error", err)
	}

	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}

	if result.Status != "200 OK" {
		t.Errorf("Status = %q, want %q", result.Status, "200 OK")
	}

	if result.Proto != "HTTP/1.1" {
		t.Errorf("Proto = %q, want %q", result.Proto, "HTTP/1.1")
	}

	if result.LocalAddr == nil {
		t.Error("LocalAddr is nil, want the address the probe was sourced from")
	}

	if result.ConnectDuration == 0 || result.TimeToFirstByte == 0 {
		t.Errorf("ConnectDuration = %v and TimeToFirstByte = %v, want both to be measured",
			result.ConnectDuration, result.TimeToFirstByte)
	}

	if result.TLSVersion != "" {
		t.Errorf("TLSVersion = %q for a plain HTTP probe, want an empty string", result.TLSVersion)
	}
}

// TestHTTPing_ErrorStatusFails covers the rule that a reachable host serving
// 4xx or 5xx is a failed probe, while its response is still handed back so
// the printers can say which status it was.
func TestHTTPing_ErrorStatusFails(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{"success", http.StatusOK, false},
		{"redirect", http.StatusFound, false},
		{"not found", http.StatusNotFound, true},
		{"service unavailable", http.StatusServiceUnavailable, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.status == http.StatusFound {
					http.Redirect(w, r, "http://somewhere-else.invalid/", tt.status)
					return
				}
				w.WriteHeader(tt.status)
			}))
			t.Cleanup(srv.Close)

			h, ip := httpingFor(t, srv, "/")

			result, err := h.Ping(context.Background(), ip)

			if (err != nil) != tt.wantErr {
				t.Errorf("Ping() error = %v, wantErr %v", err, tt.wantErr)
			}

			if result.StatusCode != tt.status {
				t.Errorf("StatusCode = %d, want %d even when the probe fails", result.StatusCode, tt.status)
			}
		})
	}
}

// TestHTTPing_DoesNotFollowRedirects makes sure a redirect is reported as it
// is. Following it would dial a host the prober never resolved.
func TestHTTPing_DoesNotFollowRedirects(t *testing.T) {
	var hits int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/moved", http.StatusMovedPermanently)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	h, ip := httpingFor(t, srv, "/")

	result, err := h.Ping(context.Background(), ip)
	if err != nil {
		t.Fatalf("Ping() error = %v, want no error", err)
	}

	if result.StatusCode != http.StatusMovedPermanently {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusMovedPermanently)
	}

	if hits != 1 {
		t.Errorf("the server was hit %d times, want 1: the redirect must not be followed", hits)
	}
}

// TestHTTPing_TLS checks that an HTTPS probe reports the TLS details, and
// that the handshake is timed separately from the TCP connect.
func TestHTTPing_TLS(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)

	h, ip := httpingFor(t, srv, "/")

	result, err := h.Ping(context.Background(), ip)
	if err != nil {
		t.Fatalf("Ping() error = %v, want no error", err)
	}

	if result.TLSVersion == "" || result.TLSCipherSuite == "" {
		t.Errorf("TLSVersion = %q and TLSCipherSuite = %q, want both to be set",
			result.TLSVersion, result.TLSCipherSuite)
	}

	if result.CertExpiry.IsZero() {
		t.Error("CertExpiry is zero, want the leaf certificate's expiry")
	}

	if result.TLSDuration == 0 {
		t.Error("TLSDuration is zero, want the handshake to be timed")
	}
}

// TestHTTPing_SkipTLSVerify covers the -insecure flag: the test server's
// certificate is self-signed, so the probe only gets through with it on.
func TestHTTPing_SkipTLSVerify(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)

	h, ip := httpingFor(t, srv, "/")
	h.skipTLSVerify = false

	if _, err := h.Ping(context.Background(), ip); err == nil {
		t.Error("Ping() with certificate checking on succeeded, want a verification error")
	}

	h.skipTLSVerify = true

	if _, err := h.Ping(context.Background(), ip); err != nil {
		t.Errorf("Ping() with -insecure error = %v, want no error", err)
	}
}

// TestHTTPing_DialsTheGivenIP is the whole point of the custom transport:
// the address comes from the prober, not from resolving the URL's hostname.
// The URL below points at a name that cannot resolve, so the probe can only
// work if the given IP is what gets dialed.
func TestHTTPing_DialsTheGivenIP(t *testing.T) {
	var gotHost string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHost = r.Host
	}))
	t.Cleanup(srv.Close)

	_, ip := httpingFor(t, srv, "/")

	u, _ := url.Parse(srv.URL)
	addrPort, _ := netip.ParseAddrPort(u.Host)

	h := HTTPing{
		timeout:  5 * time.Second,
		port:     addrPort.Port(),
		url:      "http://does-not-resolve.invalid/",
		hostname: "does-not-resolve.invalid",
	}

	result, err := h.Ping(context.Background(), ip)
	if err != nil {
		t.Fatalf("Ping() error = %v, want the probe to reach the server at the given IP", err)
	}

	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}

	if gotHost != "does-not-resolve.invalid" {
		t.Errorf("the server saw Host %q, want the URL's hostname to be kept", gotHost)
	}
}

func TestHTTPing_ConnectionRefused(t *testing.T) {
	port := reservedButClosedPort(t)

	h := HTTPing{
		timeout:  2 * time.Second,
		port:     port,
		url:      "http://127.0.0.1/",
		hostname: "127.0.0.1",
	}

	result, err := h.Ping(context.Background(), netip.MustParseAddr("127.0.0.1"))
	if err == nil {
		t.Fatal("Ping() to a closed port succeeded, want an error")
	}

	if result.StatusCode != 0 {
		t.Errorf("StatusCode = %d, want 0 when nothing answered", result.StatusCode)
	}
}

func TestHTTPing_CancelledContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)

	h, ip := httpingFor(t, srv, "/")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := h.Ping(ctx, ip); err == nil {
		t.Error("Ping() with a cancelled context succeeded, want an error")
	}
}

// TestHTTPing_InterfaceWithoutMatchingFamily mirrors the TCP prober: when
// -I names an interface that has no address of the target's family, the
// probe fails instead of going out of some other interface.
func TestHTTPing_InterfaceWithoutMatchingFamily(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)

	h, ip := httpingFor(t, srv, "/")
	// IPv6 only, while the test server listens on IPv4.
	h.networkInterface = nic.NetworkInterface{Use: true, SourceIPv6: netip.MustParseAddr("::1").AsSlice()}

	if _, err := h.Ping(context.Background(), ip); err == nil {
		t.Error("Ping() succeeded, want it to fail when the interface has no address of the target's family")
	}
}
