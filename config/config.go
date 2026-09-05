// Package config holds the settings a tcping run uses. It is types only:
// the command line that fills them in lives in internal/cli.
package config

import (
	"net/netip"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/dns"
	"github.com/pouriyajamshidi/tcping/v3/nic"
)

// Config holds all user provided settings
type Config struct {
	URL                        string // Full target URL. HTTP(S) targets only, empty otherwise.
	Hostname                   string
	IP                         netip.Addr
	Port                       uint16
	Protocol                   Protocol
	RetryResolveAfterNFailures uint
	ProbesBeforeQuit           uint
	Timeout                    time.Duration
	IntervalBetweenProbes      time.Duration
	NetworkInterface           nic.NetworkInterface
	TargetIsIP                 bool          // Flag indicating whether the destination is an IP address (not a hostname).
	NameResolutionDuration     time.Duration // How long the initial hostname resolution took. Meaningless (and unset) when TargetIsIP.
	ShouldRetryResolve         bool
	ResolveEveryProbe          bool // Resolve the hostname before every probe, superseding ShouldRetryResolve.
	ShowFailuresOnly           bool
	SkipTLSVerify              bool // Do not check the server certificate. HTTPS targets only.
	UDPServer                  bool // Listen on the given address and echo datagrams back instead of probing.
	Resolver                   *dns.Resolver
}

// Protocol is the protocol a probe speaks, picked from the target the user
// gave us.
type Protocol string

// The protocols tcping can probe with.
const (
	TCP   Protocol = "TCP"
	UDP   Protocol = "UDP"
	HTTP  Protocol = "HTTP"
	HTTPS Protocol = "HTTPS"
	ICMP  Protocol = "ICMP"
)
