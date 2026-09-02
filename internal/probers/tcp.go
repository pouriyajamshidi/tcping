package probers

import (
	"context"
	"net/netip"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/nic"
)

const tcp = "tcp"

type Tcping struct {
	networkInterface nic.NetworkInterface
	timeout          time.Duration
	port             uint16
}

func NewTcping(cfg config.Config) Tcping {
	return Tcping{
		networkInterface: cfg.NetworkInterface,
		timeout:          cfg.Timeout,
		port:             cfg.Port,
	}
}

// Ping dials ip:port, sourcing the connection from the configured network
// interface when there is one.
func (t Tcping) Ping(ctx context.Context, ip netip.Addr) (ProbeResult, error) {
	d, err := dialer(tcp, t.networkInterface, t.timeout, ip)
	if err != nil {
		return ProbeResult{}, err
	}

	conn, err := d.DialContext(ctx, tcp, address(ip, t.port))
	if err != nil {
		return ProbeResult{}, err
	}
	defer conn.Close()

	return ProbeResult{LocalAddr: conn.LocalAddr()}, nil
}
