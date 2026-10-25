// Package cli turns the command line into the settings a run uses. It owns
// everything that only makes sense when tcping is driven from a terminal:
// the flags, the usage text, the target parsing and the update check.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/config"
	"github.com/pouriyajamshidi/tcping/v3/dns"
	"github.com/pouriyajamshidi/tcping/v3/nic"
	"github.com/pouriyajamshidi/tcping/v3/printers"
)

// defaultHTTPTimeout is how long an HTTP(S) probe waits when -t is not given.
const defaultHTTPTimeout = 3 * time.Second

// proxyFor parses the -proxy flag the probes go through. It returns nil
// when the probes connect directly. The proxy environment variables are left
// alone, so a proxy is only ever used when asked for.
func proxyFor(proxyFlag string, protocol config.Protocol) (*url.URL, error) {
	if proxyFlag == "" {
		return nil, nil
	}

	// A UDP probe cannot go through the proxy, and quietly probing directly
	// would hide that.
	if protocol == config.UDP {
		return nil, errors.New("--proxy only works with TCP and HTTP(S) targets")
	}

	proxy, err := url.Parse(proxyFlag)
	if err != nil || proxy.Host == "" {
		return nil, fmt.Errorf("invalid proxy %q, use a URL such as http://proxy.example.com:3128", proxyFlag)
	}

	switch proxy.Scheme {
	case "http", "https", "socks5", "socks5h":
		return proxy, nil
	default:
		return nil, fmt.Errorf("unsupported proxy %q, use an http://, https://, socks5:// or socks5h:// URL", proxyFlag)
	}
}

// proxyAddress looks the proxy's hostname up, so -4, -6 and --dns-server
// decide which address of the proxy the probes connect to. A proxy URL
// without a port gets the usual one for its scheme.
func proxyAddress(resolver *dns.Resolver, proxy *url.URL) (netip.AddrPort, error) {
	ip, err := resolver.ResolveHostname(proxy.Hostname())
	if err != nil {
		return netip.AddrPort{}, fmt.Errorf("could not resolve the proxy %s: %w", proxy.Hostname(), err)
	}

	port := proxy.Port()
	if port == "" {
		port = map[string]string{"http": "80", "https": "443", "socks5": "1080", "socks5h": "1080"}[proxy.Scheme]
	}

	portNumber, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return netip.AddrPort{}, fmt.Errorf("invalid proxy port %q", port)
	}

	return netip.AddrPortFrom(ip, uint16(portNumber)), nil
}

// ProcessUserInput turns the command line into the settings for a run:
// what to probe, and where the results go.
func ProcessUserInput() (config.Config, printers.Config) {
	f := registerFlags()

	flag.CommandLine.Usage = usage

	permuteArgs(os.Args[1:])

	flag.Parse()

	if f.showVer {
		showVersion()
	}

	if f.checkUpdates {
		checkForUpdates()
	}

	if f.completionsShell != "" {
		printCompletions(f.completionsShell)
	}

	f.validate()

	// The target says which protocol to speak: an http(s):// URL selects an
	// HTTP probe, anything else is a TCP one given as "host port" or
	// "host:port".
	target, err := parseTarget(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if target.hostname == "" || target.port == "" {
		fmt.Fprintln(os.Stderr, "At least the host and port or host:port format must be specified")
		usage()
	}

	validatedPort, err := convertAndValidatePort(target.port)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	proxy, err := proxyFor(f.proxy, target.protocol)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Resolved before the DNS resolver so hostname lookups can also be
	// bound to it (see dns.NewResolver).
	var networkInterface nic.NetworkInterface
	if f.interfaceName != "" {
		networkInterface, err = nic.NewNetworkInterface(
			f.interfaceName,
			f.useIPv4,
			f.useIPv6,
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	resolver, err := dns.NewResolver(
		f.customDNSServer,
		secondsToDuration(f.dnsTimeout),
		f.useIPv4,
		f.useIPv6,
		networkInterface,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Through a proxy, the proxy looks the target's hostname up, so we look
	// up the proxy's instead.
	var resolvedIP netip.Addr
	var nameResolutionDuration time.Duration
	var proxyAddr netip.AddrPort
	if proxy == nil {
		resolveStart := time.Now()
		resolvedIP, err = resolver.ResolveHostname(target.hostname)
		nameResolutionDuration = time.Since(resolveStart)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not resolve %s: %v\n", target.hostname, err)
			os.Exit(1)
		}
	} else {
		proxyAddr, err = proxyAddress(resolver, proxy)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	_, err = netip.ParseAddr(target.hostname)
	targetIsAlreadyIP := err == nil

	shouldRetryResolve := f.retryHostnameResolveAfterNFailures > 0 && !targetIsAlreadyIP

	protocol := target.protocol
	if f.udpServer {
		protocol = config.UDP
	}

	// An HTTP(S) probe has to connect, do the TLS handshake and wait for the
	// server to answer, so it gets a longer timeout unless -t is given.
	timeout := secondsToDuration(f.timeout)
	timeoutGiven := false
	flag.Visit(func(fl *flag.Flag) { timeoutGiven = timeoutGiven || fl.Name == "t" })
	if (protocol == config.HTTP || protocol == config.HTTPS) && !timeoutGiven {
		timeout = defaultHTTPTimeout
	}

	cfg := config.Config{
		Hostname:                   target.hostname,
		URL:                        target.url,
		IP:                         resolvedIP,
		Port:                       validatedPort,
		Protocol:                   protocol,
		Timeout:                    timeout,
		ProbesBeforeQuit:           f.probesBeforeQuit,
		TargetIsIP:                 targetIsAlreadyIP,
		NameResolutionDuration:     nameResolutionDuration,
		IntervalBetweenProbes:      secondsToDuration(f.intervalBetweenProbes),
		ShowFailuresOnly:           f.showFailuresOnly,
		SkipTLSVerify:              f.skipTLSVerify,
		Proxy:                      proxy,
		ProxyAddr:                  proxyAddr,
		UDPServer:                  f.udpServer,
		JSONServer:                 f.jsonServer,
		Resolver:                   resolver,
		ShouldRetryResolve:         shouldRetryResolve,
		ResolveEveryProbe:          f.resolveEveryProbe,
		RetryResolveAfterNFailures: f.retryHostnameResolveAfterNFailures,
		NetworkInterface:           networkInterface,
	}

	return cfg, f.newPrinterConfig(target.hostname, validatedPort)
}
