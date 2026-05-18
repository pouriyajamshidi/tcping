package main

import (
	"context"
	"flag"
	"net"
	"net/netip"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v45/github"
)

type userInput struct {
	ip                       netip.Addr
	hostname                 string
	networkInterface         networkInterface
	retryHostnameLookupAfter uint // Retry resolving target's hostname after a certain number of failed requests
	probesBeforeQuit         uint
	timeout                  time.Duration
	intervalBetweenProbes    time.Duration
	port                     uint16
	useIPv4                  bool
	useIPv6                  bool
	shouldRetryResolve       bool
	showFailuresOnly         bool
	showSourceAddress        bool
	nonInteractive           bool // Enable the program to run in the background. e.g. nohup, disown
}

type genericUserInputArgs struct {
	retryResolve         *uint
	probesBeforeQuit     *uint
	timeout              *float64
	secondsBetweenProbes *float64
	intName              *string
	showFailuresOnly     *bool
	showSourceAddress    *bool
	nonInteractive       *bool
	args                 []string
}

type networkInterface struct {
	remoteAddr *net.TCPAddr
	dialer     net.Dialer
	use        bool
}

// usage prints how tcping should be run
func usage() {
	executableName := os.Args[0]

	colorLightCyan("\nTCPING version %s\n\n", version)
	colorRed("Try running %s like:\n", executableName)
	colorRed("%s <hostname/ip> <port number>. For example:\n", executableName)
	colorRed("%s www.example.com 443\n", executableName)
	colorRed("Or use the <hostname/ip:port> format:\n")
	colorRed("%s www.example.com:443\n", executableName)
	colorYellow("\n[optional flags]\n")

	flag.VisitAll(func(f *flag.Flag) {
		flagName := f.Name
		if len(f.Name) > 1 {
			flagName = "-" + flagName
		}

		colorYellow("  -%s : %s\n", flagName, f.Usage)
	})

	os.Exit(1)
}

// setPrinter selects the printer
func setPrinter(tcping *tcping, outputJSON, prettyJSON *bool, noColor *bool, timeStamp *bool, sourceAddress *bool, outputDb *string, outputCSV *string, args []string) {
	if *prettyJSON && !*outputJSON {
		colorRed("--pretty has no effect without the -j flag.")
		usage()
	}

	if *outputJSON {
		tcping.printer = newJSONPrinter(*prettyJSON)
	} else if *outputDb != "" {
		tcping.printer = newDB(*outputDb, args)
	} else if *outputCSV != "" {
		var err error
		tcping.printer, err = newCSVPrinter(*outputCSV, timeStamp, sourceAddress)
		if err != nil {
			tcping.printError("Failed to create CSV file: %s", err)
			os.Exit(1)
		}
	} else if *noColor {
		tcping.printer = newPlainPrinter(timeStamp)
	} else {
		tcping.printer = newColorPrinter(timeStamp)
	}
}

// showVersion displays the version and exits
func showVersion(tcping *tcping) {
	tcping.printVersion()
	os.Exit(0)
}

// setIPFlags ensures that either IPv4 or IPv6 is specified by the user and not both and sets it
func setIPFlags(tcping *tcping, ip4, ip6 *bool) {
	if *ip4 && *ip6 {
		tcping.printError("Only one IP version can be specified")
		usage()
	}
	if *ip4 {
		tcping.userInput.useIPv4 = true
	}
	if *ip6 {
		tcping.userInput.useIPv6 = true
	}
}

// setPort validates and sets the TCP/UDP port range
func setPort(tcping *tcping, args []string) {
	port, err := strconv.ParseUint(args[1], 10, 16)
	if err != nil {
		tcping.printError("Invalid port number: %s", args[1])
		os.Exit(1)
	}

	if port < 1 || port > 65535 {
		tcping.printError("Port should be in 1..65535 range")
		os.Exit(1)
	}
	tcping.userInput.port = uint16(port)
}

// parseHostPortArgs handles both "host port" and "host:port" formats
// It returns a slice with exactly 2 elements [host, port] if successful
func parseHostPortArgs(args []string) []string {
	if len(args) == 1 {
		// Check if the single argument is in "host:port" format
		parts := strings.Split(args[0], ":")
		if len(parts) == 2 {
			// Valid "host:port" format
			return parts
		} else if len(parts) > 2 {
			// Could be IPv6 address with port like [::1]:8080 or ::1:8080
			// Try to find the last colon as port separator
			lastColonIndex := strings.LastIndex(args[0], ":")
			if lastColonIndex > 0 {
				host := args[0][:lastColonIndex]
				port := args[0][lastColonIndex+1:]
				// Remove brackets if present for IPv6
				host = strings.TrimPrefix(host, "[")
				host = strings.TrimSuffix(host, "]")
				return []string{host, port}
			}
		}
	}
	return args
}

// setGenericArgs assigns the generic flags after sanity checks
func setGenericArgs(tcping *tcping, genericArgs genericUserInputArgs) {
	if *genericArgs.retryResolve > 0 {
		tcping.userInput.retryHostnameLookupAfter = *genericArgs.retryResolve
	}

	tcping.userInput.hostname = genericArgs.args[0]
	tcping.userInput.ip = resolveHostname(tcping)
	tcping.startTime = time.Now()
	tcping.userInput.probesBeforeQuit = *genericArgs.probesBeforeQuit
	tcping.userInput.timeout = secondsToDuration(*genericArgs.timeout)

	tcping.userInput.intervalBetweenProbes = secondsToDuration(*genericArgs.secondsBetweenProbes)
	if tcping.userInput.intervalBetweenProbes < 2*time.Millisecond {
		tcping.printError("Wait interval should be more than 2 ms")
		os.Exit(1)
	}

	// this serves as a default starting value for tracking IP changes.
	tcping.hostnameChanges = []hostnameChange{
		{tcping.userInput.ip, time.Now()},
	}

	if tcping.userInput.hostname == tcping.userInput.ip.String() {
		tcping.destIsIP = true
	}

	if tcping.userInput.retryHostnameLookupAfter > 0 && !tcping.destIsIP {
		tcping.userInput.shouldRetryResolve = true
	}

	if *genericArgs.intName != "" {
		tcping.userInput.networkInterface = newNetworkInterface(tcping, *genericArgs.intName)
	}

	tcping.userInput.showFailuresOnly = *genericArgs.showFailuresOnly

	tcping.userInput.showSourceAddress = *genericArgs.showSourceAddress
	tcping.userInput.nonInteractive = *genericArgs.nonInteractive
}

// processUserInput gets and validate user input
func processUserInput(tcping *tcping) {
	useIPv4 := flag.Bool("4", false, "only use IPv4.")
	useIPv6 := flag.Bool("6", false, "only use IPv6.")
	retryHostnameResolveAfter := flag.Uint("r", 0, "retry resolving target's hostname after <n> number of failed probes. e.g. -r 10 to retry after 10 failed probes.")
	probesBeforeQuit := flag.Uint("c", 0, "stop after <n> probes, regardless of the result. By default, no limit will be applied.")
	outputJSON := flag.Bool("j", false, "output in JSON format.")
	prettyJSON := flag.Bool("pretty", false, "use indentation when using json output format. No effect without the '-j' flag.")
	nonInteractive := flag.Bool("non-interactive", false, "let tcping run in the background, for instance using nohup or disown")
	noColor := flag.Bool("no-color", false, "do not colorize output.")
	showTimestamp := flag.Bool("D", false, "show timestamp in output.")
	saveToCSV := flag.String("csv", "", "path and file name to store tcping output to CSV file...If user prompts for stats, it will be saved to a file with the same name and _stats appended.")
	showVer := flag.Bool("v", false, "show version.")
	checkUpdates := flag.Bool("u", false, "check for updates and exit.")
	secondsBetweenProbes := flag.Float64("i", 1, "interval between sending probes. Real number allowed with dot as a decimal separator. The default is one second")
	timeout := flag.Float64("t", 1, "time to wait for a response, in seconds. Real number allowed. 0 means infinite timeout.")
	outputDB := flag.String("db", "", "path and file name to store tcping output to sqlite database.")
	interfaceName := flag.String("I", "", "interface name or address.")
	showSourceAddress := flag.Bool("show-source-address", false, "Show source address and port used for probes.")
	showFailuresOnly := flag.Bool("show-failures-only", false, "Show only the failed probes.")
	showHelp := flag.Bool("h", false, "show help message.")

	flag.CommandLine.Usage = usage

	permuteArgs(os.Args[1:])
	flag.Parse()

	// validation for flag and args
	args := flag.Args()

	// we need to set printers first, because they're used for
	// error reporting and other output.
	setPrinter(tcping, outputJSON, prettyJSON, noColor, showTimestamp, showSourceAddress, outputDB, saveToCSV, args)

	// Handle -v flag
	if *showVer {
		showVersion(tcping)
	}

	// Handle -h flag
	if *showHelp {
		usage()
	}

	// Handle -u flag
	if *checkUpdates {
		checkForUpdates(tcping)
	}

	// host and port must be specified
	// Support both "host port" and "host:port" formats
	args = parseHostPortArgs(args)
	if len(args) != 2 {
		usage()
	}

	// Check whether both the ipv4 and ipv6 flags are attempted set if ony one, error otherwise.
	setIPFlags(tcping, useIPv4, useIPv6)

	// Check if the port is valid and set it.
	setPort(tcping, args)

	// set generic args
	genericArgs := genericUserInputArgs{
		retryResolve:         retryHostnameResolveAfter,
		probesBeforeQuit:     probesBeforeQuit,
		timeout:              timeout,
		secondsBetweenProbes: secondsBetweenProbes,
		intName:              interfaceName,
		showFailuresOnly:     showFailuresOnly,
		showSourceAddress:    showSourceAddress,
		nonInteractive:       nonInteractive,
		args:                 args,
	}

	setGenericArgs(tcping, genericArgs)
}

/*
permuteArgs permute args for flag parsing stops just before the first non-flag argument.

see: https://pkg.go.dev/flag
*/
func permuteArgs(args []string) {
	var flagArgs []string
	var nonFlagArgs []string

	for i := 0; i < len(args); i++ {
		v := args[i]
		if v[0] == '-' {
			var optionName string
			if v[1] == '-' {
				optionName = v[2:]
			} else {
				optionName = v[1:]
			}
			switch optionName {
			case "c":
				fallthrough
			case "t":
				fallthrough
			case "db":
				fallthrough
			case "I":
				fallthrough
			case "i":
				fallthrough
			case "csv":
				fallthrough
			case "r":
				/* out of index */
				if len(args) <= i+1 {
					usage()
				}
				/* the next flag has come */
				optionVal := args[i+1]
				if optionVal[0] == '-' {
					usage()
				}
				flagArgs = append(flagArgs, args[i:i+2]...)
				i++
			default:
				flagArgs = append(flagArgs, args[i])
			}
		} else {
			nonFlagArgs = append(nonFlagArgs, args[i])
		}
	}
	permutedArgs := append(flagArgs, nonFlagArgs...)

	/* replace args */
	for i := range args {
		args[i] = permutedArgs[i]
	}
}

// newNetworkInterface uses the 1st ip address of the interface
// if any err occurs it calls `tcpStats.printError` and exits with status code 1.
// or return `networkInterface`
func newNetworkInterface(tcping *tcping, netInterface string) networkInterface {
	var interfaceAddress net.IP

	interfaceAddress = net.ParseIP(netInterface)

	if interfaceAddress == nil {
		ief, err := net.InterfaceByName(netInterface)
		if err != nil {
			tcping.printError("Interface %s not found", netInterface)
			os.Exit(1)
		}

		addrs, err := ief.Addrs()
		if err != nil {
			tcping.printError("Unable to get Interface addresses")
			os.Exit(1)
		}

		// Iterating through the available addresses to identify valid IP configurations
		for _, addr := range addrs {
			if ip := addr.(*net.IPNet).IP; ip != nil {
				// netip.Addr
				nipAddr, err := netip.ParseAddr(ip.String())
				if err != nil {
					continue
				}

				if nipAddr.Is4() && !tcping.userInput.useIPv6 {
					interfaceAddress = ip
					break
				} else if nipAddr.Is6() && !tcping.userInput.useIPv4 {
					if nipAddr.IsLinkLocalUnicast() {
						continue
					}
					interfaceAddress = ip
					break
				}
			}
		}

		if interfaceAddress == nil {
			tcping.printError("Unable to get Interface's IP Address")
			os.Exit(1)
		}
	}

	// Initializing a networkInterface struct and setting the 'use' field to true
	ni := networkInterface{
		use: true,
	}

	ni.remoteAddr = &net.TCPAddr{
		IP:   net.ParseIP(tcping.userInput.ip.String()),
		Port: int(tcping.userInput.port),
	}

	sourceAddr := &net.TCPAddr{
		IP: interfaceAddress,
	}

	ni.dialer = net.Dialer{
		LocalAddr: sourceAddr,
		Timeout:   tcping.userInput.timeout, // Set the timeout duration
	}

	return ni
}

// compareVersions is used to compare tcping versions
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		n1, _ := strconv.Atoi(parts1[i])
		n2, _ := strconv.Atoi(parts2[i])

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	// for cases in which version numbers differ in length
	if len(parts1) < len(parts2) {
		return -1
	}

	if len(parts1) > len(parts2) {
		return 1
	}

	return 0
}

// checkForUpdates checks for newer versions of tcping
func checkForUpdates(tcping *tcping) {
	c := github.NewClient(nil)

	/* unauthenticated requests from the same IP are limited to 60 per hour. */
	latestRelease, _, err := c.Repositories.GetLatestRelease(context.Background(), owner, repo)
	if err != nil {
		tcping.printError("Failed to check for updates %s", err.Error())
		os.Exit(1)
	}

	reg := `^v?(\d+\.\d+\.\d+)$`
	latestTagName := latestRelease.GetTagName()
	latestVersion := regexp.MustCompile(reg).FindStringSubmatch(latestTagName)

	if len(latestVersion) == 0 {
		tcping.printError("Failed to check for updates. The version name does not match the rule: %s", latestTagName)
		os.Exit(1)
	}

	comparison := compareVersions(version, latestVersion[1])

	if comparison < 0 {
		tcping.printInfo("Found newer version %s", latestVersion[1])
		tcping.printInfo("Please update TCPING from the URL below:")
		tcping.printInfo("https://github.com/%s/%s/releases/tag/%s",
			owner, repo, latestTagName)
	} else if comparison > 0 {
		tcping.printInfo("Current version %s is newer than the latest release %s",
			version, latestVersion[1])
	} else {
		tcping.printInfo("You have the latest version: %s", version)
	}

	os.Exit(0)
}
