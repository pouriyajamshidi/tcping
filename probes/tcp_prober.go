package probes

import (
	"context"
	"net"
	"net/netip"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/types"
)

type TCPProberOption func(*TCPPinger)

func WithIP(ip netip.Addr) TCPProberOption {
	return func(t *TCPPinger) {
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
	DefaultIPAddr = netip.MustParseAddr(DefaultIP)      //nolint:gochecknoglobals // this is a default dat
	DefaultNIC    = &types.NetworkInterface{Use: false} //nolint:gochecknoglobals // this is a default dat
)

func NewTCPProber(opts ...TCPProberOption) *TCPPinger {
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

func (t *TCPPinger) Ping(ctx context.Context) error {
	conn, err := t.connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

func (t *TCPPinger) connect(ctx context.Context) (net.Conn, error) {
	if t.nic.Use {
		// The timeout value of this Dialer is set inside the `newNetworkInterface` function
		return t.nic.Dialer.DialContext(ctx, "tcp", t.nic.RemoteAddr.String())
	}
	dailer := &net.Dialer{
		Timeout: t.timeout,
	}
	return dailer.DialContext(ctx, "tcp", netip.AddrPortFrom(t.ip, t.port).String())
}
