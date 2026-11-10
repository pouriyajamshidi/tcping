// Package dns turns a target hostname into the IP address to probe. The
// resolver it builds honors the DNS server, the address family and the
// network interface the user asked for.
package dns

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"net/netip"
	"strings"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/nic"
)

// Defaults used when the user does not say otherwise.
const (
	DefaultTimeout = 2 * time.Second
	DefaultPort    = "53"
)

// Resolver looks up a hostname the way the user asked for it: through a
// given DNS server, limited to an address family, and out of a given network
// interface.
type Resolver struct {
	resolver         *net.Resolver
	timeout          time.Duration
	useIPv4          bool
	useIPv6          bool
	networkInterface nic.NetworkInterface
}

// getDialAddress computes the override address for the resolver's Dial func.
// Returns "" if no override should happen, and an error if dnsServer is not
// an IP address with an optional port.
func getDialAddress(dnsServer string) (string, error) {
	if dnsServer == "" {
		return "", nil
	}

	host, port := dnsServer, DefaultPort
	if h, p, err := net.SplitHostPort(dnsServer); err == nil {
		host, port = h, p
	}

	ip, err := netip.ParseAddr(host)
	if err != nil {
		return "", fmt.Errorf("invalid DNS server %q: want an IP address with an optional port, like 1.1.1.1, 1.1.1.1:53, 2606:4700:4700::1111 or [2606:4700:4700::1111]:53", dnsServer)
	}

	return net.JoinHostPort(ip.String(), port), nil
}

// localAddrForDial returns the net.Addr net.Dialer.LocalAddr should use to
// bind a DNS lookup to networkInterface's matching-family address, or nil
// if none is available. net.Dialer requires LocalAddr and the remote
// address to be the same family (both IPv4 or both IPv6); binding a
// mismatched family fails the dial outright with "no suitable address
// found", which is worse than just not binding and letting the OS pick
// a default route.
//
// A loopback DNS server (e.g. systemd-resolved's 127.0.0.53 stub) is left
// unbound too: the kernel can't route a packet sourced from a physical
// interface's address to a loopback destination, so forcing the bind would
// make every lookup time out instead of just not honoring -I for it.
func localAddrForDial(network, address string, networkInterface nic.NetworkInterface) net.Addr {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		host = address
	}

	serverIP, err := netip.ParseAddr(host)
	if err != nil {
		return nil
	}

	if serverIP.IsLoopback() {
		return nil
	}

	localIP := networkInterface.LocalIPFor(serverIP)
	if localIP == nil {
		return nil
	}

	if strings.HasPrefix(network, "tcp") {
		return &net.TCPAddr{IP: localIP}
	}
	return &net.UDPAddr{IP: localIP}
}

// createDNSResolver creates a new net.Resolver that sends every lookup to
// dialAddress, or to what is configured on the device if dialAddress is
// empty. It helps bypass incorrect OS DNS cache entries.
// See https://github.com/pouriyajamshidi/tcping/issues/416 for more info.
// When networkInterface.Use is set, lookups are dialed from whichever of
// its addresses matches the DNS server's address family. dialTimeout of 0
// means no timeout, matching net.Dialer's own zero-value semantics.
func createDNSResolver(dialAddress string, dialTimeout time.Duration, networkInterface nic.NetworkInterface) *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			if dialAddress != "" {
				address = dialAddress
			}

			d := net.Dialer{Timeout: dialTimeout}
			if networkInterface.Use {
				d.LocalAddr = localAddrForDial(network, address, networkInterface)
			}

			return d.DialContext(ctx, network, address)
		},
	}
}

// NewResolver creates a Resolver that queries dnsServer (or the system
// default, if empty), giving up after timeout (0 means no timeout). When
// networkInterface.Use is set, lookups are performed from its source
// address, matching the -I flag's interface for probes - using whichever
// of its addresses matches the DNS server's own address family.
//
// dnsServer must be an IP address, optionally with a port. Anything else,
// such as a hostname, is refused instead of quietly falling back to the
// system default.
func NewResolver(dnsServer string, timeout time.Duration, useIPv4, useIPv6 bool, networkInterface nic.NetworkInterface) (*Resolver, error) {
	dialAddress, err := getDialAddress(dnsServer)
	if err != nil {
		return nil, err
	}

	return &Resolver{
		resolver:         createDNSResolver(dialAddress, timeout, networkInterface),
		timeout:          timeout,
		useIPv4:          useIPv4,
		useIPv6:          useIPv6,
		networkInterface: networkInterface,
	}, nil
}

// selectRandomIP returns one of ipAddrs at random, out of the ones in the
// same family as the first. Go sorts the answers so the ones this machine
// can reach come first, so a machine without IPv6 never gets an IPv6
// address. When the interface from -I has an address of only one family,
// that family is picked instead.
func selectRandomIP(ipAddrs []netip.Addr, networkInterface nic.NetworkInterface) (netip.Addr, error) {
	if len(ipAddrs) == 0 {
		return netip.Addr{}, errors.New("no ip addresses")
	}

	wantIPv4 := ipAddrs[0].Is4()
	if networkInterface.Use {
		switch {
		case networkInterface.SourceIPv4 != nil && networkInterface.SourceIPv6 == nil:
			wantIPv4 = true
		case networkInterface.SourceIPv6 != nil && networkInterface.SourceIPv4 == nil:
			wantIPv4 = false
		default:
			// It has both, so the first address decides.
		}
	}

	var sameFamily []netip.Addr
	for _, ip := range ipAddrs {
		if ip.Is4() == wantIPv4 {
			sameFamily = append(sameFamily, ip)
		}
	}

	// None of the answers is of the interface's family, so the probe fails
	// whichever one is picked. Pick from all of them.
	if len(sameFamily) == 0 {
		sameFamily = ipAddrs
	}

	return sameFamily[rand.IntN(len(sameFamily))], nil
}

func unmapAddresses(ipAddrs []netip.Addr) []netip.Addr {
	ipList := make([]netip.Addr, len(ipAddrs))

	for i, ip := range ipAddrs {
		ipList[i] = ip.Unmap()
	}
	return ipList
}

// ResolveHostname handles hostname resolution, giving up after r.timeout
// (0 means no timeout).
func (r *Resolver) ResolveHostname(hostname string) (netip.Addr, error) {
	// Ensure the target isn't already an IP address
	ip, err := netip.ParseAddr(hostname)
	if err == nil {
		return ip, nil
	}

	ctx := context.Background()
	if r.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.timeout)
		defer cancel()
	}

	// Ask only for the family we will use, so a slow or failing answer for
	// the other one cannot fail the lookup.
	var network string
	switch {
	case r.useIPv4:
		network = "ip4"
	case r.useIPv6:
		network = "ip6"
	default:
		network = "ip"
	}

	ipAddrs, err := r.resolver.LookupNetIP(ctx, network, hostname)
	if err != nil {
		return ip, err
	}

	return selectRandomIP(unmapAddresses(ipAddrs), r.networkInterface)
}
