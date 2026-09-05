package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/dns"
	"github.com/pouriyajamshidi/tcping/v3/printers"
)

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

// permuteArgs moves the flags in args ahead of everything else, so that
// `tcping example.com 443 -4 -c 5` works as well as `tcping -4 -c 5
// example.com 443`. Go's flag package stops looking for flags at the first
// argument that is not one, so without this the -4 and the -c would be
// ignored.
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

// flags holds the raw values flag.Parse fills in, before they are turned
// into a config.Config and a printers.Config.
type flags struct {
	useIPv4 bool
	useIPv6 bool

	probesBeforeQuit                   uint
	intervalBetweenProbes              float64
	timeout                            float64
	retryHostnameResolveAfterNFailures uint
	resolveEveryProbe                  bool

	customDNSServer string
	dnsTimeout      float64
	interfaceName   string

	showTimestamp     bool
	showSourceAddress bool
	showFailuresOnly  bool
	omitStatistics    bool
	verbose           bool
	noColor           bool

	skipTLSVerify bool
	udpServer     bool

	outputJSON bool
	prettyJSON bool

	CSVPath        string
	csvNoTimestamp bool
	DBPath         string

	alloyURL string

	influxDBURL    string
	influxDBOrg    string
	influxDBBucket string
	influxDBToken  string

	statsInterval float64

	sourceLabel string

	showVer      bool
	checkUpdates bool
}

// registerFlags declares every command line flag and returns the struct that
// flag.Parse will fill in.
func registerFlags() *flags {
	var f flags

	flag.BoolVar(&f.useIPv4, "4", false, "Only use IPv4 to initiate probes.")

	flag.BoolVar(&f.useIPv6, "6", false, "Only use IPv6 to initiate probes.")

	flag.UintVar(&f.probesBeforeQuit,
		"c",
		0,
		`Stop after <n> probes, regardless of the result.
		By default, no limit will be applied.`)

	flag.Float64Var(&f.intervalBetweenProbes,
		"i",
		1,
		`Interval between probes.
		Real number allowed with dot as a decimal separator.
		The default value is one second`)

	flag.Float64Var(&f.timeout,
		"t",
		1,
		`Time to wait for a response in seconds.
		Real number allowed.
		0 means infinite timeout.`)

	flag.BoolVar(&f.showTimestamp,
		"D",
		false,
		"Show a timestamp for each probe in the output.")

	flag.UintVar(&f.retryHostnameResolveAfterNFailures,
		"r",
		0,
		`Retry resolving target's hostname after <n> number of failed probes.
		e.g. -r 10 to retry after 10 failed probes.`)

	flag.BoolVar(&f.resolveEveryProbe,
		"resolve-every-probe",
		false,
		`Resolve the target's hostname before every single probe instead of
		only at startup (and after failures, if -r is set). Has no effect
		when the target is already an IP address. Takes precedence over -r.`)

	flag.StringVar(&f.customDNSServer,
		"dns-server",
		"",
		`Custom DNS server IP to use. Defaults to the system-wide server.
		IP and port combination is allowed: 1.1.1.1:53`)

	flag.Float64Var(&f.dnsTimeout,
		"dns-timeout",
		dns.DefaultTimeout.Seconds(),
		`Time to wait for a DNS response in seconds.
		Real number allowed.
		0 means infinite timeout.`)

	flag.StringVar(&f.interfaceName,
		"I",
		"",
		"Use a specific interface name or IP address to initiate the probes.")

	flag.BoolVar(&f.showSourceAddress,
		"show-source-address",
		false,
		"Show source address and port used for probes.")

	flag.BoolVar(&f.showFailuresOnly,
		"failures-only",
		false,
		"Show only the failed probes.")

	flag.BoolVar(&f.omitStatistics,
		"no-stats",
		false,
		`Do not show the statistics at program exit. Only applicable to terminal output.
		Pressing enter still shows the statistics.`)

	flag.BoolVar(&f.verbose,
		"v",
		false,
		`Show all the details an HTTP(S) probe collects: the HTTP version,
		the TLS version and cipher, the certificate expiry and the
		connect/TLS/first-byte timings. For a UDP target, shows the number of
		the probe and whether the reply carried it back, so a lost probe can
		be told apart from the rest. No effect on other targets.`)

	flag.BoolVar(&f.skipTLSVerify,
		"insecure",
		false,
		`Do not check the server certificate when probing an https:// target.
		Useful for self-signed or expired certificates.`)

	flag.BoolVar(&f.udpServer,
		"udp-server",
		false,
		`Do not probe. Listen on the given host and port and echo every UDP
		datagram back to its sender, so a UDP probe pointed at this machine
		gets a reply and can tell that its packet arrived.`)

	flag.BoolVar(&f.noColor, "no-color", false, "Do not colorize output.")

	flag.BoolVar(&f.outputJSON,
		"j",
		false,
		"Output in JSON format.")

	flag.BoolVar(&f.prettyJSON,
		"pretty",
		false,
		`Prettify the JSON output.
		No effect without the '-j' flag.`)

	flag.StringVar(&f.CSVPath,
		"csv",
		"",
		`Path and file name to store the output in a CSV file.
		The stats will be automatically saved with the same name and '_stats' suffix.`)

	flag.BoolVar(&f.csvNoTimestamp,
		"csv-fixed-name",
		false,
		`Use the CSV filename given via -csv as it is, without a date/time
		suffix. Repeated runs will then overwrite the same file instead of
		creating a new one.`)

	flag.StringVar(&f.alloyURL,
		"alloy",
		"",
		`Send the results to a Grafana Alloy OTLP HTTP endpoint as metrics
		instead of printing them, e.g. http://localhost:4318`)

	flag.StringVar(&f.influxDBURL,
		"influxdb",
		"",
		`Send the results to an InfluxDB v2 or v3 server as metrics instead
		of printing them, e.g. http://localhost:8086`)

	flag.StringVar(&f.influxDBOrg,
		"influxdb-org",
		"",
		`InfluxDB organization to write to. Required with the -influxdb flag.`)

	flag.StringVar(&f.influxDBBucket,
		"influxdb-bucket",
		"",
		`InfluxDB bucket to write to. Required with the -influxdb flag.`)

	flag.StringVar(&f.influxDBToken,
		"influxdb-token",
		"",
		`InfluxDB API token. Required with the -influxdb flag.
		Can also be given in the INFLUXDB_TOKEN environment variable, which
		keeps it out of the shell history.`)

	flag.Float64Var(&f.statsInterval,
		"stats-interval",
		10,
		`How often to send the statistics to Alloy or InfluxDB, in seconds.
		They are sent along with the next probe, so a longer probe interval
		delays them. No effect without the -alloy or -influxdb flag.`)

	flag.StringVar(&f.sourceLabel,
		"source-label",
		"",
		`Name this machine in the metrics sent to Alloy or InfluxDB, so that
		several machines probing the same target can be told apart.
		Defaults to the machine's hostname. No effect without the -alloy or
		-influxdb flag.`)

	flag.StringVar(&f.DBPath,
		"db",
		"",
		"Path and file name to store the output in a sqlite3 database.")

	flag.BoolVar(&f.showVer, "version", false, "Show version and exit.")

	flag.BoolVar(&f.checkUpdates, "u", false, "Check for updates and exit.")

	return &f
}

// validate rejects flag combinations that do not make sense. It exits the
// program instead of returning, like the rest of the input handling.
func (f *flags) validate() {
	if f.useIPv4 && f.useIPv6 {
		fmt.Fprintln(os.Stderr, "Only one IP version can be specified")
		usage()
	}

	if f.prettyJSON && !f.outputJSON {
		fmt.Fprintln(os.Stderr, "--pretty has no effect without the -j flag")
		usage()
	}

	// The file and metric printers always write their final record, so
	// there are no statistics to omit.
	if f.omitStatistics && (f.DBPath != "" || f.CSVPath != "" || f.alloyURL != "" || f.influxDBURL != "") {
		fmt.Fprintln(os.Stderr, "--no-stats has no effect when the output goes to a file, a database or a metrics endpoint")
		usage()
	}

	// Probing is driven by a time.Ticker, which panics on a zero or
	// negative interval.
	if secondsToDuration(f.intervalBetweenProbes) <= 0 {
		fmt.Fprintln(os.Stderr, "Interval between probes should be more than 0 seconds")
		os.Exit(1)
	}

	if (f.alloyURL != "" || f.influxDBURL != "") && secondsToDuration(f.statsInterval) <= 0 {
		fmt.Fprintln(os.Stderr, "Statistics interval should be more than 0 seconds")
		os.Exit(1)
	}
}

// newPrinterConfig collects everything the printers need out of the flags.
func (f *flags) newPrinterConfig(target string, port uint16) printers.Config {
	// The flag wins, but the environment variable is still accepted so the
	// token can be kept out of the shell history.
	influxDBToken := f.influxDBToken
	if influxDBToken == "" {
		influxDBToken = os.Getenv("INFLUXDB_TOKEN")
	}

	// Without this, several machines probing the same target would all
	// write to the same series and their numbers would be mixed together.
	sourceLabel := f.sourceLabel
	if sourceLabel == "" {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}
		sourceLabel = hostname
	}

	return printers.Config{
		Target:            target,
		Port:              port,
		OutputJSON:        f.outputJSON,
		PrettyJSON:        f.prettyJSON,
		NoColor:           f.noColor,
		WithTimestamp:     f.showTimestamp,
		WithSourceAddress: f.showSourceAddress,
		OmitStatistics:    f.omitStatistics,
		Verbose:           f.verbose,
		OutputDBPath:      f.DBPath,
		OutputCSVPath:     f.CSVPath,
		CSVNoTimestamp:    f.csvNoTimestamp,
		AlloyURL:          f.alloyURL,

		InfluxDBURL:    f.influxDBURL,
		InfluxDBOrg:    f.influxDBOrg,
		InfluxDBBucket: f.influxDBBucket,
		InfluxDBToken:  influxDBToken,

		StatsInterval: secondsToDuration(f.statsInterval),
		SourceLabel:   sourceLabel,
	}
}

// secondsToDuration returns the corresponding duration from seconds expressed with a float.
func secondsToDuration(seconds float64) time.Duration {
	return time.Duration(seconds * float64(time.Second))
}
