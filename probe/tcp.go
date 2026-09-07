package probe

import (
	"context"
	"net/netip"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/config"
	"github.com/pouriyajamshidi/tcping/v3/nic"
)

const tcp = "tcp"

// Tcping probes a target by opening a TCP connection to it.
type Tcping struct {
	networkInterface nic.NetworkInterface
	timeout          time.Duration
	port             uint16
}

// NewTcping creates a TCP prober for the target in cfg.
func NewTcping(cfg config.Config) Tcping {
	return Tcping{
		networkInterface: cfg.NetworkInterface,
		timeout:          cfg.Timeout,
		port:             cfg.Port,
	}
}

// Ping dials ip:port, sourcing the connection from the configured network
// interface when there is one.
func (t Tcping) Ping(ctx context.Context, ip netip.Addr) (Result, error) {
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
