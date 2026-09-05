// Package cli turns the command line into the settings a run uses. It owns
// everything that only makes sense when tcping is driven from a terminal:
// the flags, the usage text, the target parsing and the update check.
package cli

import (
	"flag"
	"fmt"
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

	resolver := dns.NewResolver(
		f.customDNSServer,
		secondsToDuration(f.dnsTimeout),
		f.useIPv4,
		f.useIPv6,
		networkInterface,
	)

	resolveStart := time.Now()
	resolvedIP, err := resolver.ResolveHostname(target.hostname)
	nameResolutionDuration := time.Since(resolveStart)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not resolve %s: %v\n", target.hostname, err)
		os.Exit(1)
	}

	targetIsAlreadyIP := resolvedIP.String() == target.hostname

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
		UDPServer:                  f.udpServer,
		Resolver:                   resolver,
		ShouldRetryResolve:         shouldRetryResolve,
		ResolveEveryProbe:          f.resolveEveryProbe,
		RetryResolveAfterNFailures: f.retryHostnameResolveAfterNFailures,
		NetworkInterface:           networkInterface,
	}

	return cfg, f.newPrinterConfig(target.hostname, validatedPort)
}
