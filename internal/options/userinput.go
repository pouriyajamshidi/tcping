// Package options handles the user input
package options

import (
	"flag"
	"fmt"
	"net"
	"net/netip"
	"os"
	"strconv"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/internal/consts"
	"github.com/pouriyajamshidi/tcping/v2/internal/dns"
	"github.com/pouriyajamshidi/tcping/v2/internal/utils"
	"github.com/pouriyajamshidi/tcping/v2/printers"
	"github.com/pouriyajamshidi/tcping/v2/types"
)

type options struct {
	useIPv4               *bool
	useIPv6               *bool
	showFailuresOnly      *bool
	showSourceAddress     *bool
	nonInteractive        *bool
	retryResolve          *uint
	probesBeforeQuit      *uint
	intName               *string
	timeout               *float64
	intervalBetweenProbes *float64
	args                  []string
}

// newNetworkInterface uses the given IP address or a NIC to find the first IP address
// to use as the source of the probes. The given IP address must exist on the system.
func newNetworkInterface(tcping *types.Tcping, ipAddress string) types.NetworkInterface {
	interfaceAddress := net.ParseIP(ipAddress)
	isInvalid := true

	if interfaceAddress != nil {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			tcping.PrintError("Unable to get IP addresses")
			os.Exit(1)
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if ok && interfaceAddress.Equal(ipNet.IP) {
				isInvalid = false
				break
			}
		}
	} else { // we are probably given an interface name
		iface, err := net.InterfaceByName(ipAddress)
		if err != nil {
			tcping.PrintError("Interface %s not found", ipAddress)
			os.Exit(1)
		}

		addrs, err := iface.Addrs()
		if err != nil {
			tcping.PrintError("Unable to get Interface addresses")
			os.Exit(1)
		}

		for _, addr := range addrs {
			if ip := addr.(*net.IPNet).IP; ip != nil {
				nipAddr, err := netip.ParseAddr(ip.String())
				if err != nil {
					continue
				}

				if nipAddr.Is4() && !tcping.Options.UseIPv6 {
					interfaceAddress = ip
					isInvalid = false
					break
				} else if nipAddr.Is6() && !tcping.Options.UseIPv4 {
					if nipAddr.IsLinkLocalUnicast() {
						continue
					}
					interfaceAddress = ip
					isInvalid = false
					break
				}
			}
		}

		if interfaceAddress == nil {
			tcping.PrintError("Unable to get interface's IP address")
			os.Exit(1)
		}
	}

	if isInvalid {
		tcping.PrintError("IP address %s is not assigned to any interface", ipAddress)
		os.Exit(1)
	}

	netIface := types.NetworkInterface{
		Use: true,
	}

	netIface.RemoteAddr = &net.TCPAddr{
		IP:   net.ParseIP(tcping.Options.IP.String()),
		Port: int(tcping.Options.Port),
	}

	netIface.Dialer = net.Dialer{
		LocalAddr: &net.TCPAddr{
			IP: interfaceAddress,
		},
		Timeout: tcping.Options.Timeout, // Set the timeout duration
	}

	return netIface
}

// setOptions assigns the user provided flags after sanity checks
func setOptions(t *types.Tcping, opts options) {
	if *opts.retryResolve > 0 {
		t.Options.RetryHostnameLookupAfter = *opts.retryResolve
	}

	if *opts.useIPv4 {
		t.Options.UseIPv4 = true
	} else if *opts.useIPv6 {
		t.Options.UseIPv6 = true
	}

	t.Options.Hostname = opts.args[0]
	t.Options.Port = convertAndValidatePort(t, opts.args[1])
	t.Options.IP = dns.ResolveHostname(t)
	t.Options.ProbesBeforeQuit = *opts.probesBeforeQuit
	t.Options.Timeout = utils.SecondsToDuration(*opts.timeout)

	t.Options.NonInteractive = *opts.nonInteractive

	t.Options.IntervalBetweenProbes = utils.SecondsToDuration(*opts.intervalBetweenProbes)
	if t.Options.IntervalBetweenProbes < 2*time.Millisecond {
		t.PrintError("Wait interval should be more than 2 ms")
		os.Exit(1)
	}

	if t.Options.Hostname == t.Options.IP.String() {
		t.DestIsIP = true
	} else {
		// The default starting value for tracking IP changes.
		t.HostnameChanges = []types.HostnameChange{
			{Addr: t.Options.IP, When: time.Now()},
		}
	}

	if t.Options.RetryHostnameLookupAfter > 0 && !t.DestIsIP {
		t.Options.ShouldRetryResolve = true
	}

	if *opts.intName != "" {
		t.Options.NetworkInterface = newNetworkInterface(t, *opts.intName)
	}

	t.Options.ShowFailuresOnly = *opts.showFailuresOnly
}

// convertAndValidatePort validates and returns the TCP/UDP port
func convertAndValidatePort(t *types.Tcping, portStr string) uint16 {
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		t.PrintError("Invalid port number: %s", portStr)
		os.Exit(1)
	}

	if port < 1 || port > 65535 {
		t.PrintError("Port should be in 1..65535 range")
		os.Exit(1)
	}

	return uint16(port)
}

// permuteArgs permute args for flag parsing stops just before the first non-flag argument.
// see: https://pkg.go.dev/flag
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
					utils.Usage()
				}
				/* the next flag has come */
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
	permutedArgs := append(flagArgs, nonFlagArgs...)

	/* replace args */
	for i := range len(args) {
		args[i] = permutedArgs[i]
	}
}

// ProcessUserInput gets and validate user input
func ProcessUserInput(tcping *types.Tcping) {
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
		"Enforce using a specific interface name or IP address to initiate probes.")
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
		consts.ColorRed("Only one IP version can be specified")
		utils.Usage()
	}

	config := printers.PrinterConfig{
		OutputJSON:        *outputJSON,
		PrettyJSON:        *prettyJSON,
		NoColor:           *noColor,
		WithTimestamp:     *showTimestamp,
		WithSourceAddress: *showSourceAddress,
		OutputDBPath:      *saveToDB,
		OutputCSVPath:     *saveToCSV,
		Target:            args[0],
		Port:              args[1],
	}

	printer, err := printers.NewPrinter(config)
	if err != nil {
		fmt.Printf("Failed to create printer: %s\n", err)
		os.Exit(1)
	}

	tcping.Printer = printer

	opts := options{
		useIPv4:               useIPv4,
		useIPv6:               useIPv6,
		nonInteractive:        nonInteractive,
		retryResolve:          retryHostnameResolveAfter,
		probesBeforeQuit:      probesBeforeQuit,
		timeout:               timeout,
		intervalBetweenProbes: intervalBetweenProbes,
		intName:               interfaceName,
		showFailuresOnly:      showFailuresOnly,
		args:                  args,
	}

	setOptions(tcping, opts)
}
