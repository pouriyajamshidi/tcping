// Package dns handles all the name resolution logic for the program to function
package dns

import (
	"context"
	"math/rand"
	"net"
	"net/netip"
	"os"
	"time"
)

const (
	dnsTimeout = 2 * time.Second
)

// selectResolvedIP returns a single IPv4 or IPv6 address from the net.IP slice of resolved addresses
func selectResolvedIP(tcping *tcping, ipAddrs []netip.Addr) netip.Addr {
	var index int
	var ipList []netip.Addr
	var ip netip.Addr

	switch {
	case tcping.userInput.useIPv4:
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
			tcping.printError("Failed to find IPv4 address for %s", tcping.userInput.hostname)
			os.Exit(1)
		}

		if len(ipList) > 1 {
			index = rand.Intn(len(ipList))
		} else {
			index = 0
		}

		ip, _ = netip.ParseAddr(ipList[index].Unmap().String())

	case tcping.userInput.useIPv6:
		for _, ip := range ipAddrs {
			if ip.Is6() {
				ipList = append(ipList, ip)
			}
		}

		if len(ipList) == 0 {
			tcping.printError("Failed to find IPv6 address for %s", tcping.userInput.hostname)
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

// resolveHostname handles hostname resolution with a timeout value of a second
func resolveHostname(tcping *tcping) netip.Addr {
	ip, err := netip.ParseAddr(tcping.userInput.hostname)
	if err == nil {
		return ip
	}

	ctx, cancel := context.WithTimeout(context.Background(), dnsTimeout)
	defer cancel()

	ipAddrs, err := net.DefaultResolver.LookupNetIP(ctx, "ip", tcping.userInput.hostname)

	// Prevent tcping to exit if it has been running for a while
	if err != nil && (tcping.totalSuccessfulProbes != 0 || tcping.totalUnsuccessfulProbes != 0) {
		return tcping.userInput.ip
	} else if err != nil {
		tcping.printError("Failed to resolve %s: %s", tcping.userInput.hostname, err)
		os.Exit(1)
	}

	return selectResolvedIP(tcping, ipAddrs)
}

// retryResolveHostname retries resolving a hostname after certain number of failures
func retryResolveHostname(tcping *tcping) {
	if tcping.ongoingUnsuccessfulProbes >= tcping.userInput.retryHostnameLookupAfter {
		tcping.printRetryingToResolve(tcping.userInput.hostname)
		tcping.userInput.ip = resolveHostname(tcping)
		tcping.ongoingUnsuccessfulProbes = 0
		tcping.retriedHostnameLookups++

		// At this point hostnameChanges should have len > 0, but just in case
		if len(tcping.hostnameChanges) == 0 {
			return
		}

		lastAddr := tcping.hostnameChanges[len(tcping.hostnameChanges)-1].Addr
		if lastAddr != tcping.userInput.ip {
			tcping.hostnameChanges = append(tcping.hostnameChanges, hostnameChange{
				Addr: tcping.userInput.ip,
				When: time.Now(),
			})
		}
	}
}
