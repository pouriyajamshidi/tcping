package probe

import (
	"bytes"
	"context"
	"errors"
	"net/netip"
	"strconv"
	"syscall"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/config"
	"github.com/pouriyajamshidi/tcping/v3/nic"
)

const udp = "udp"

// maxReplySize is the buffer we read a reply into. A datagram larger than
// this gets truncated, which is fine: we only need to know that something
// answered and whether the start of it is what we sent.
const maxReplySize = 1500

// UDPing probes a UDP port. UDP has no handshake, so unlike TCP there is
// nothing to succeed or fail on its own: the probe only succeeds when
// something actually answers us. Each probe sends its own number as the
// payload so the reply can be matched against it, which is what
// `tcping --udp-server` on the other end (or an echo service like
// `socat UDP-RECVFROM:9999,fork EXEC:cat`) is for.
type UDPing struct {
	networkInterface nic.NetworkInterface
	timeout          time.Duration
	port             uint16
	probeNumber      uint64 // Sent as the payload, so a reply can be matched to its probe.
}

func NewUDPing(cfg config.Config) *UDPing {
	return &UDPing{
		networkInterface: cfg.NetworkInterface,
		timeout:          cfg.Timeout,
		port:             cfg.Port,
	}
}

// Ping sends one datagram to ip:port and waits for an answer, sourcing it
// from the configured network interface when there is one.
//
// There are three things that can happen, and only two of them are an
// answer: a reply means the port is open, an ICMP port unreachable means
// something refused us, and silence means we cannot tell whether the port
// is open or the packet was dropped on the way.
func (u *UDPing) Ping(ctx context.Context, ip netip.Addr) (ProbeResult, error) {
	d, err := dialer(udp, u.networkInterface, u.timeout, ip)
	if err != nil {
		return ProbeResult{}, err
	}

	// Dialing UDP sends nothing, it only binds the socket and remembers the
	// destination, so nothing is on the wire until the write below.
	conn, err := d.DialContext(ctx, udp, address(ip, u.port))
	if err != nil {
		return ProbeResult{}, err
	}
	defer conn.Close()

	// A read deadline is the only thing that unblocks the read, so when the
	// user hits Ctrl+C we set one in the past to give up right away.
	stop := context.AfterFunc(ctx, func() { conn.SetReadDeadline(time.Now()) })
	defer stop()

	if u.timeout > 0 {
		conn.SetDeadline(time.Now().Add(u.timeout))
	}

	u.probeNumber++
	payload := []byte(strconv.FormatUint(u.probeNumber, 10))

	// Carried on every result, including the failed ones, so the output can
	// say which probe was lost.
	result := ProbeResult{LocalAddr: conn.LocalAddr()}
	result.ProbeNumber = u.probeNumber

	if _, err := conn.Write(payload); err != nil {
		return result, err
	}

	reply := make([]byte, maxReplySize)

	n, err := conn.Read(reply)
	if err != nil {
		// The target sent back an ICMP port unreachable, which the kernel
		// hands us here because the socket is connected. Something is there
		// and it refused us. Any other error means nothing came back.
		if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNRESET) {
			result.Rejected = true
		}

		return result, err
	}

	result.ReplySize = n
	result.Echoed = bytes.Equal(reply[:n], payload)

	return result, nil
}
