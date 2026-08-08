// Package config handles user input and the defaults to run tcping
package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/dns"
	"github.com/pouriyajamshidi/tcping/v3/internal/models"
	"github.com/pouriyajamshidi/tcping/v3/internal/nic"
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

// ProcessUserInput gets and validate user input
func ProcessUserInput() models.Config {
	useIPv4 := flag.Bool("4", false, "Only use IPv4 to initiate probes.")

	useIPv6 := flag.Bool("6", false, "Only use IPv6 to initiate probes.")

	probesBeforeQuit := flag.Uint(
		"c",
		0,
		`Stop after <n> probes, regardless of the result.
		By default, no limit will be applied.`)

	nonInteractive := flag.Bool(
		"non-interactive",
		false,
		`Let tcping run in detach or background mode.
		For instance by using nohup or disown`)

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
		"Custom DNS server IP to use. Defaults to the system-wide server")

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

	if *showVer {
		showVersion()
	}

	if *checkUpdates {
		checkForUpdates()
	}

	if *useIPv4 && *useIPv6 {
		fmt.Println("Only one IP version can be specified")
		usage()
	}

	permuteArgs(os.Args[1:])

	flag.Parse()

	args := flag.Args()

	// host and port must be specified
	// Support both "host port" and "host:port" formats
	target, port := parseHostPortArgs(args)

	if target == "" || port == "" {
		fmt.Println("At least the host and port or host:port format must be specified")
		usage()
	}

	validatedPort, err := convertAndValidatePort(port)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	intervalBetweenProbesDuration := utils.SecondsToDuration(*intervalBetweenProbes)
	if intervalBetweenProbesDuration < minProbeInterval {
		// TODO: Do we keep this constraint?
		fmt.Println("Wait interval should be more than 2 ms")
		os.Exit(1)
	}

	DNSResolver := dns.NewDNSResolver(*customDNSServer)

	var targetIsAlreadyIP bool
	var hostnameChanges []models.HostnameChange

	resolvedIP, err := DNSResolver.ResolveHostname(target, *useIPv4, *useIPv6)
	if err != nil {
		fmt.Printf("Could not resolve %s\n", target)
		os.Exit(1)
	}
	if resolvedIP.String() == target {
		targetIsAlreadyIP = true
	} else {
		// track IP changes.
		hostnameChanges = []models.HostnameChange{
			{Addr: resolvedIP, When: time.Now()},
		}
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
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

	printerConfig := models.PrinterConfig{
		Target:            target,
		Port:              validatedPort,
		OutputJSON:        *outputJSON,
		PrettyJSON:        *prettyJSON,
		NoColor:           *noColor,
		WithTimestamp:     *showTimestamp,
		WithSourceAddress: *showSourceAddress,
		OutputDBPath:      *saveToDB,
		OutputCSVPath:     *saveToCSV,
	}

	probeOptions := models.ProbeOptions{
		IP:                         resolvedIP,
		Hostname:                   target,
		NetworkInterface:           networkInterface,
		RetryResolveAfterNFailures: *retryHostnameResolveAfterNFailures,
		ShouldRetryResolve:         shouldRetryResolve,
		ProbesBeforeQuit:           *probesBeforeQuit,
		Timeout:                    timeoutInDuration,
		IntervalBetweenProbes:      intervalBetweenProbesDuration,
		Port:                       validatedPort,
		UseIPv4:                    *useIPv4,
		UseIPv6:                    *useIPv6,
		NonInteractive:             *nonInteractive,
		ShowFailuresOnly:           *showFailuresOnly,
		TargetIsIP:                 targetIsAlreadyIP,
	}

	// TODO: Remove the duplicates from `probeOptions` and make field associations logical
	return models.Config{
		IP:                         resolvedIP,
		Hostname:                   target,
		Port:                       validatedPort,
		UseIPv4:                    *useIPv4,
		UseIPv6:                    *useIPv6,
		NonInteractive:             *nonInteractive,
		RetryResolveAfterNFailures: *retryHostnameResolveAfterNFailures,
		ProbesBeforeQuit:           *probesBeforeQuit,
		Timeout:                    timeoutInDuration,
		IntervalBetweenProbes:      intervalBetweenProbesDuration,
		IfaceNameOrIPAddress:       *interfaceName,
		ShowFailuresOnly:           *showFailuresOnly,
		Args:                       args,
		NetworkInterface:           networkInterface,
		ShouldRetryResolve:         shouldRetryResolve,
		PrinterConfig:              printerConfig,
		ProbeOptions:               probeOptions,
		DNSResolver:                DNSResolver,
		HostnameChanges:            hostnameChanges,
	}
}
