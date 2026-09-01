package probers

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strconv"
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

func address(ip netip.Addr, port uint16) string {
	return net.JoinHostPort(ip.String(), strconv.Itoa(int(port)))
}

// Ping dials ip:port. When a network interface is configured (-I), the
// connection is sourced from it, matching ip's address family - if the
// interface has no address of that family, the probe fails cleanly rather
// than silently dialing out from an unrelated, unrequested interface.
func (t Tcping) Ping(ctx context.Context, ip netip.Addr) (ProbeResult, error) {
	d := net.Dialer{Timeout: t.timeout}

	if t.networkInterface.Use {
		localIP := t.networkInterface.LocalIPFor(ip)
		if localIP == nil {
			return ProbeResult{}, fmt.Errorf("network interface has no source address to reach %s", ip)
		}
		d.LocalAddr = &net.TCPAddr{IP: localIP}
	}

	conn, err := d.DialContext(ctx, tcp, address(ip, t.port))
	if err != nil {
		return ProbeResult{}, err
	}
	defer conn.Close()

	return ProbeResult{LocalAddr: conn.LocalAddr()}, nil
}
