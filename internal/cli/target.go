package cli

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/pouriyajamshidi/tcping/v3/config"
)

// convertAndValidatePort validates and returns the TCP/UDP port
func convertAndValidatePort(port string) (uint16, error) {
	parsedPort, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid port number %q", port)
	}

	if parsedPort == 0 {
		return 0, fmt.Errorf("port should be in 1..65535 range")
	}

	return uint16(parsedPort), nil
}

// parseHostPortArgs handles both "host port" and "host:port" formats
func parseHostPortArgs(args []string) (host string, port string) {
	if len(args) == 1 {
		// We have `host:port`
		if h, p, err := net.SplitHostPort(args[0]); err == nil {
			return h, p
		}
		return args[0], ""
	}

	// We were given `host port`
	return args[0], args[1]
}

// probeTarget is what the user asked us to probe, extracted from the
// positional arguments.
type probeTarget struct {
	hostname string
	port     string
	protocol config.Protocol
	url      string // Full URL, HTTP(S) targets only.
}

// parseTarget picks the protocol from the user input: an
// http:// or https:// prefix means an HTTP probe, a udp:// one means a UDP
// probe, anything else is the plain TCP probe tcping has always done. HTTP
// gets a default port (80 or 443) but it can still be overridden by the
// URL's own port or a trailing port argument. UDP has no default port.
func parseTarget(args []string) (probeTarget, error) {
	if len(args) == 0 {
		return probeTarget{}, nil
	}

	protocol, isURL := schemeProtocol(args[0])
	if !isURL {
		host, port := parseHostPortArgs(args)
		return probeTarget{hostname: host, port: port, protocol: config.TCP}, nil
	}

	u, err := url.Parse(args[0])
	if err != nil || u.Hostname() == "" {
		return probeTarget{}, fmt.Errorf("invalid URL %q", args[0])
	}

	target := probeTarget{
		hostname: u.Hostname(),
		port:     u.Port(),
		protocol: protocol,
	}

	// A trailing port argument wins over the one embedded in the URL.
	if len(args) > 1 {
		target.port = args[1]
	}

	// UDP has no URL to request, just a host and a port to send a datagram
	// to, so there is nothing left to build here.
	if protocol == config.UDP {
		return target, nil
	}

	if target.port == "" {
		target.port = defaultPort(protocol)
	}

	// Rebuild the host so the Host header matches the port we will dial,
	// while still omitting the port when it is the scheme's default - some
	// virtual hosts only match the bare name.
	if target.port == defaultPort(protocol) {
		u.Host = target.hostname
	} else {
		u.Host = net.JoinHostPort(target.hostname, target.port)
	}

	target.url = u.String()

	return target, nil
}

// schemeProtocol reports the protocol implied by target's URL scheme, and
// whether it had one we handle at all.
func schemeProtocol(target string) (config.Protocol, bool) {
	switch {
	case strings.HasPrefix(target, "http://"):
		return config.HTTP, true
	case strings.HasPrefix(target, "https://"):
		return config.HTTPS, true
	case strings.HasPrefix(target, "udp://"):
		return config.UDP, true
	default:
		return "", false
	}
}

func defaultPort(protocol config.Protocol) string {
	if protocol == config.HTTPS {
		return "443"
	}
	return "80"
}
