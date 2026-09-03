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

// PrinterConfig holds all configuration options for Printer creation
type PrinterConfig struct {
	OutputJSON        bool
	PrettyJSON        bool
	NoColor           bool
	WithTimestamp     bool
	WithSourceAddress bool
	OmitStatistics    bool // Do not show the statistics. Only available for terminal printers
	Verbose           bool // Show everything an HTTP(S) probe learned, not just the status.
	OutputDBPath      string
	OutputCSVPath     string
	CSVNoTimestamp    bool // Omit the date/time suffix from CSV filenames, using OutputCSVPath as-is.

	Target string
	Port   uint16

	AlloyURL           string        // Address of a Grafana Alloy OTLP HTTP endpoint. Empty unless -alloy was given.
	AlloyStatsInterval time.Duration // How often the run summary is sent to Alloy.

	InfluxDBURL           string        // Address of an InfluxDB server. Empty unless -influxdb was given.
	InfluxDBOrg           string        // InfluxDB organization to write to.
	InfluxDBBucket        string        // InfluxDB bucket to write to.
	InfluxDBToken         string        // InfluxDB API token, taken from the INFLUXDB_TOKEN environment variable.
	InfluxDBStatsInterval time.Duration // How often the run summary is written to InfluxDB.

	SourceLabel string // Names the machine tcping runs on in the metrics sent to Alloy and InfluxDB. Defaults to the hostname.
}

// Config holds all user provided settings
type Config struct {
	URL                        string // Full target URL. HTTP(S) targets only, empty otherwise.
	Hostname                   string
	IP                         netip.Addr
	Port                       uint16
	Protocol                   consts.Protocol
	RetryResolveAfterNFailures uint
	ProbesBeforeQuit           uint
	Timeout                    time.Duration
	IntervalBetweenProbes      time.Duration
	PrinterConfig              PrinterConfig
	NetworkInterface           nic.NetworkInterface
	TargetIsIP                 bool          // Flag indicating whether the destination is an IP address (not a hostname).
	NameResolutionDuration     time.Duration // How long the initial hostname resolution took. Meaningless (and unset) when TargetIsIP.
	ShouldRetryResolve         bool
	ResolveEveryProbe          bool // Resolve the hostname before every probe, superseding ShouldRetryResolve.
	ShowFailuresOnly           bool
	SkipTLSVerify              bool // Do not check the server certificate. HTTPS targets only.
	UDPServer                  bool // Listen on the given address and echo datagrams back instead of probing.
	Resolver                   *dns.Resolver
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

	resolveEveryProbe := flag.Bool(
		"resolve-every-probe",
		false,
		`Resolve the target's hostname before every single probe instead of
		only at startup (and after failures, if -r is set). Has no effect
		when the target is already an IP address. Takes precedence over -r.`)

	customDNSServer := flag.String(
		"dns-server",
		"",
		`Custom DNS server IP to use. Defaults to the system-wide server.
		IP and port combination is allowed: 1.1.1.1:53`)

	dnsTimeout := flag.Float64(
		"dns-timeout",
		dns.DefaultTimeout.Seconds(),
		`Time to wait for a DNS response in seconds.
		Real number allowed.
		0 means infinite timeout.`)

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

	omitStatistics := flag.Bool(
		"omit-stats",
		false,
		`Do not show the statistics at program exit. Only applicable to terminal output.
		Pressing enter still shows the statistics.`)

	verbose := flag.Bool(
		"v",
		false,
		`Show all the details an HTTP(S) probe collects: the HTTP version,
		the TLS version and cipher, the certificate expiry and the
		connect/TLS/first-byte timings. For a UDP target, shows the number of
		the probe and whether the reply carried it back, so a lost probe can
		be told apart from the rest. No effect on other targets.`)

	skipTLSVerify := flag.Bool(
		"skip-tls",
		false,
		`Do not check the server certificate when probing an https:// target.
		Useful for self-signed or expired certificates.`)

	udpServer := flag.Bool(
		"udp-server",
		false,
		`Do not probe. Listen on the given host and port and echo every UDP
		datagram back to its sender, so a UDP probe pointed at this machine
		gets a reply and can tell that its packet arrived.`)

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

	csvNoTimestamp := flag.Bool(
		"csv-no-timestamp",
		false,
		`Do not append a date/time suffix to the CSV filename given via -csv.
		Repeated runs will then overwrite the same file instead of creating a new one.`)

	alloyURL := flag.String(
		"alloy",
		"",
		`Send the results to a Grafana Alloy OTLP HTTP endpoint as metrics
		instead of printing them, e.g. http://localhost:4318`)

	alloyStatsInterval := flag.Float64(
		"alloy-stats-interval",
		10,
		`How often to send the statistics to Alloy, in seconds.
		They are sent along with the next probe, so a longer probe interval
		delays them. No effect without the -alloy flag.`)

	influxDBURL := flag.String(
		"influxdb",
		"",
		`Send the results to an InfluxDB v2 or v3 server as metrics instead
		of printing them, e.g. http://localhost:8086
		The API token is read from the INFLUXDB_TOKEN environment variable.`)

	influxDBOrg := flag.String(
		"influxdb-org",
		"",
		`InfluxDB organization to write to. Required with the -influxdb flag.`)

	influxDBBucket := flag.String(
		"influxdb-bucket",
		"",
		`InfluxDB bucket to write to. Required with the -influxdb flag.`)

	influxDBStatsInterval := flag.Float64(
		"influxdb-stats-interval",
		10,
		`How often to write the statistics to InfluxDB, in seconds.
		They are written along with the next probe, so a longer probe
		interval delays them. No effect without the -influxdb flag.`)

	sourceLabel := flag.String(
		"source-label",
		"",
		`Name this machine in the metrics sent to Alloy or InfluxDB, so that
		several machines probing the same target can be told apart.
		Defaults to the machine's hostname. No effect without the -alloy or
		-influxdb flag.`)

	DBPath := flag.String(
		"db",
		"",
		"Path and file name to store the output in a sqlite3 database.")

	showVer := flag.Bool("version", false, "Show version and exit.")

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

	if *prettyJSON && !*outputJSON {
		fmt.Fprintln(os.Stderr, "--pretty has no effect without the -j flag")
		usage()
	}

	// The file and metric printers always write their final record, so
	// there are no statistics to omit.
	if *omitStatistics && (*DBPath != "" || *CSVPath != "" || *alloyURL != "" || *influxDBURL != "") {
		fmt.Fprintln(os.Stderr, "--omit-stats has no effect when the output goes to a file, a database or a metrics endpoint")
		usage()
	}

	args := flag.Args()

	// The target says which protocol to speak: an http(s):// URL selects an
	// HTTP probe, anything else is a TCP one given as "host port" or
	// "host:port".
	target, err := parseTarget(args)
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

	intervalBetweenProbesDuration := SecondsToDuration(*intervalBetweenProbes)
	if intervalBetweenProbesDuration < minProbeInterval {
		// TODO: Do we keep this constraint?
		fmt.Fprintln(os.Stderr, "Wait interval should be more than 2 ms")
		os.Exit(1)
	}

	alloyStatsIntervalDuration := SecondsToDuration(*alloyStatsInterval)
	if *alloyURL != "" && alloyStatsIntervalDuration <= 0 {
		fmt.Fprintln(os.Stderr, "Alloy statistics interval should be more than 0 seconds")
		os.Exit(1)
	}

	influxDBStatsIntervalDuration := SecondsToDuration(*influxDBStatsInterval)
	if *influxDBURL != "" && influxDBStatsIntervalDuration <= 0 {
		fmt.Fprintln(os.Stderr, "InfluxDB statistics interval should be more than 0 seconds")
		os.Exit(1)
	}

	// Resolved before the DNS resolver so hostname lookups can also be
	// bound to it (see createDNSResolver).
	var networkInterface nic.NetworkInterface
	if *interfaceName != "" {
		networkInterface, err = nic.NewNetworkInterface(
			*interfaceName,
			*useIPv4,
			*useIPv6,
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	dnsTimeoutDuration := SecondsToDuration(*dnsTimeout)

	resolver := dns.NewResolver(
		*customDNSServer,
		dnsTimeoutDuration,
		*useIPv4,
		*useIPv6,
		networkInterface,
	)

	var targetIsAlreadyIP bool

	resolveStart := time.Now()
	resolvedIP, err := resolver.ResolveHostname(target.hostname)
	nameResolutionDuration := time.Since(resolveStart)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not resolve %s: %v\n", target.hostname, err)
		os.Exit(1)
	}
	if resolvedIP.String() == target.hostname {
		targetIsAlreadyIP = true
	}

	var shouldRetryResolve bool
	if *retryHostnameResolveAfterNFailures > 0 && !targetIsAlreadyIP {
		shouldRetryResolve = true
	}

	timeoutInDuration := SecondsToDuration(*timeout)

	// Without this, several machines probing the same target would all
	// write to the same series and their numbers would be mixed together.
	srclabel := *sourceLabel
	if srclabel == "" {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}
		srclabel = hostname
	}

	printerConfig := PrinterConfig{
		Target:             target.hostname,
		Port:               validatedPort,
		OutputJSON:         *outputJSON,
		PrettyJSON:         *prettyJSON,
		NoColor:            *noColor,
		WithTimestamp:      *showTimestamp,
		WithSourceAddress:  *showSourceAddress,
		OmitStatistics:     *omitStatistics,
		Verbose:            *verbose,
		OutputDBPath:       *DBPath,
		OutputCSVPath:      *CSVPath,
		CSVNoTimestamp:     *csvNoTimestamp,
		AlloyURL:           *alloyURL,
		AlloyStatsInterval: alloyStatsIntervalDuration,

		InfluxDBURL:    *influxDBURL,
		InfluxDBOrg:    *influxDBOrg,
		InfluxDBBucket: *influxDBBucket,
		// The token is kept out of the flags so it never ends up in the
		// shell history.
		InfluxDBToken:         os.Getenv("INFLUXDB_TOKEN"),
		InfluxDBStatsInterval: influxDBStatsIntervalDuration,

		SourceLabel: srclabel,
	}

	protocol := target.protocol
	if *udpServer {
		protocol = consts.UDP
	}

	return Config{
		Hostname:                   target.hostname,
		URL:                        target.url,
		IP:                         resolvedIP,
		Port:                       validatedPort,
		Protocol:                   protocol,
		Timeout:                    timeoutInDuration,
		ProbesBeforeQuit:           *probesBeforeQuit,
		TargetIsIP:                 targetIsAlreadyIP,
		NameResolutionDuration:     nameResolutionDuration,
		IntervalBetweenProbes:      intervalBetweenProbesDuration,
		ShowFailuresOnly:           *showFailuresOnly,
		SkipTLSVerify:              *skipTLSVerify,
		UDPServer:                  *udpServer,
		Resolver:                   resolver,
		ShouldRetryResolve:         shouldRetryResolve,
		ResolveEveryProbe:          *resolveEveryProbe,
		RetryResolveAfterNFailures: *retryHostnameResolveAfterNFailures,
		NetworkInterface:           networkInterface,
		PrinterConfig:              printerConfig,
	}
}
