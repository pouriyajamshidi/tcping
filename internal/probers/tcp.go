package probers

import (
	"context"
	"net"
	"net/netip"
	"strconv"

	"github.com/pouriyajamshidi/tcping/v3/internal/config"
)

const tcp = "tcp"

type Tcping struct {
	dialer *net.Dialer
	ip     netip.Addr
	port   uint16
}

func NewTcping(cfg config.Config) Tcping {
	if cfg.NetworkInterface.Use {
		cfg.NetworkInterface.Dialer.Timeout = cfg.Timeout
		return Tcping{dialer: &cfg.NetworkInterface.Dialer, ip: cfg.IP, port: cfg.Port}
	}

	return Tcping{dialer: &net.Dialer{Timeout: cfg.Timeout}, ip: cfg.IP, port: cfg.Port}
}

func (t *Tcping) address() string {
	return net.JoinHostPort(t.ip.String(), strconv.Itoa(int(t.port)))
}

func (t Tcping) Ping(ctx context.Context) error {
	conn, err := t.dialer.DialContext(ctx, tcp, t.address())
	if err != nil {
		return err
	}

	defer conn.Close()

	return nil
}
