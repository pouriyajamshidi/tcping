// Package dns handles all hostname resolution logic
package dns

import (
	"context"
	"math/rand"
	"net"
	"net/netip"
	"os"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/internal/consts"
	"github.com/pouriyajamshidi/tcping/v2/types"
)

// selectResolvedIP returns an IPv4, IPv6 or a random resolved address
// if the IP version usage is not enforced from the `net.IP` slice of received addresses
func selectResolvedIP(tcping *types.Tcping, ipAddrs []netip.Addr) netip.Addr {
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
	case tcping.Options.UseIPv4:
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
			tcping.PrintError("Failed to find an IPv4 address for %s", tcping.Options.Hostname)
			os.Exit(1)
		}

		return selectRandomIP(ipList)

	case tcping.Options.UseIPv6:
		for _, ip := range ipAddrs {
			if ip.Is6() {
				ipList = append(ipList, ip)
			}
		}

		if len(ipList) == 0 {
			tcping.PrintError("Failed to find an IPv6 address for %s", tcping.Options.Hostname)
			os.Exit(1)
		}

		return selectRandomIP(ipList)

	default:
		return selectRandomIP(ipAddrs)
	}
}

// ResolveHostname handles hostname resolution with a timeout value of `consts.DNSTimeout`
func ResolveHostname(tcping *types.Tcping) netip.Addr {
	// Ensure hostname is not an IP address
	ip, err := netip.ParseAddr(tcping.Options.Hostname)
	if err == nil {
		return ip
	}

	ctx, cancel := context.WithTimeout(context.Background(), consts.DNSTimeout)
	defer cancel()

	ipAddrs, err := net.DefaultResolver.LookupNetIP(ctx, "ip", tcping.Options.Hostname)

	// Prevent tcping from exiting if it has been running already
	if err != nil && (tcping.TotalSuccessfulProbes != 0 || tcping.TotalUnsuccessfulProbes != 0) {
		return tcping.Options.IP
	} else if err != nil {
		tcping.PrintError("Failed to resolve %s in %s seconds: %s ",
			tcping.Options.Hostname,
			consts.DNSTimeout.String(),
			err)
		os.Exit(1)
	}

	return selectResolvedIP(tcping, ipAddrs)
}

// RetryResolveHostname retries resolving a hostname after certain number of failures
func RetryResolveHostname(tcping *types.Tcping) {
	if tcping.OngoingUnsuccessfulProbes >= tcping.Options.RetryHostnameLookupAfter {
		tcping.PrintRetryingToResolve(tcping.Options.Hostname)
		tcping.Options.IP = ResolveHostname(tcping)
		tcping.OngoingUnsuccessfulProbes = 0
		tcping.RetriedHostnameLookups++

		// This is to report on hostname changes in the stats
		lastAddr := tcping.HostnameChanges[len(tcping.HostnameChanges)-1].Addr
		if lastAddr != tcping.Options.IP {
			tcping.HostnameChanges = append(tcping.HostnameChanges, types.HostnameChange{
				Addr: tcping.Options.IP,
				When: time.Now(),
			})
		}
	}
}
