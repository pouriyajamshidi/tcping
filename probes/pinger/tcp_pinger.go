package pinger

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/types"
)

type TCPProberOption func(*TCPPinger)

func WithIP(s string) TCPProberOption {
	return func(t *TCPPinger) {
		ip := netip.MustParseAddr(s)
		t.ip = ip
	}
}

func WithPort(port uint16) TCPProberOption {
	return func(t *TCPPinger) {
		t.port = port
	}
}

func WithNetworkInterface(nic *types.NetworkInterface) TCPProberOption {
	return func(t *TCPPinger) {
		t.nic = nic
	}
}

const (
	DefaultPort = 80
	DefaultIP   = "0.0.0.0"
)

var (
	DefaultIPAddr  = netip.MustParseAddr(DefaultIP)
	DefaultNIC     = &types.NetworkInterface{Use: false}
	DefaultTimeout = 1 * time.Second
)

func NewTCPPinger(opts ...TCPProberOption) *TCPPinger {
	t := &TCPPinger{
		ip:      netip.MustParseAddr(DefaultIP),
		port:    DefaultPort,
		nic:     DefaultNIC,
		timeout: DefaultTimeout,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

type TCPPinger struct {
	ip   netip.Addr
	port uint16
	nic  *types.NetworkInterface

	timeout time.Duration
}

var ErrPingCompleted = errors.New("ping completed")

func (t *TCPPinger) Ping(ctx context.Context) error {
	// Create a derived context with the timeout
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	connCh := make(chan net.Conn, 1)
	errCh := make(chan error, 1)

	go func() {
		conn, err := t.connect(ctx)
		if err != nil {
			// Check if the error is related to context cancellation
			if ctx.Err() != nil {
				// Don't propagate the error, the select will catch ctx.Done()
				return
			}
			errCh <- err
			return
		}
		connCh <- conn
	}()

	select {
	case conn := <-connCh:
		defer conn.Close()
		return nil
	case err := <-errCh:
		// Only pass through non-context related errors
		return err
	case <-ctx.Done():
		// This is not an error state, but the completion of the window
		return ErrPingCompleted
	}
}

func (t *TCPPinger) connect(ctx context.Context) (net.Conn, error) {
	if t.nic.Use {
		// The timeout value of this Dialer is set inside the `newNetworkInterface` function
		return t.nic.Dialer.DialContext(ctx, "tcp", t.nic.RemoteAddr.String())
	}

	dialer := &net.Dialer{}
	return dialer.DialContext(ctx, "tcp", netip.AddrPortFrom(t.ip, t.port).String())
}

func (t *TCPPinger) IP() string {
	return t.ip.String()
}

func (t *TCPPinger) Port() uint16 {
	return t.port
}
