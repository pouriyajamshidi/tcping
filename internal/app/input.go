package app

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/netip"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/pouriyajamshidi/tcping/v3"
	"github.com/pouriyajamshidi/tcping/v3/statistics"
)

var (
	// ErrUsageRequested indicates usage help was requested
	ErrUsageRequested = errors.New("usage requested")

	// ErrVersionRequested indicates version display was requested
	ErrVersionRequested = errors.New("version requested")

	// ErrUpdateCheckRequested indicates update check was requested
	ErrUpdateCheckRequested = errors.New("update check requested")
)

// ProberConfig contains all configuration needed to create and run a prober.
type ProberConfig struct {
	// Target configuration
	Hostname string
	Port     uint16

	// Network options
	UseIPv4          bool
	UseIPv6          bool
	InterfaceName    string
	InterfaceDialer  *net.Dialer
	ShowSourceAddress bool

	// Timing options
	Timeout  time.Duration
	Interval time.Duration

	// Probe control
	ProbeCountLimit uint
	ShowFailuresOnly bool

	// DNS options
	RetryResolveAfter uint

	// Output options
	PrinterConfig tcping.PrinterConfig

	// Runtime options
	NonInteractive bool
}

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
func newNetworkInterface(ipAddress string, useIPv4, useIPv6 bool) (*net.Dialer, error) {
	interfaceAddress := net.ParseIP(ipAddress)
	isInvalid := true

	if interfaceAddress != nil {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return nil, fmt.Errorf("get ip addresses: %w", err)
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
			return nil, fmt.Errorf("interface %s not found: %w", ipAddress, err)
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return nil, fmt.Errorf("get interface addresses: %w", err)
		}

		for _, addr := range addrs {
			if ip := addr.(*net.IPNet).IP; ip != nil {
				nipAddr, err := netip.ParseAddr(ip.String())
				if err != nil {
					continue
				}

				if nipAddr.Is4() && !useIPv6 {
					interfaceAddress = ip
					isInvalid = false
					break
				} else if nipAddr.Is6() && !useIPv4 {
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
			return nil, fmt.Errorf("get interface ip address")
		}
	}

	if isInvalid {
		return nil, fmt.Errorf("ip address %s not assigned to any interface", ipAddress)
	}

	return &net.Dialer{
		LocalAddr: &net.TCPAddr{
			IP: interfaceAddress,
		},
	}, nil
}

// setOptions assigns the user provided flags after sanity checks
func setOptions(config *ProberConfig, opts options) error {
	config.RetryResolveAfter = *opts.retryResolve

	if *opts.useIPv4 {
		config.UseIPv4 = true
	} else if *opts.useIPv6 {
		config.UseIPv6 = true
	}

	config.Hostname = opts.args[0]
	config.Port = convertAndValidatePort(opts.args[1])
	config.ProbeCountLimit = *opts.probesBeforeQuit
	config.Timeout = statistics.SecondsToDuration(*opts.timeout)
	config.NonInteractive = *opts.nonInteractive
	config.Interval = statistics.SecondsToDuration(*opts.intervalBetweenProbes)

	if config.Interval < 2*time.Millisecond {
		return fmt.Errorf("wait interval should be more than 2 ms")
	}

	if *opts.intName != "" {
		dialer, err := newNetworkInterface(*opts.intName, config.UseIPv4, config.UseIPv6)
		if err != nil {
			return fmt.Errorf("setup network interface: %w", err)
		}
		config.InterfaceDialer = dialer
		config.InterfaceName = *opts.intName
	}

	config.ShowFailuresOnly = *opts.showFailuresOnly
	return nil
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

// permuteArgs permute args for flag parsing stops just before the first non-flag argument.
// see: https://pkg.go.dev/flag
func permuteArgs(args []string) error {
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
				// out of index
				if len(args) <= i+1 {
					return ErrUsageRequested
				}
				// the next flag has come
				optionVal := args[i+1]
				if optionVal[0] == '-' {
					return ErrUsageRequested
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

	return nil
}

// ProcessUserInput parses command-line flags. Returns ErrUsageRequested,
// ErrVersionRequested, or ErrUpdateCheckRequested for special control flow.
func ProcessUserInput() (ProberConfig, error) {
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

	flag.CommandLine.Usage = func() {
		// no-op, we'll handle usage in app package
	}

	if err := permuteArgs(os.Args[1:]); err != nil {
		return ProberConfig{}, err
	}

	flag.Parse()

	args := flag.Args()

	if *showVer {
		return ProberConfig{}, ErrVersionRequested
	}

	if *checkUpdates {
		return ProberConfig{}, ErrUpdateCheckRequested
	}

	if len(args) != 2 {
		return ProberConfig{}, ErrUsageRequested
	}

	if *useIPv4 && *useIPv6 {
		return ProberConfig{}, fmt.Errorf("%w: only one IP version can be specified", ErrUsageRequested)
	}

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

	config := ProberConfig{
		ShowSourceAddress: *showSourceAddress,
		PrinterConfig: tcping.PrinterConfig{
			OutputJSON:        *outputJSON,
			PrettyJSON:        *prettyJSON,
			NoColor:           *noColor,
			WithTimestamp:     *showTimestamp,
			WithSourceAddress: *showSourceAddress,
			OutputDBPath:      *saveToDB,
			OutputCSVPath:     *saveToCSV,
			Target:            args[0],
			Port:              args[1],
		},
	}

	if err := setOptions(&config, opts); err != nil {
		return ProberConfig{}, fmt.Errorf("set options: %w", err)
	}

	return config, nil
}
