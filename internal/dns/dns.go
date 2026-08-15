// Package dns handles all the name resolution logic for the program to function
package dns

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"net/netip"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
)

const (
	DefaultTimeout = 2 * time.Second
	DefaultPort    = "53"
)

// IPv4OrIPv6 allows LookupNetIP to use both IPv4 and IPv6 addresses
const IPv4OrIPv6 = "ip"

type Resolver struct {
	Resolver *net.Resolver
}

func NewResolver(DNSServer string) *Resolver {
	return &Resolver{
		Resolver: createDNSResolver(DNSServer),
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
func createDNSResolver(DNSServer string) *net.Resolver {
	dialAddress := getDialAddress(DNSServer)

	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			if dialAddress != "" {
				address = dialAddress
			}
			d := net.Dialer{Timeout: DefaultTimeout}
			return d.DialContext(ctx, network, address)
		},
	}
}

// RetryResolveHostname retries resolving a hostname after a certain number of failures
func (d *Resolver) RetryResolveHostname(s *stats.Statistics, useIPv4, useIPv6 bool) error {
	newIP, err := d.ResolveHostname(s.Hostname, useIPv4, useIPv6)
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

// randomlySelectResolvedIP returns an IPv4, IPv6 or a random resolved address
// if the IP version usage is not enforced from the `net.IP` slice of received addresses
func randomlySelectResolvedIP(ipAddrs []netip.Addr, useIPv4, useIPv6 bool) (netip.Addr, error) {
	selectRandomIP := func(ipList []netip.Addr) netip.Addr {
		var index int
		if len(ipList) > 1 {
			index = rand.Intn(len(ipList))
		} else {
			index = 0
		}

		return netip.MustParseAddr(ipList[index].Unmap().String())
	}

	var ipList []netip.Addr

	switch {
	case useIPv4:
		for _, ip := range ipAddrs {
			if ip.Is4() {
				ipList = append(ipList, ip)
			}
			// static builds (CGO=0) return IPv4-mapped IPv6 addresses
			if ip.Is4In6() {
				ipList = append(ipList, ip.Unmap())
			}
		}

		if len(ipList) == 0 {
			return netip.Addr{}, errors.New("Failed to find an IPv4 address")
		}

		return selectRandomIP(ipList), nil

	case useIPv6:
		for _, ip := range ipAddrs {
			if ip.Is6() {
				ipList = append(ipList, ip)
			}
		}

		if len(ipList) == 0 {
			return netip.Addr{}, errors.New("Failed to find an IPv6 address")
		}

		return selectRandomIP(ipList), nil

	default:
		return selectRandomIP(ipAddrs), nil
	}
}

// ResolveHostname handles hostname resolution with a timeout value of `DNSTimeout (2 seconds)`
func (d *Resolver) ResolveHostname(target string, useIPv4, useIPv6 bool) (netip.Addr, error) {
	// Ensure the target isn't already an IP address
	ip, err := netip.ParseAddr(target)
	if err == nil {
		return ip, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), DNSTimeout)
	defer cancel()

	ipAddrs, err := d.Resolver.LookupNetIP(ctx, IPv4OrIPv6, target)
	if err != nil {
		return ip, err
	}

	return randomlySelectResolvedIP(ipAddrs, useIPv4, useIPv6)
}
