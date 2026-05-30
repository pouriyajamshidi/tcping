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
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/dns"
	"github.com/pouriyajamshidi/tcping/v3/internal/models"
	"github.com/pouriyajamshidi/tcping/v3/internal/printers"
	"github.com/pouriyajamshidi/tcping/v3/internal/utils"
)

// Config holds all user provided settings
type Config struct {
	Hostname                 string
	IP                       netip.Addr
	Port                     uint16
	UseIPv4                  *bool
	UseIPv6                  *bool
	showFailuresOnly         *bool
	showSourceAddress        *bool
	NonInteractive           *bool
	RetryResolveAfter        *uint
	probesBeforeQuit         *uint
	ifaceNameOrIPAddress     *string
	Timeout                  time.Duration
	IntervalBetweenProbes    time.Duration
	args                     []string
	PrinterConfig            printers.PrinterConfig
	ProbeOptions             models.ProbeOptions
	NetworkInterface         models.NetworkInterface
	RetryHostnameLookupAfter uint // Number of failed requests before retrying to resolve the hostname.
	ShouldRetryResolve       bool
	ShowFailuresOnly         bool
}

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

// permuteArgs rearranges user provided args for flag parsing,
// it stops just before the first non-flag argument.
// see: https://pkg.go.dev/flag
func permuteArgs(args []string) {
	var flagArgs []string
	var nonFlagArgs []string

	for i := range len(args) {
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
func ProcessUserInput() Config {
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

	retryHostnameResolveAfter := flag.Uint("r",
		0,
		"retry resolving target's hostname after <n> number of failed probes. e.g. -r 10 to retry after 10 failed probes.")

	showFailuresOnly := flag.Bool("show-failures-only", false, "Show only the failed probes.")

	showVer := flag.Bool("v", false, "show version and exit.")

	checkUpdates := flag.Bool("u", false, "check for updates and exit.")

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

	// At least the host and port must be specified
	if len(args) != 2 {
		utils.Usage()
	}

	if *useIPv4 && *useIPv6 {
		fmt.Println("Only one IP version can be specified")
		utils.Usage()
	}

	target := args[0]
	validatedPort := convertAndValidatePort(args[1])

	printerConfig := printers.PrinterConfig{
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

	intervalBetweenProbesConv := utils.SecondsToDuration(*intervalBetweenProbes)
	if intervalBetweenProbesConv < 2*time.Millisecond {
		// TODO: Do we keep this constraint?
		fmt.Println("Wait interval should be more than 2 ms")
		os.Exit(1)
	}

	var targetIsAlreadyIP bool
	var hostnameChanges []models.HostnameChange
	resolvedIP := dns.ResolveHostname2(target, *useIPv4, *useIPv6)
	if resolvedIP.String() == target {
		targetIsAlreadyIP = true
	} else {
		// track IP changes.
		hostnameChanges = []models.HostnameChange{
			{Addr: resolvedIP, When: time.Now()},
		}
	}

	var shouldRetryResolve bool
	if *retryHostnameResolveAfter > 0 && !targetIsAlreadyIP {
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
		IntervalBetweenProbes:    intervalBetweenProbesConv,
		Port:                     validatedPort,
		UseIPv4:                  *useIPv4,
		UseIPv6:                  *useIPv6,
		NonInteractive:           *nonInteractive,
		ShouldRetryResolve:       false,
		ShowFailuresOnly:         *showFailuresOnly,
		TargetIsIP:               targetIsAlreadyIP,
		HostnameChanges:          hostnameChanges,
	}

	return Config{
		IP:                    resolvedIP,
		Hostname:              target,
		Port:                  validatedPort,
		UseIPv4:               useIPv4,
		UseIPv6:               useIPv6,
		NonInteractive:        nonInteractive,
		RetryResolveAfter:     retryHostnameResolveAfter,
		probesBeforeQuit:      probesBeforeQuit,
		Timeout:               utils.SecondsToDuration(*timeout),
		IntervalBetweenProbes: intervalBetweenProbesConv,
		ifaceNameOrIPAddress:  interfaceName,
		showFailuresOnly:      showFailuresOnly,
		args:                  args,
		NetworkInterface:      networkInterface,
		ShowFailuresOnly:      *showFailuresOnly,
		ShouldRetryResolve:    shouldRetryResolve,
		PrinterConfig:         printerConfig,
		ProbeOptions:          probeOptions,
	}
}
