// Package pingers implements protocol-specific ping functionality for network connectivity testing.
package pingers

import (
	"context"
	"net"
	"net/netip"
	"strconv"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/options"
)

// TCPPinger implements the Pinger interface for TCP connectivity testing.
type TCPPinger struct {
	dialer *net.Dialer
	ip     netip.Addr
	port   uint16
}

// IP implements Pinger.
func (t *TCPPinger) IP() string {
	return t.ip.String()
}

const tcp = "tcp"

func (t *TCPPinger) address() string {
	return net.JoinHostPort(t.ip.String(), strconv.Itoa(int(t.port)))
}

// Ping implements Pinger.
func (t *TCPPinger) Ping(ctx context.Context) error {
	conn, err := t.dialer.DialContext(ctx, tcp, t.address())
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

// Port implements Pinger.
func (t *TCPPinger) Port() uint16 {
	return t.port
}

type TCPOptions = options.Option[TCPPinger]

// NewTCPPinger creates a new TCP pinger for the specified IP address and port with optional configuration.
func NewTCPPinger(ip netip.Addr, port uint16, opts ...TCPOptions) *TCPPinger {
	t := &TCPPinger{
		ip:   ip,
		port: port,
		dialer: &net.Dialer{
			Timeout: 5 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// WithDialer configures a custom net.Dialer for TCP connections.
func WithDialer(dialer *net.Dialer) TCPOptions {
	return func(t *TCPPinger) {
		t.dialer = dialer
	}
}

// WithTimeout configures the connection timeout for TCP dial operations.
func WithTimeout(timeout time.Duration) TCPOptions {
	return func(t *TCPPinger) {
		if t.dialer == nil {
			t.dialer = &net.Dialer{}
		}
		t.dialer.Timeout = timeout
	}
}
