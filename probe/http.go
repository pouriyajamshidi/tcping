package probe

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/netip"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/version"

	"github.com/pouriyajamshidi/tcping/v3/config"
	"github.com/pouriyajamshidi/tcping/v3/nic"
)

// HTTPing probes an HTTP(S) endpoint. It is a TCP probe with the rest of the
// request on top, so a probe only succeeds when the server answers with a
// status code below 400.
type HTTPing struct {
	networkInterface nic.NetworkInterface
	timeout          time.Duration
	port             uint16
	url              string // full target URL, e.g. https://example.com/health
	hostname         string // used for the Host header and the TLS SNI
	skipTLSVerify    bool   // do not check the server certificate
}

func NewHTTPing(cfg config.Config) HTTPing {
	return HTTPing{
		networkInterface: cfg.NetworkInterface,
		timeout:          cfg.Timeout,
		port:             cfg.Port,
		url:              cfg.URL,
		hostname:         cfg.Hostname,
		skipTLSVerify:    cfg.SkipTLSVerify,
	}
}

// transport always dials ip:port instead of letting the HTTP client resolve
// the URL's hostname again, so the prober decides which address we probe and
// -r and -resolve-every-probe keep working. Keep-alives are off so every
// probe opens a real connection, which is what we are timing.
func (h HTTPing) transport(d net.Dialer, ip netip.Addr) *http.Transport {
	return &http.Transport{
		DisableKeepAlives: true,
		// Go only tries HTTP/2 on its own when the transport is left alone,
		// and ours is not, so ask for it. Otherwise we would always report
		// HTTP/1.1 even where the server speaks HTTP/2.
		ForceAttemptHTTP2: true,
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return d.DialContext(ctx, network, address(ip, h.port))
		},
		TLSClientConfig: &tls.Config{
			ServerName:         h.hostname,
			InsecureSkipVerify: h.skipTLSVerify,
		},
		TLSHandshakeTimeout: h.timeout,
	}
}

// Ping sends one GET to the target URL, sourcing the connection from the
// configured network interface when there is one.
func (h HTTPing) Ping(ctx context.Context, ip netip.Addr) (ProbeResult, error) {
	d, err := dialer(tcp, h.networkInterface, h.timeout, ip)
	if err != nil {
		return ProbeResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.url, nil)
	if err != nil {
		return ProbeResult{}, err
	}

	var result ProbeResult
	var connectStart, tlsStart time.Time
	start := time.Now()

	// These timings are what an HTTP probe can tell us that a plain TCP
	// connect cannot: where the time went.
	trace := &httptrace.ClientTrace{
		GotConn:      func(info httptrace.GotConnInfo) { result.LocalAddr = info.Conn.LocalAddr() },
		ConnectStart: func(string, string) { connectStart = time.Now() },
		ConnectDone: func(_, _ string, err error) {
			if err == nil {
				result.ConnectDuration = time.Since(connectStart)
			}
		},
		TLSHandshakeStart: func() { tlsStart = time.Now() },
		TLSHandshakeDone: func(_ tls.ConnectionState, err error) {
			if err == nil {
				result.TLSDuration = time.Since(tlsStart)
			}
		},
		GotFirstResponseByte: func() { result.TimeToFirstByte = time.Since(start) },
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	req.Header.Set("User-Agent", version.UserAgent)

	client := &http.Client{
		Timeout:   h.timeout,
		Transport: h.transport(d, ip),
		// A redirect is still an answer, and following it would dial a host
		// we never resolved, so report the 3xx as it is.
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return ProbeResult{}, err
	}
	defer resp.Body.Close()

	// Read the body out so the probe covers the whole response and not just
	// its headers.
	io.Copy(io.Discard, resp.Body)

	result.StatusCode = resp.StatusCode
	result.Status = resp.Status
	result.Proto = resp.Proto

	if resp.TLS != nil {
		result.TLSVersion = tls.VersionName(resp.TLS.Version)
		result.TLSCipherSuite = tls.CipherSuiteName(resp.TLS.CipherSuite)

		if len(resp.TLS.PeerCertificates) > 0 {
			result.CertExpiry = resp.TLS.PeerCertificates[0].NotAfter
		}
	}

	// An HTTP error is a failed probe: the host answers but is not serving.
	// The result is still returned so callers can report which status it was.
	if resp.StatusCode >= http.StatusBadRequest {
		return result, fmt.Errorf("%s", resp.Status)
	}

	return result, nil
}
