package nic

import (
	"fmt"
	"net"
	"net/netip"
)

// NetworkInterface represents a network interface used for connecting to the target.
type NetworkInterface struct {
	SourceIP net.IP     // Source IP address to use for outgoing connections, including DNS lookups.
	Dialer   net.Dialer // Dialer used to make network connections.
	Use      bool       // Flag indicating whether to use this network interface.
}

// NewNetworkInterface uses the given source IP address or NIC name (to find its first IP address)
// to use as the source IP address for the probes. The given IP address must exist on a NIC.
func NewNetworkInterface(
	sourceAddress string,
	useIPv4,
	useIPv6 bool,
) (NetworkInterface, error) {
	found := false
	interfaceAddress := net.ParseIP(sourceAddress)

	if interfaceAddress != nil { // we are given an IP address
		ifaceAddrs, err := net.InterfaceAddrs()
		if err != nil {
			return NetworkInterface{}, fmt.Errorf("unable to get interface IP addresses")
		}

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
	} else { // we are probably given an interface name
		iface, err := net.InterfaceByName(sourceAddress)
		if err != nil {
			return NetworkInterface{}, fmt.Errorf("interface %s was not found", sourceAddress)
		}

		netAddrs, err := iface.Addrs()
		if err != nil {
			return NetworkInterface{}, fmt.Errorf("unable to get IP addresses of %s", iface.Name)
		}

		for _, netAddr := range netAddrs {
			if ip := netAddr.(*net.IPNet).IP; ip != nil {
				netIPAddr, err := netip.ParseAddr(ip.String())
				if err != nil {
					continue
				}

				if netIPAddr.Is4() && !useIPv6 {
					interfaceAddress = ip
					found = true
					break
				} else if netIPAddr.Is6() && !useIPv4 {
					if netIPAddr.IsLinkLocalUnicast() {
						continue
					}
					interfaceAddress = ip
					found = true
					break
				}
			}
		}

		if interfaceAddress == nil {
			return NetworkInterface{}, fmt.Errorf("unable to find an IP address associated with %s", iface.Name)
		}
	}

	return NetworkInterface{
		Use:      true,
		SourceIP: interfaceAddress,
		Dialer: net.Dialer{
			LocalAddr: &net.TCPAddr{
				IP: interfaceAddress,
			},
		},
	}, nil
}
