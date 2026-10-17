package probe

import (
	"context"
	"crypto/tls"
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/config"
	"github.com/pouriyajamshidi/tcping/v3/nic"
)

const tcp = "tcp"

// Tcping probes a target by opening a TCP connection to it.
type Tcping struct {
	networkInterface nic.NetworkInterface
	timeout          time.Duration
	hostname         string // what the proxy is asked to connect to
	port             uint16
	proxy            *url.URL       // nil when the probe connects directly
	proxyAddr        netip.AddrPort // where the proxy is, zero without one
}

// NewTcping creates a TCP prober for the target in cfg.
func NewTcping(cfg config.Config) Tcping {
	return Tcping{
		networkInterface: cfg.NetworkInterface,
		timeout:          cfg.Timeout,
		hostname:         cfg.Hostname,
		port:             cfg.Port,
		proxy:            cfg.Proxy,
		proxyAddr:        cfg.ProxyAddr,
	}
}

// pingThroughProxy asks the proxy to connect to the hostname for us. The
// probe succeeds when the proxy says it got through, so the time it takes
// covers the way to the proxy and the proxy's own way to the target.
func (t Tcping) pingThroughProxy(ctx context.Context) (Result, error) {
	d, err := dialer(tcp, t.networkInterface, t.timeout, t.proxyAddr.Addr())
	if err != nil {
		return Result{}, err
	}

	conn, err := d.DialContext(ctx, tcp, t.proxyAddr.String())
	if err != nil {
		return Result{}, err
	}
	defer conn.Close()

	// The dial timeout only covered reaching the proxy. The deadline covers
	// the talk with it as well.
	if t.timeout > 0 {
		if err := conn.SetDeadline(time.Now().Add(t.timeout)); err != nil {
			return Result{}, err
		}
	}

	// A deadline is the only thing that unblocks a read, so when the user
	// hits Ctrl+C we set one in the past to give up right away.
	stop := context.AfterFunc(ctx, func() { _ = conn.SetDeadline(time.Now()) })
	defer stop()

	result := Result{LocalAddr: conn.LocalAddr()}

	if t.proxy.Scheme == "socks5" || t.proxy.Scheme == "socks5h" {
		return result, socks5Connect(conn, t.hostname, t.port, t.proxy.User)
	}

	if t.proxy.Scheme == "https" {
		conn = tls.Client(conn, &tls.Config{ServerName: t.proxy.Hostname()})
	}

	target := net.JoinHostPort(t.hostname, strconv.Itoa(int(t.port)))

	return result, httpConnect(conn, target, t.proxy.User)
}

// Ping dials ip:port, sourcing the connection from the configured network
// interface when there is one.
func (t Tcping) Ping(ctx context.Context, ip netip.Addr) (Result, error) {
	if t.proxy != nil {
		return t.pingThroughProxy(ctx)
	}

	d, err := dialer(tcp, t.networkInterface, t.timeout, ip)
	if err != nil {
		return Result{}, err
	}

	conn, err := d.DialContext(ctx, tcp, address(ip, t.port))
	if err != nil {
		return Result{}, err
	}
	defer conn.Close()

	return Result{LocalAddr: conn.LocalAddr()}, nil
}
