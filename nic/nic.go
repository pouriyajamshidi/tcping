package nic

import (
	"fmt"
	"net"
	"net/netip"
)

// NetworkInterface represents a network interface used for connecting to the target.
type NetworkInterface struct {
	SourceIPv4 net.IP // Source IPv4 address to use for outgoing connections, if the interface has one.
	SourceIPv6 net.IP // Source IPv6 address to use for outgoing connections, if the interface has one.
	Use        bool   // Flag indicating whether to use this network interface.
}

// LocalIPFor returns the local IP address that should be used to source a
// connection to target, or nil if this interface has no address of
// target's family (e.g. an IPv4-only interface asked to source an IPv6
// connection). Callers should treat a nil result as "this interface can't
// reach a target of this family" rather than silently falling back to an
// unrelated interface.
func (n NetworkInterface) LocalIPFor(target netip.Addr) net.IP {
	if target.Is4() || target.Is4In6() {
		return n.SourceIPv4
	}
	return n.SourceIPv6
}

// NewNetworkInterface uses the given source IP address or NIC name to find
// the IP address(es) to use as the source for probes and DNS lookups. The
// given IP address must exist on a NIC. When given an interface name that
// has both an IPv4 and an IPv6 address, both are captured (unless -4/-6
// restrict resolution to a single family), so probing can keep working
// correctly if the target's resolved address family changes mid-run (e.g.
// via hostname-retry-resolve).
func NewNetworkInterface(
	sourceAddress string,
	useIPv4,
	useIPv6 bool,
) (NetworkInterface, error) {
	interfaceAddress := net.ParseIP(sourceAddress)

	if interfaceAddress != nil { // we are given an IP address
		ifaceAddrs, err := net.InterfaceAddrs()
		if err != nil {
			return NetworkInterface{}, fmt.Errorf("unable to get interface IP addresses")
		}

		found := false
		for _, ifaceAddr := range ifaceAddrs {
			ipNet, ok := ifaceAddr.(*net.IPNet)
			if ok && interfaceAddress.Equal(ipNet.IP) {
				// we don't need to set anything here
				// just validating that the given IP belongs to an interface
				found = true
				break
			}
		}

		if !found {
			return NetworkInterface{}, fmt.Errorf("IP address %s is not assigned to any interfaces", sourceAddress)
		}

		ni := NetworkInterface{Use: true}
		if interfaceAddress.To4() != nil {
			ni.SourceIPv4 = interfaceAddress
		} else {
			ni.SourceIPv6 = interfaceAddress
		}
		return ni, nil
	}

	// we are probably given an interface name
	iface, err := net.InterfaceByName(sourceAddress)
	if err != nil {
		return NetworkInterface{}, fmt.Errorf("interface %s was not found", sourceAddress)
	}

	netAddrs, err := iface.Addrs()
	if err != nil {
		return NetworkInterface{}, fmt.Errorf("unable to get IP addresses of %s", iface.Name)
	}

	ni := NetworkInterface{Use: true}

	for _, netAddr := range netAddrs {
		ip, ok := netAddr.(*net.IPNet)
		if !ok || ip.IP == nil {
			continue
		}

		netIPAddr, err := netip.ParseAddr(ip.IP.String())
		if err != nil {
			continue
		}

		switch {
		case netIPAddr.Is4() && !useIPv6 && ni.SourceIPv4 == nil:
			ni.SourceIPv4 = ip.IP

		case netIPAddr.Is6() && !useIPv4 && ni.SourceIPv6 == nil:
			if netIPAddr.IsLinkLocalUnicast() {
				continue
			}
			ni.SourceIPv6 = ip.IP
		}
	}

	if ni.SourceIPv4 == nil && ni.SourceIPv6 == nil {
		return NetworkInterface{}, fmt.Errorf("unable to find an IP address associated with %s", iface.Name)
	}

	return ni, nil
}
