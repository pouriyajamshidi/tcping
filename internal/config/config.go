// Package config handles user input and the defaults to run tcping
package config

import (
	"flag"
	"fmt"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/consts"
	"github.com/pouriyajamshidi/tcping/v3/internal/dns"
	"github.com/pouriyajamshidi/tcping/v3/internal/nic"
	"github.com/pouriyajamshidi/tcping/v3/internal/printers"
	"github.com/pouriyajamshidi/tcping/v3/internal/utils"
)

const minProbeInterval = 2 * time.Millisecond

// flagsRequiringValue inspects every flag registered on flag.CommandLine and
// returns the set of flag names that expect a value on the command line
// (i.e. anything that isn't a bool flag). Derived directly from the flag
// definitions, so it can never drift out of sync when a flag is added,
// renamed, or removed.
func flagsRequiringValue() map[string]bool {
	flagsWithValues := make(map[string]bool)

	flag.VisitAll(func(f *flag.Flag) {
		// Flags created via flag.Bool implement this interface, it's the
		// same check the flag package uses internally to decide whether a
		// flag needs a following argument.
		if bf, ok := f.Value.(interface{ IsBoolFlag() bool }); ok && bf.IsBoolFlag() {
			return
		}

		flagsWithValues[f.Name] = true
	})

	return flagsWithValues
}

// permuteArgs rearranges user provided args for flag parsing.
// Go's standard flag package stops parsing flags as soon as it encounters the first non-flag argument,
// causing us to lose our flags, so we need to override that behavior.
// Given [tcping.go example.com 443 -4 -r 1 -c 2] or [tcping.go -4 -r 1 -c 2 example.com 443]
// returns [tcping.go -4 -r 1 -c 2 example.com 443]
// nonFlagArgs are ["example.com", "443"]
// flagArgs are ["-4", "-c 2"]
//
// Without permutation, `tcping example.com 443 -4 -c 5` becomes:
// flags:
//
//	none
//
// args:
//
//	example.com
//	443
//	-4
//	-c
//	5
//
// With permutation:
// flags:
//
//	-4
//	-c 5
//
// args:
//
//	example.com
//	443
//
// In memory of Takaya, you will be missed my friend.
func permuteArgs(args []string) {
	flagsWithValues := flagsRequiringValue()

	flagArgs := make([]string, 0, len(args))
	nonFlagArgs := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if !strings.HasPrefix(arg, "-") {
			nonFlagArgs = append(nonFlagArgs, arg)
			continue
		}

		option := strings.TrimLeft(arg, "-")

		if flagsWithValues[option] {
			if i+1 >= len(args) {
				fmt.Printf("-%s option requires a value\n", option)
				usage()
			}

			if strings.HasPrefix(args[i+1], "-") {
				fmt.Printf("-%s option requires a value\n", option)
				usage()
			}

			flagArgs = append(flagArgs, arg, args[i+1])

			i++
			continue
		}

		// bool flags
		flagArgs = append(flagArgs, arg)
	}

	copy(args, append(flagArgs, nonFlagArgs...))
}

// Config holds all user provided settings
type Config struct {
	Hostname                   string
	IP                         netip.Addr
	Port                       uint16
	Protocol                   consts.Protocol
	UseIPv4                    bool
	UseIPv6                    bool
	ShowSourceAddress          bool
	RetryResolveAfterNFailures uint
	ProbesBeforeQuit           uint
	IfaceNameOrIPAddress       string
	Timeout                    time.Duration
	IntervalBetweenProbes      time.Duration
	PrinterConfig              printers.PrinterConfig
	NetworkInterface           nic.NetworkInterface
	RetryHostnameLookupAfter   uint // Number of failed requests before retrying to resolve the hostname.
	TargetIsIP                 bool // Flag indicating whether the destination is an IP address (not a hostname).
	ShouldRetryResolve         bool
	ShowFailuresOnly           bool
	Resolver                   *dns.Resolver
}

func (c Config) GetHostname() string {
	return c.Hostname
}
func (c Config) GetIP() netip.Addr {
	return c.IP
}
func (c Config) GetPort() uint16 {
	return c.Port
}
func (c Config) GetProtocol() consts.Protocol {
	return c.Protocol
}
func (c Config) GetUseIPv4() bool {
	return c.UseIPv4
}
func (c Config) GetUseIPv6() bool {
	return c.UseIPv6
}
func (c Config) GetTimeout() string {
	return c.Timeout.String()
}
func (c Config) GetProbesBeforeQuit() uint {
	return c.ProbesBeforeQuit
}
func (c Config) GetTargetIsIP() bool {
	return c.TargetIsIP
}
func (c Config) GetIntervalBetweenProbes() string {
	return c.IntervalBetweenProbes.String()
}
func (c Config) GetShowFailuresOnly() bool {
	return c.ShowFailuresOnly
}
func (c Config) GetShouldRetryResolve() bool {
	return c.ShouldRetryResolve
}
func (c Config) GetRetryResolveAfterNFailures() uint {
	return c.RetryHostnameLookupAfter
}
func (c Config) GetNetworkInterface() nic.NetworkInterface {
	return c.NetworkInterface
}
func (c Config) GetPrinterConfig() printers.PrinterConfig {
	return c.PrinterConfig
}
func (c Config) GetWithTimestamp() bool {
	return c.PrinterConfig.WithTimestamp
}
func (c Config) GetWithSourceAddress() bool {
	return c.PrinterConfig.WithSourceAddress
}

// ProcessUserInput gets and validate user input
func ProcessUserInput() Config {
	useIPv4 := flag.Bool("4", false, "Only use IPv4 to initiate probes.")

	useIPv6 := flag.Bool("6", false, "Only use IPv6 to initiate probes.")

	probesBeforeQuit := flag.Uint(
		"c",
		0,
		`Stop after <n> probes, regardless of the result.
		By default, no limit will be applied.`)

	intervalBetweenProbes := flag.Float64(
		"i",
		1,
		`Interval between probes.
		Real number allowed with dot as a decimal separator.
		The default value is one second`)

	timeout := flag.Float64(
		"t",
		1,
		`Time to wait for a response in seconds.
		Real number allowed.
		0 means infinite timeout.`)

	showTimestamp := flag.Bool(
		"D",
		false,
		"Show a timestamp for each probe in the output.")

	retryHostnameResolveAfterNFailures := flag.Uint(
		"r",
		0,
		`Retry resolving target's hostname after <n> number of failed probes.
		e.g. -r 10 to retry after 10 failed probes.`)

	customDNSServer := flag.String(
		"dns-server",
		"",
		`Custom DNS server IP to use. Defaults to the system-wide server.
		IP and port combination is allowed: 1.1.1.1:53`)

	interfaceName := flag.String(
		"I",
		"",
		"Use a specific interface name or IP address to initiate the probes.")

	showSourceAddress := flag.Bool(
		"show-source-address",
		false,
		"Show source address and port used for probes.")

	showFailuresOnly := flag.Bool(
		"show-failures-only",
		false,
		"Show only the failed probes.")

	noColor := flag.Bool("no-color", false, "Do not colorize output.")

	outputJSON := flag.Bool(
		"j",
		false,
		"Output in JSON format.")

	prettyJSON := flag.Bool(
		"pretty",
		false,
		`Prettify the JSON output.
		No effect without the '-j' flag.`)

	CSVPath := flag.String(
		"csv",
		"",
		`Path and file name to store the output in a CSV file.
		The stats will be automatically saved with the same name and '_stats' suffix.`)

	DBPath := flag.String(
		"db",
		"",
		"Path and file name to store the output in a sqlite3 database.")

	showVer := flag.Bool("v", false, "Show version and exit.")

	checkUpdates := flag.Bool("u", false, "Check for updates and exit.")

	flag.CommandLine.Usage = usage

	permuteArgs(os.Args[1:])

	flag.Parse()

	if *showVer {
		showVersion()
	}

	if *checkUpdates {
		checkForUpdates()
	}

	if *useIPv4 && *useIPv6 {
		fmt.Fprintln(os.Stderr, "Only one IP version can be specified")
		usage()
	}

	args := flag.Args()

	// host and port must be specified
	// Support both "host port" and "host:port" formats
	target, port := parseHostPortArgs(args)

	if target == "" || port == "" {
		fmt.Fprintln(os.Stderr, "At least the host and port or host:port format must be specified")
		usage()
	}

	validatedPort, err := convertAndValidatePort(port)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	intervalBetweenProbesDuration := utils.SecondsToDuration(*intervalBetweenProbes)
	if intervalBetweenProbesDuration < minProbeInterval {
		// TODO: Do we keep this constraint?
		fmt.Fprintln(os.Stderr, "Wait interval should be more than 2 ms")
		os.Exit(1)
	}

	resolver := dns.NewResolver(
		*customDNSServer,
		2*time.Second, // TODO: make this configurable
		*useIPv4,
		*useIPv6,
	)

	var targetIsAlreadyIP bool

	resolvedIP, err := resolver.ResolveHostname(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not resolve %s: %v\n", target, err)
		os.Exit(1)
	}
	if resolvedIP.String() == target {
		targetIsAlreadyIP = true
	}

	var shouldRetryResolve bool
	if *retryHostnameResolveAfterNFailures > 0 && !targetIsAlreadyIP {
		shouldRetryResolve = true
	}

	timeoutInDuration := utils.SecondsToDuration(*timeout)

	// TODO: double check
	var networkInterface nic.NetworkInterface
	if *interfaceName != "" {
		networkInterface, err = nic.NewNetworkInterface(
			*interfaceName,
			resolvedIP,
			validatedPort,
			*useIPv4,
			*useIPv6,
			timeoutInDuration,
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	printerConfig := printers.PrinterConfig{
		Target:            target,
		Port:              validatedPort,
		OutputJSON:        *outputJSON,
		PrettyJSON:        *prettyJSON,
		NoColor:           *noColor,
		WithTimestamp:     *showTimestamp,
		WithSourceAddress: *showSourceAddress,
		OutputDBPath:      *DBPath,
		OutputCSVPath:     *CSVPath,
	}

	return Config{
		Hostname:                   target,
		IP:                         resolvedIP,
		Port:                       validatedPort,
		UseIPv4:                    *useIPv4,
		UseIPv6:                    *useIPv6,
		Timeout:                    timeoutInDuration,
		ProbesBeforeQuit:           *probesBeforeQuit,
		TargetIsIP:                 targetIsAlreadyIP,
		IntervalBetweenProbes:      intervalBetweenProbesDuration,
		ShowFailuresOnly:           *showFailuresOnly,
		Resolver:                   resolver,
		ShouldRetryResolve:         shouldRetryResolve,
		RetryResolveAfterNFailures: *retryHostnameResolveAfterNFailures,
		IfaceNameOrIPAddress:       *interfaceName,
		NetworkInterface:           networkInterface,
		PrinterConfig:              printerConfig,
	}
}
