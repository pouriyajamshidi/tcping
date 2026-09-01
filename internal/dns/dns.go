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

	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
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

// TODO: make use of these fields
type Resolver struct {
	Resolver *net.Resolver
	timeout  time.Duration
	useIPv4  bool
	useIPv6  bool
}

// NewResolver creates a Resolver that queries DNSServer (or the system
// default, if empty). When sourceIP is non-nil, lookups are performed from
// that source address, matching the -I flag's interface for probes.
func NewResolver(DNSServer string, timeout time.Duration, useIPv4, useIPv6 bool, sourceIP net.IP) *Resolver {
	return &Resolver{
		Resolver: createDNSResolver(DNSServer, sourceIP),
		timeout:  DefaultTimeout,
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
// When sourceIP is non-nil, lookups are dialed from that source address.
func createDNSResolver(DNSServer string, sourceIP net.IP) *net.Resolver {
	dialAddress := getDialAddress(DNSServer)

	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			if dialAddress != "" {
				address = dialAddress
			}

			d := net.Dialer{Timeout: DefaultTimeout}
			if sourceIP != nil {
				d.LocalAddr = localAddrForNetwork(network, sourceIP)
			}

			return d.DialContext(ctx, network, address)
		},
	}
}

// localAddrForNetwork builds the net.Addr type net.Dialer.LocalAddr requires
// for the given dial network ("udp"/"udp4"/"udp6" vs "tcp"/"tcp4"/"tcp6") -
// the resolver may dial either, depending on response size and OS resolver
// configuration.
func localAddrForNetwork(network string, ip net.IP) net.Addr {
	if strings.HasPrefix(network, "tcp") {
		return &net.TCPAddr{IP: ip}
	}
	return &net.UDPAddr{IP: ip}
}

// RetryResolveHostname retries resolving a hostname after a certain number of failures
func (r *Resolver) RetryResolveHostname(s *stats.Statistics) error {
	newIP, err := r.ResolveHostname(s.Hostname)
	if err != nil {
		return err
	}

	s.IP = newIP

	if len(s.HostnameChanges) == 0 || s.HostnameChanges[len(s.HostnameChanges)-1].Addr != newIP {
		s.HostnameChanges = append(s.HostnameChanges, stats.HostnameChange{
			Addr: newIP,
			When: time.Now(),
		})
	}

	return nil
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

// ResolveHostname handles hostname resolution with a timeout value of `DNSTimeout (2 seconds)`
func (r *Resolver) ResolveHostname(hostname string) (netip.Addr, error) {
	// Ensure the target isn't already an IP address
	ip, err := netip.ParseAddr(hostname)
	if err == nil {
		return ip, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	ipAddrs, err := r.Resolver.LookupNetIP(ctx, IPv4OrIPv6, hostname)
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
