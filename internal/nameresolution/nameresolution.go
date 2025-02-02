// Package nameresolution handles all hostname resolution logic
package nameresolution

import (
	"context"
	"math/rand"
	"net"
	"net/netip"
	"os"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/consts"
	"github.com/pouriyajamshidi/tcping/v2/types"
)

// selectResolvedIP returns a single IPv4 or IPv6 address from the net.IP slice of resolved addresses
func selectResolvedIP(tcping *types.Tcping, ipAddrs []netip.Addr) netip.Addr {
	var index int
	var ipList []netip.Addr
	var ip netip.Addr

	switch {
	case tcping.Options.UseIPv4:
		for _, ip := range ipAddrs {
			if ip.Is4() {
				ipList = append(ipList, ip)
			}
			// static builds (CGO=0) return IPv4-mapped IPv6 address
			if ip.Is4In6() {
				ipList = append(ipList, ip.Unmap())
			}
		}

		if len(ipList) == 0 {
			tcping.PrintError("Failed to find IPv4 address for %s", tcping.Options.Hostname)
			os.Exit(1)
		}

		if len(ipList) > 1 {
			index = rand.Intn(len(ipList))
		} else {
			index = 0
		}

		ip, _ = netip.ParseAddr(ipList[index].Unmap().String())

	case tcping.Options.UseIPv6:
		for _, ip := range ipAddrs {
			if ip.Is6() {
				ipList = append(ipList, ip)
			}
		}

		if len(ipList) == 0 {
			tcping.PrintError("Failed to find IPv6 address for %s", tcping.Options.Hostname)
			os.Exit(1)
		}

		if len(ipList) > 1 {
			index = rand.Intn(len(ipList))
		} else {
			index = 0
		}

		ip, _ = netip.ParseAddr(ipList[index].Unmap().String())

	default:
		if len(ipAddrs) > 1 {
			index = rand.Intn(len(ipAddrs))
		} else {
			index = 0
		}

		ip, _ = netip.ParseAddr(ipAddrs[index].Unmap().String())
	}

	return ip
}

// ResolveHostname handles hostname resolution with a timeout value of a second
func ResolveHostname(tcping *types.Tcping) netip.Addr {
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
		tcping.PrintError("Failed to resolve %s: %s", tcping.Options.Hostname, err)
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

		// At this point hostnameChanges should have len > 0, but just in case
		if len(tcping.HostnameChanges) == 0 {
			return
		}

		lastAddr := tcping.HostnameChanges[len(tcping.HostnameChanges)-1].Addr
		if lastAddr != tcping.Options.IP {
			tcping.HostnameChanges = append(tcping.HostnameChanges, types.HostnameChange{
				Addr: tcping.Options.IP,
				When: time.Now(),
			})
		}
	}
}
