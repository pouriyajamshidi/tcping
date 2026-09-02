// Package dns handles all the name resolution logic for the program to function
package dns

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/netip"
	"strings"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/nic"
)

const (
	DefaultTimeout = 2 * time.Second
	DefaultPort    = "53"
)

// IPv4OrIPv6 allows LookupNetIP to use both IPv4 and IPv6 addresses
const IPv4OrIPv6 = "ip"

var (
	ErrNoIPv4Address = errors.New("no ipv4 address found")
	ErrNoIPv6Address = errors.New("no ipv6 address found")
	ErrNoIPAddresses = errors.New("no ip addresses")
)

type Resolver struct {
	resolver *net.Resolver
	timeout  time.Duration
	useIPv4  bool
	useIPv6  bool
}

// NewResolver creates a Resolver that queries DNSServer (or the system
// default, if empty), giving up after timeout (0 means no timeout). When
// networkInterface.Use is set, lookups are performed from its source
// address, matching the -I flag's interface for probes - using whichever
// of its addresses matches the DNS server's own address family.
func NewResolver(DNSServer string, timeout time.Duration, useIPv4, useIPv6 bool, networkInterface nic.NetworkInterface) *Resolver {
	return &Resolver{
		resolver: createDNSResolver(DNSServer, timeout, networkInterface),
		timeout:  timeout,
		useIPv4:  useIPv4,
		useIPv6:  useIPv6,
	}
}

// getDialAddress computes the override address for the resolver's Dial func.
// Returns "" if no override should happen.
func getDialAddress(DNSServer string) string {
	if DNSServer == "" {
		return ""
	}

	host, port := DNSServer, DefaultPort
	if h, p, err := net.SplitHostPort(DNSServer); err == nil {
		host, port = h, p
	}

	if ip, err := netip.ParseAddr(host); err == nil {
		return net.JoinHostPort(ip.String(), port)
	}

	return ""
}

// createDNSResolver creates a new net.Resolver and uses DNSServer as the DNS server IP
// or falls back to what is configured on the device if DNSServer is empty.
// It helps bypass incorrect OS DNS cache entries.
// DNSServer can be in 1.2.3.4 or 1.2.3.4:53 format.
// See https://github.com/pouriyajamshidi/tcping/issues/416 for more info.
// When networkInterface.Use is set, lookups are dialed from whichever of
// its addresses matches the DNS server's address family. dialTimeout of 0
// means no timeout, matching net.Dialer's own zero-value semantics.
func createDNSResolver(DNSServer string, dialTimeout time.Duration, networkInterface nic.NetworkInterface) *net.Resolver {
	dialAddress := getDialAddress(DNSServer)

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

// selectRandomIP returns an IPv4, IPv6 or a random resolved address,
// if the IP version usage is not enforced from the `net.IP` slice of received addresses
func selectRandomIP(ipAddrs []netip.Addr) (netip.Addr, error) {
	if len(ipAddrs) == 0 {
		return netip.Addr{}, ErrNoIPAddresses
	}
	return ipAddrs[rand.Intn(len(ipAddrs))], nil
}

func filterIPv4(ipAddrs []netip.Addr) []netip.Addr {
	var ipList []netip.Addr

	for _, ip := range ipAddrs {
		// static builds (CGO=0) return IPv4-mapped IPv6 addresses
		if ip.Is4() || ip.Is4In6() {
			ipList = append(ipList, ip.Unmap())
		}
	}
	return ipList
}

func filterIPv6(ipAddrs []netip.Addr) []netip.Addr {
	var ipList []netip.Addr

	for _, ip := range ipAddrs {
		if ip.Is6() && !ip.Is4In6() {
			ipList = append(ipList, ip)
		}
	}
	return ipList
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

	ipAddrs, err := r.resolver.LookupNetIP(ctx, IPv4OrIPv6, hostname)
	if err != nil {
		return ip, err
	}

	var filteredIPs []netip.Addr

	switch {
	case r.useIPv4:
		filteredIPs = filterIPv4(ipAddrs)
		if len(filteredIPs) == 0 {
			return netip.Addr{}, fmt.Errorf("%w: %s", ErrNoIPv4Address, hostname)
		}
	case r.useIPv6:
		filteredIPs = filterIPv6(ipAddrs)
		if len(filteredIPs) == 0 {
			return netip.Addr{}, fmt.Errorf("%w: %s", ErrNoIPv6Address, hostname)
		}
	default:
		filteredIPs = unmapAddresses(ipAddrs)
	}

	return selectRandomIP(filteredIPs)
}
