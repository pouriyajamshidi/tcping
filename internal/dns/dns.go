// Package dns handles all the name resolution logic for the program to function
package dns

import (
	"context"
	"math/rand"
	"net"
	"net/netip"
	"os"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/models"
	"github.com/pouriyajamshidi/tcping/v3/internal/printers"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
)

// DNSTimeout is the accepted duration when doing hostname resolution
const DNSTimeout = 2 * time.Second

// IPv4OrIPv6 allows LookupNetIP to use both IPv4 and IPv6 addresses
const IPv4OrIPv6 = "ip"

// selectResolvedIP returns an IPv4, IPv6 or a random resolved address
// if the IP version usage is not enforced from the `net.IP` slice of received addresses
// TODO: The `os.exit`s here cause program to exit even if it has been running for a while, fix it
func selectResolvedIP(p printers.Printer, s *stats.Statistics, useIPv4, useIPv6 bool, ipAddrs []netip.Addr) netip.Addr {
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
			p.PrintError("Failed to find an IPv4 address for %s", s.Hostname)
			os.Exit(1)
		}

		return selectRandomIP(ipList)

	case useIPv6:
		for _, ip := range ipAddrs {
			if ip.Is6() {
				ipList = append(ipList, ip)
			}
		}

		if len(ipList) == 0 {
			p.PrintError("Failed to find an IPv6 address for %s", s.Hostname)
			os.Exit(1)
		}

		return selectRandomIP(ipList)

	default:
		return selectRandomIP(ipAddrs)
	}
}

// ResolveHostname handles hostname resolution with a timeout value of `consts.DNSTimeout`
func ResolveHostname(p printers.Printer, s *stats.Statistics, useIPv4, useIPv6 bool) netip.Addr {
	// Ensure hostname is not an IP address
	ip, err := netip.ParseAddr(s.Hostname)
	if err == nil {
		return ip
	}

	ctx, cancel := context.WithTimeout(context.Background(), DNSTimeout)
	defer cancel()

	ipAddrs, err := net.DefaultResolver.LookupNetIP(ctx, IPv4OrIPv6, s.Hostname)

	// Prevent tcping from exiting if it has been running already
	if err != nil && (s.TotalSuccessfulProbes != 0 || s.TotalUnsuccessfulProbes != 0) {
		return s.IP
	} else if err != nil {
		p.PrintError("Failed to resolve %s in %s seconds: %s ",
			s.Hostname,
			DNSTimeout.String(),
			err)

		os.Exit(1)
	}

	return selectResolvedIP(p, s, useIPv4, useIPv6, ipAddrs)
}

// RetryResolveHostname retries resolving a hostname after certain number of failures
func RetryResolveHostname(p printers.Printer, s *stats.Statistics, after uint, useIPv4, useIPv6 bool) {
	if s.OngoingUnsuccessfulProbes >= after {
		p.PrintRetryingToResolve(s)
		s.IP = ResolveHostname(p, s, useIPv4, useIPv6)
		s.OngoingUnsuccessfulProbes = 0
		s.RetriedHostnameLookups++

		// This is to report on hostname changes in the stats
		lastAddr := s.HostnameChanges[len(s.HostnameChanges)-1].Addr
		if lastAddr != s.IP {
			s.HostnameChanges = append(s.HostnameChanges, models.HostnameChange{
				Addr: s.IP,
				When: time.Now(),
			})
		}
	}
}

// selectResolvedIP2 returns an IPv4, IPv6 or a random resolved address
// if the IP version usage is not enforced from the `net.IP` slice of received addresses
func selectResolvedIP2(ipAddrs []netip.Addr, useIPv4, useIPv6 bool) netip.Addr {
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
			// TODO: how to notify?
			// p.PrintError("Failed to find an IPv4 address for %s", s.Hostname)
			os.Exit(1)
		}

		return selectRandomIP(ipList)

	case useIPv6:
		for _, ip := range ipAddrs {
			if ip.Is6() {
				ipList = append(ipList, ip)
			}
		}

		if len(ipList) == 0 {
			// TODO: how to notify?
			// p.PrintError("Failed to find an IPv6 address for %s", s.Hostname)
			os.Exit(1)
		}

		return selectRandomIP(ipList)

	default:
		return selectRandomIP(ipAddrs)
	}
}

// ResolveHostname2 handles hostname resolution with a timeout value of `consts.DNSTimeout`
func ResolveHostname2(target string, useIPv4, useIPv6 bool) netip.Addr {
	// Ensure hostname is not an IP address
	ip, err := netip.ParseAddr(target)
	if err == nil {
		return ip
	}

	ctx, cancel := context.WithTimeout(context.Background(), DNSTimeout)
	defer cancel()

	ipAddrs, err := net.DefaultResolver.LookupNetIP(ctx, IPv4OrIPv6, target)

	// TODO: Implement the same logic
	// Prevent tcping from exiting if it has been running already
	// if err != nil && (s.TotalSuccessfulProbes != 0 || s.TotalUnsuccessfulProbes != 0) {
	// 	return s.IP
	// } else if err != nil {
	// 	p.PrintError("Failed to resolve %s in %s seconds: %s ",
	// 		s.Hostname,
	// 		DNSTimeout.String(),
	// 		err)

	// 	os.Exit(1)
	// }

	return selectResolvedIP2(ipAddrs, useIPv4, useIPv6)
}
