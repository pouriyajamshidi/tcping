// Package cli turns the command line into the settings a run uses. It owns
// everything that only makes sense when tcping is driven from a terminal:
// the flags, the usage text, the target parsing and the update check.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/config"

	"github.com/pouriyajamshidi/tcping/v3/dns"
	"github.com/pouriyajamshidi/tcping/v3/nic"
	"github.com/pouriyajamshidi/tcping/v3/printers"
)

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

	proxy, err := proxyFor(f.proxy, target.url)
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

	resolveStart := time.Now()
	resolvedIP, err := resolver.ResolveHostname(target.hostname)
	nameResolutionDuration := time.Since(resolveStart)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not resolve %s: %v\n", target.hostname, err)
		os.Exit(1)
	}

	_, err = netip.ParseAddr(target.hostname)
	targetIsAlreadyIP := err == nil

	shouldRetryResolve := f.retryHostnameResolveAfterNFailures > 0 && !targetIsAlreadyIP

	protocol := target.protocol
	if f.udpServer {
		protocol = config.UDP
	}

	cfg := config.Config{
		Hostname:                   target.hostname,
		URL:                        target.url,
		IP:                         resolvedIP,
		Port:                       validatedPort,
		Protocol:                   protocol,
		Timeout:                    secondsToDuration(f.timeout),
		ProbesBeforeQuit:           f.probesBeforeQuit,
		TargetIsIP:                 targetIsAlreadyIP,
		NameResolutionDuration:     nameResolutionDuration,
		IntervalBetweenProbes:      secondsToDuration(f.intervalBetweenProbes),
		ShowFailuresOnly:           f.showFailuresOnly,
		SkipTLSVerify:              f.skipTLSVerify,
		Proxy:                      proxy,
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

// proxyFor picks the proxy an HTTP(S) target is probed through: the -proxy
// flag when it is given, the HTTPS_PROXY, HTTP_PROXY and NO_PROXY environment
// variables otherwise. It returns nil when the probes connect directly.
func proxyFor(proxyFlag, targetURL string) (*url.URL, error) {
	// Only an HTTP(S) target has a URL. A TCP or UDP probe cannot go through
	// the proxy, and quietly probing directly would hide that.
	if targetURL == "" {
		if proxyFlag != "" {
			return nil, errors.New("--proxy only works with http:// and https:// targets")
		}
		return nil, nil
	}

	if proxyFlag == "" {
		req, err := http.NewRequest(http.MethodGet, targetURL, nil)
		if err != nil {
			return nil, err
		}

		proxy, err := http.ProxyFromEnvironment(req)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy in the environment: %w", err)
		}
		return proxy, nil
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
