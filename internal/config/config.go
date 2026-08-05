// Package config handles user input and the defaults to run tcping
package config

import (
	"flag"
	"fmt"
	"net"
	"net/netip"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/dns"
	"github.com/pouriyajamshidi/tcping/v3/internal/models"
	"github.com/pouriyajamshidi/tcping/v3/internal/utils"
)

// newNetworkInterface uses the given source IP address or NIC name (to find its first IP address)
// to use as the source IP address for the probes. The given IP address must exist on a NIC.
func newNetworkInterface(
	sourceAddress,
	target string,
	port uint16,
	useIPv4,
	useIPv6 bool,
	timeout time.Duration,
) models.NetworkInterface {
	interfaceAddress := net.ParseIP(sourceAddress)
	isInvalid := true

	if interfaceAddress != nil { // we are given an IP address
		ifaceAddrs, err := net.InterfaceAddrs()
		if err != nil {
			fmt.Println("Unable to get interface IP addresses")
			os.Exit(1)
		}

		for _, ifaceAddr := range ifaceAddrs {
			ipNet, ok := ifaceAddr.(*net.IPNet)
			if ok && interfaceAddress.Equal(ipNet.IP) {
				// we don't need to set anything here
				// just validating that the given IP belongs to an interface
				isInvalid = false
				break
			}
		}

		if isInvalid {
			fmt.Printf("IP address %s is not assigned to any interfaces\n", sourceAddress)
			os.Exit(1)
		}
	} else { // we are probably given an interface name
		iface, err := net.InterfaceByName(sourceAddress)
		if err != nil {
			fmt.Printf("Interface %s was not found\n", sourceAddress)
			os.Exit(1)
		}

		netAddrs, err := iface.Addrs()
		if err != nil {
			fmt.Printf("Unable to get IP addresses of %s", iface.Name)
			os.Exit(1)
		}

		for _, netAddr := range netAddrs {
			if ip := netAddr.(*net.IPNet).IP; ip != nil {
				netIPAddr, err := netip.ParseAddr(ip.String())
				if err != nil {
					continue
				}

				if netIPAddr.Is4() && !useIPv6 {
					interfaceAddress = ip
					isInvalid = false
					break
				} else if netIPAddr.Is6() && !useIPv4 {
					if netIPAddr.IsLinkLocalUnicast() {
						continue
					}
					interfaceAddress = ip
					isInvalid = false
					break
				}
			}
		}

		if interfaceAddress == nil {
			fmt.Printf("Unable to find an IP address associated with %s", iface.Name)
			os.Exit(1)
		}
	}

	netIface := models.NetworkInterface{
		Use: true,
	}

	netIface.RemoteAddr = &net.TCPAddr{
		IP:   net.ParseIP(target),
		Port: int(port),
	}

	netIface.Dialer = net.Dialer{
		LocalAddr: &net.TCPAddr{
			IP: interfaceAddress,
		},
		Timeout: timeout, // Set the timeout duration
	}

	return netIface
}

// convertAndValidatePort validates and returns the TCP/UDP port
func convertAndValidatePort(portStr string) uint16 {
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		fmt.Printf("Invalid port number: %s\n", portStr)
		os.Exit(1)
	}

	if port < 1 || port > 65535 {
		fmt.Println("Port should be in 1..65535 range")
		os.Exit(1)
	}

	return uint16(port)
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

// permuteArgs rearranges user provided args for flag parsing,
// it stops just before the first non-flag argument.
// see: https://pkg.go.dev/flag
func permuteArgs(args []string) {
	var flagArgs []string
	var nonFlagArgs []string

	// we cannot use the newer `for i := range len(args)` syntax here
	// since we are mutating i, otherwise:
	// `tcping example.com 443 -4` works but `tcping -4 cats.com 443` doens't
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
				// out of index
				if len(args) <= i+1 {
					utils.Usage()
				}
				// the next flag has come
				optionVal := args[i+1]
				if optionVal[0] == '-' {
					utils.Usage()
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

	permutedArgs := slices.Concat(flagArgs, nonFlagArgs)

	// replace args in place
	for i := range len(args) {
		args[i] = permutedArgs[i]
	}
}

// ProcessUserInput gets and validate user input
func ProcessUserInput() models.Config {
	useIPv4 := flag.Bool("4", false, "only use IPv4 to initiate probes.")
	useIPv6 := flag.Bool("6", false, "only use IPv6 to initiate probes.")

	probesBeforeQuit := flag.Uint("c",
		0,
		"stop after <n> probes, regardless of the result. By default, no limit will be applied.")

	showTimestamp := flag.Bool("D", false, "show timestamp for each probe in the output.")

	outputJSON := flag.Bool("j", false, "output in JSON format.")
	prettyJSON := flag.Bool("pretty",
		false,
		"use indentation when using json output format. No effect without the '-j' flag.")

	nonInteractive := flag.Bool("non-interactive",
		false,
		"let tcping run in the background, for instance using nohup or disown")

	noColor := flag.Bool("no-color", false, "do not colorize output.")

	saveToCSV := flag.String("csv",
		"",
		"path and file name to store output to a CSV file. The stats will be saved with the same name and `_stats` suffix.")

	saveToDB := flag.String("db", "", "path and file name to store output to a sqlite3 database.")

	intervalBetweenProbes := flag.Float64("i",
		1,
		"interval between sending probes. Real number allowed with dot as a decimal separator. The default is one second")

	timeout := flag.Float64("t",
		1,
		"time to wait for a response, in seconds. Real number allowed. 0 means infinite timeout.")

	interfaceName := flag.String("I",
		"",
		"use a specific interface name or IP address to initiate probes.")

	showSourceAddress := flag.Bool("show-source-address", false, "Show source address and port used for probes.")

	retryHostnameResolveAfterNFailures := flag.Uint("r",
		0,
		"retry resolving target's hostname after <n> number of failed probes. e.g. -r 10 to retry after 10 failed probes.")

	showFailuresOnly := flag.Bool("show-failures-only", false, "Show only the failed probes.")

	showVer := flag.Bool("v", false, "show version and exit.")

	checkUpdates := flag.Bool("u", false, "check for updates and exit.")

	customDNSServer := flag.String("dns-server", "", "Custom DNS server IP to use. Defaults to system-wide server")

	flag.CommandLine.Usage = utils.Usage

	permuteArgs(os.Args[1:])

	flag.Parse()

	args := flag.Args()

	if *showVer {
		utils.ShowVersion()
	}

	if *checkUpdates {
		utils.CheckForUpdates()
	}

	if *useIPv4 && *useIPv6 {
		fmt.Println("Only one IP version can be specified")
		utils.Usage()
	}

	// host and port must be specified
	// Support both "host port" and "host:port" formats
	args = parseHostPortArgs(args)

	// At least the host and port or host:port format must be specified
	if len(args) != 2 && len(args) == 1 && !strings.Contains(args[0], ":") {
		utils.Usage()
	}

	target := args[0]
	validatedPort := convertAndValidatePort(args[1])

	printerConfig := models.PrinterConfig{
		OutputJSON:        *outputJSON,
		PrettyJSON:        *prettyJSON,
		NoColor:           *noColor,
		WithTimestamp:     *showTimestamp,
		WithSourceAddress: *showSourceAddress,
		OutputDBPath:      *saveToDB,
		OutputCSVPath:     *saveToCSV,
		Target:            target,
		Port:              validatedPort,
	}

	intervalBetweenProbesDuration := utils.SecondsToDuration(*intervalBetweenProbes)
	if intervalBetweenProbesDuration < 2*time.Millisecond {
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

	// TODO: double check
	var networkInterface models.NetworkInterface
	if *interfaceName != "" {
		networkInterface = newNetworkInterface(
			*interfaceName,
			target,
			validatedPort,
			*useIPv4,
			*useIPv6,
			utils.SecondsToDuration(*timeout),
		)
	}

	probeOptions := models.ProbeOptions{
		IP:                       resolvedIP,
		Hostname:                 target,
		NetworkInterface:         networkInterface,
		RetryHostnameLookupAfter: 0,
		ProbesBeforeQuit:         *probesBeforeQuit,
		Timeout:                  utils.SecondsToDuration(*timeout),
		IntervalBetweenProbes:    intervalBetweenProbesDuration,
		Port:                     validatedPort,
		UseIPv4:                  *useIPv4,
		UseIPv6:                  *useIPv6,
		NonInteractive:           *nonInteractive,
		ShouldRetryResolve:       false,
		ShowFailuresOnly:         *showFailuresOnly,
		TargetIsIP:               targetIsAlreadyIP,
	}

	// TODO: Remove the duplicates from `probeOptions` and make field associations logical
	return models.Config{
		IP:                         resolvedIP,
		Hostname:                   target,
		Port:                       validatedPort,
		UseIPv4:                    useIPv4,
		UseIPv6:                    useIPv6,
		NonInteractive:             nonInteractive,
		RetryResolveAfterNFailures: retryHostnameResolveAfterNFailures,
		ProbesBeforeQuit:           probesBeforeQuit,
		Timeout:                    utils.SecondsToDuration(*timeout),
		IntervalBetweenProbes:      intervalBetweenProbesDuration,
		IfaceNameOrIPAddress:       interfaceName,
		ShowFailuresOnly:           showFailuresOnly,
		Args:                       args,
		NetworkInterface:           networkInterface,
		ShouldRetryResolve:         shouldRetryResolve,
		PrinterConfig:              printerConfig,
		ProbeOptions:               probeOptions,
		DNSResolver:                DNSResolver,
		HostnameChanges:            hostnameChanges,
	}
}
