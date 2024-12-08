package main

import (
	"bufio"
	"context"
	"flag"
	"math/rand"
	"net"
	"net/netip"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
	"github.com/google/go-github/v45/github"
)

var version = "" // set at compile time

const (
	owner      = "pouriyajamshidi"
	repo       = "tcping"
	dnsTimeout = 2 * time.Second
)

// printer is a set of methods for printers to implement.
//
// Printers should NOT modify any existing data nor do any calculations.
// They should only perform visual operations on given data.
type printer interface {
	// printStart should print the first message, after the program starts.
	// This message is printed only once, at the very beginning.
	printStart(hostname string, port uint16)

	// printProbeSuccess should print a message after each successful probe.
	// hostname could be empty, meaning it's pinging an address.
	// streak is the number of successful consecutive probes.
	printProbeSuccess(localAddr string, userInput userInput, streak uint, rtt float32)

	// printProbeFail should print a message after each failed probe.
	// hostname could be empty, meaning it's pinging an address.
	// streak is the number of successful consecutive probes.
	printProbeFail(userInput userInput, streak uint)

	// printRetryingToResolve should print a message with the hostname
	// it is trying to resolve an ip for.
	//
	// This is only being printed when the -r flag is applied.
	printRetryingToResolve(hostname string)

	// printTotalDownTime should print a downtime duration.
	//
	// This is being called when host was unavailable for some time
	// but the latest probe was successful (became available).
	printTotalDownTime(downtime time.Duration)

	// printStatistics should print a message with
	// helpful statistics information.
	//
	// This is being called on exit and when user hits "Enter".
	printStatistics(s tcping)

	// printVersion should print the current version.
	printVersion()

	// printInfo should a message, which is not directly related
	// to the pinging and serves as a helpful information.
	//
	// Example of such: new version with -u flag.
	printInfo(format string, args ...any)

	// printError should print an error message.
	// Printer should also apply \n to the given string, if needed.
	printError(format string, args ...any)
}

type tcping struct {
	printer                   // printer holds the chosen printer implementation for outputting information and data.
	startTime                 time.Time
	endTime                   time.Time
	startOfUptime             time.Time
	startOfDowntime           time.Time
	lastSuccessfulProbe       time.Time
	lastUnsuccessfulProbe     time.Time
	ticker                    *time.Ticker // ticker is used to handle time between probes.
	longestUptime             longestTime
	longestDowntime           longestTime
	rtt                       []float32
	hostnameChanges           []hostnameChange
	userInput                 userInput
	ongoingSuccessfulProbes   uint
	ongoingUnsuccessfulProbes uint
	totalDowntime             time.Duration
	totalUptime               time.Duration
	totalSuccessfulProbes     uint
	totalUnsuccessfulProbes   uint
	retriedHostnameLookups    uint
	rttResults                rttResult
	destWasDown               bool // destWasDown is used to determine the duration of a downtime
	destIsIP                  bool // destIsIP suppresses printing the IP information twice when hostname is not provided
}

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
	showLocalAddress         bool
}

type genericUserInputArgs struct {
	retryResolve         *uint
	probesBeforeQuit     *uint
	timeout              *float64
	secondsBetweenProbes *float64
	intName              *string
	showFailuresOnly     *bool
	showLocalAddress     *bool
	args                 []string
}

type networkInterface struct {
	remoteAddr *net.TCPAddr
	dialer     net.Dialer
	use        bool
}

type longestTime struct {
	start    time.Time
	end      time.Time
	duration time.Duration
}

type rttResult struct {
	min        float32
	max        float32
	average    float32
	hasResults bool
}

type hostnameChange struct {
	Addr netip.Addr `json:"addr,omitempty"`
	When time.Time  `json:"when,omitempty"`
}

// signalHandler catches SIGINT and SIGTERM then prints tcping stats
func signalHandler(tcping *tcping) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		shutdown(tcping)
	}()
}

// monitorSTDIN checks stdin to see whether the 'Enter' key was pressed
func monitorSTDIN(stdinChan chan bool) {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')

		if input == "\n" || input == "\r" || input == "\r\n" {
			stdinChan <- true
		}
	}
}

// printStats is a helper method for printStatistics
// for the current printer.
//
// This should be used instead, as it makes
// all the necessary calculations beforehand.
func (t *tcping) printStats() {
	if t.destWasDown {
		calcLongestDowntime(t, time.Since(t.startOfDowntime))
	} else {
		calcLongestUptime(t, time.Since(t.startOfUptime))
	}
	t.rttResults = calcMinAvgMaxRttTime(t.rtt)

	t.printStatistics(*t)
}

// shutdown calculates endTime, prints statistics and calls os.Exit(0).
// This should be used as the main exit-point.
func shutdown(tcping *tcping) {
	tcping.endTime = time.Now()
	tcping.printStats()

	// if the printer type is `database`, close it before exiting
	if db, ok := tcping.printer.(*database); ok {
		db.conn.Close()
	}

	// if the printer type is `csvPrinter`, call the cleanup function before exiting
	if cp, ok := tcping.printer.(*csvPrinter); ok {
		cp.cleanup()
	}

	os.Exit(0)
}

// usage prints how tcping should be run
func usage() {
	executableName := os.Args[0]

	colorLightCyan("\nTCPING version %s\n\n", version)
	colorRed("Try running %s like:\n", executableName)
	colorRed("%s <hostname/ip> <port number>. For example:\n", executableName)
	colorRed("%s www.example.com 443\n", executableName)
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
func setPrinter(tcping *tcping, outputJSON, prettyJSON *bool, noColor *bool, timeStamp *bool, localAddress *bool, outputDb *string, outputCSV *string, args []string) {
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
		tcping.printer, err = newCSVPrinter(*outputCSV, timeStamp, localAddress)
		if err != nil {
			tcping.printError("Failed to create CSV file: %s", err)
			os.Exit(1)
		}
	} else if *noColor == true {
		tcping.printer = newPlanePrinter(timeStamp)
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

	tcping.userInput.showLocalAddress = *genericArgs.showLocalAddress
}

// processUserInput gets and validate user input
func processUserInput(tcping *tcping) {
	useIPv4 := flag.Bool("4", false, "only use IPv4.")
	useIPv6 := flag.Bool("6", false, "only use IPv6.")
	retryHostnameResolveAfter := flag.Uint("r", 0, "retry resolving target's hostname after <n> number of failed probes. e.g. -r 10 to retry after 10 failed probes.")
	probesBeforeQuit := flag.Uint("c", 0, "stop after <n> probes, regardless of the result. By default, no limit will be applied.")
	outputJSON := flag.Bool("j", false, "output in JSON format.")
	prettyJSON := flag.Bool("pretty", false, "use indentation when using json output format. No effect without the '-j' flag.")
	noColor := flag.Bool("no-color", false, "do not colorize output.")
	showTimestamp := flag.Bool("D", false, "show timestamp in output.")
	saveToCSV := flag.String("csv", "", "path and file name to store tcping output to CSV file...If user prompts for stats, it will be saved to a file with the same name but _stats appended.")
	showVer := flag.Bool("v", false, "show version.")
	checkUpdates := flag.Bool("u", false, "check for updates and exit.")
	secondsBetweenProbes := flag.Float64("i", 1, "interval between sending probes. Real number allowed with dot as a decimal separator. The default is one second")
	timeout := flag.Float64("t", 1, "time to wait for a response, in seconds. Real number allowed. 0 means infinite timeout.")
	outputDB := flag.String("db", "", "path and file name to store tcping output to sqlite database.")
	interfaceName := flag.String("I", "", "interface name or address.")
	showLocalAddress := flag.Bool("show-local-address", false, "Show source address and port used for probe.")
	showFailuresOnly := flag.Bool("show-failures-only", false, "Show only the failed probes.")
	showHelp := flag.Bool("h", false, "show help message.")


	flag.CommandLine.Usage = usage

	permuteArgs(os.Args[1:])
	flag.Parse()

	// validation for flag and args
	args := flag.Args()

	// we need to set printers first, because they're used for
	// error reporting and other output.
	setPrinter(tcping, outputJSON, prettyJSON, noColor, showTimestamp, showLocalAddress, outputDB, saveToCSV, args)

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
		showLocalAddress:     showLocalAddress,
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
	for i := 0; i < len(args); i++ {
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

	localAddr := &net.TCPAddr{
		IP: interfaceAddress,
	}

	ni.dialer = net.Dialer{
		LocalAddr: localAddr,
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

// selectResolvedIP returns a single IPv4 or IPv6 address from the net.IP slice of resolved addresses
func selectResolvedIP(tcping *tcping, ipAddrs []netip.Addr) netip.Addr {
	var index int
	var ipList []netip.Addr
	var ip netip.Addr

	switch {
	case tcping.userInput.useIPv4:
		for _, ip := range ipAddrs {
			if ip.Is4() {
				ipList = append(ipList, ip)
			}
		}

		if len(ipList) == 0 {
			tcping.printError("Failed to find IPv4 address for %s", tcping.userInput.hostname)
			os.Exit(1)
		}

		if len(ipList) > 1 {
			index = rand.Intn(len(ipList))
		} else {
			index = 0
		}

		ip, _ = netip.ParseAddr(ipList[index].Unmap().String())

	case tcping.userInput.useIPv6:
		for _, ip := range ipAddrs {
			if ip.Is6() {
				ipList = append(ipList, ip)
			}
		}

		if len(ipList) == 0 {
			tcping.printError("Failed to find IPv6 address for %s", tcping.userInput.hostname)
			os.Exit(1)
		}

		if len(ipList) > 1 {
			index = rand.Intn(len(ipList))
		} else {
			index = 0
		}

		ip, _ = netip.ParseAddr(ipList[index].Unmap().String())

	default:
		if len(ipAddrs) > 1 {
			index = rand.Intn(len(ipAddrs))
		} else {
			index = 0
		}

		ip, _ = netip.ParseAddr(ipAddrs[index].Unmap().String())
	}

	return ip
}

// resolveHostname handles hostname resolution with a timeout value of a second
func resolveHostname(tcping *tcping) netip.Addr {
	ip, err := netip.ParseAddr(tcping.userInput.hostname)
	if err == nil {
		return ip
	}

	ctx, cancel := context.WithTimeout(context.Background(), dnsTimeout)
	defer cancel()

	ipAddrs, err := net.DefaultResolver.LookupNetIP(ctx, "ip", tcping.userInput.hostname)

	// Prevent tcping to exit if it has been running for a while
	if err != nil && (tcping.totalSuccessfulProbes != 0 || tcping.totalUnsuccessfulProbes != 0) {
		return tcping.userInput.ip
	} else if err != nil {
		tcping.printError("Failed to resolve %s: %s", tcping.userInput.hostname, err)
		os.Exit(1)
	}

	return selectResolvedIP(tcping, ipAddrs)
}

// retryResolveHostname retries resolving a hostname after certain number of failures
func retryResolveHostname(tcping *tcping) {
	if tcping.ongoingUnsuccessfulProbes >= tcping.userInput.retryHostnameLookupAfter {
		tcping.printRetryingToResolve(tcping.userInput.hostname)
		tcping.userInput.ip = resolveHostname(tcping)
		tcping.ongoingUnsuccessfulProbes = 0
		tcping.retriedHostnameLookups += 1

		// At this point hostnameChanges should have len > 0, but just in case
		if len(tcping.hostnameChanges) == 0 {
			return
		}

		lastAddr := tcping.hostnameChanges[len(tcping.hostnameChanges)-1].Addr
		if lastAddr != tcping.userInput.ip {
			tcping.hostnameChanges = append(tcping.hostnameChanges, hostnameChange{
				Addr: tcping.userInput.ip,
				When: time.Now(),
			})
		}
	}
}

// newLongestTime creates LongestTime structure
func newLongestTime(startTime time.Time, duration time.Duration) longestTime {
	return longestTime{
		start:    startTime,
		end:      startTime.Add(duration),
		duration: duration,
	}
}

// calcMinAvgMaxRttTime calculates min, avg and max RTT values
func calcMinAvgMaxRttTime(timeArr []float32) rttResult {
	var sum float32
	var result rttResult

	arrLen := len(timeArr)
	// rttResults.min = ^uint(0.0)
	if arrLen > 0 {
		result.min = timeArr[0]
	}

	for i := 0; i < arrLen; i++ {
		sum += timeArr[i]

		if timeArr[i] > result.max {
			result.max = timeArr[i]
		}

		if timeArr[i] < result.min {
			result.min = timeArr[i]
		}
	}

	if arrLen > 0 {
		result.hasResults = true
		result.average = sum / float32(arrLen)
	}

	return result
}

// calcLongestUptime calculates the longest uptime and sets it to tcpStats.
func calcLongestUptime(tcping *tcping, duration time.Duration) {
	if tcping.startOfUptime.IsZero() || duration == 0 {
		return
	}

	longestUptime := newLongestTime(tcping.startOfUptime, duration)

	// It means it is the first time we're calling this function
	if tcping.longestUptime.end.IsZero() {
		tcping.longestUptime = longestUptime
		return
	}

	if longestUptime.duration >= tcping.longestUptime.duration {
		tcping.longestUptime = longestUptime
	}
}

// calcLongestDowntime calculates the longest downtime and sets it to tcpStats.
func calcLongestDowntime(tcping *tcping, duration time.Duration) {
	if tcping.startOfDowntime.IsZero() || duration == 0 {
		return
	}

	longestDowntime := newLongestTime(tcping.startOfDowntime, duration)

	// It means it is the first time we're calling this function
	if tcping.longestDowntime.end.IsZero() {
		tcping.longestDowntime = longestDowntime
		return
	}

	if longestDowntime.duration >= tcping.longestDowntime.duration {
		tcping.longestDowntime = longestDowntime
	}
}

// nanoToMillisecond returns an amount of milliseconds from nanoseconds.
// Using duration.Milliseconds() is not an option, because it drops
// decimal points, returning an int.
func nanoToMillisecond(nano int64) float32 {
	return float32(nano) / float32(time.Millisecond)
}

// secondsToDuration returns the corresponding duration from seconds expressed with a float.
func secondsToDuration(seconds float64) time.Duration {
	return time.Duration(1000*seconds) * time.Millisecond
}

// maxDuration is the implementation of the math.Max function for time.Duration types.
// returns the longest duration of x or y.
func maxDuration(x, y time.Duration) time.Duration {
	if x > y {
		return x
	}
	return y
}

// handleConnError processes failed probes
func (t *tcping) handleConnError(connTime time.Time, elapsed time.Duration) {
	if !t.destWasDown {
		t.startOfDowntime = connTime
		uptime := t.startOfDowntime.Sub(t.startOfUptime)
		calcLongestUptime(t, uptime)
		t.startOfUptime = time.Time{}
		t.destWasDown = true
	}

	t.totalDowntime += elapsed
	t.lastUnsuccessfulProbe = connTime
	t.totalUnsuccessfulProbes += 1
	t.ongoingUnsuccessfulProbes += 1

	t.printProbeFail(
		t.userInput,
		t.ongoingUnsuccessfulProbes,
	)
}

// handleConnSuccess processes successful probes
func (t *tcping) handleConnSuccess(localAddr string, rtt float32, connTime time.Time, elapsed time.Duration) {
	if t.destWasDown {
		t.startOfUptime = connTime
		downtime := t.startOfUptime.Sub(t.startOfDowntime)
		calcLongestDowntime(t, downtime)
		t.printTotalDownTime(downtime)
		t.startOfDowntime = time.Time{}
		t.destWasDown = false
		t.ongoingUnsuccessfulProbes = 0
		t.ongoingSuccessfulProbes = 0
	}

	if t.startOfUptime.IsZero() {
		t.startOfUptime = connTime
	}

	t.totalUptime += elapsed
	t.lastSuccessfulProbe = connTime
	t.totalSuccessfulProbes += 1
	t.ongoingSuccessfulProbes += 1
	t.rtt = append(t.rtt, rtt)

	if !t.userInput.showFailuresOnly {
		t.printProbeSuccess(
			localAddr,
			t.userInput,
			t.ongoingSuccessfulProbes,
			rtt,
		)
	}
}

// tcpProbe pings a host, TCP style
func tcpProbe(tcping *tcping) {
	var err error
	var conn net.Conn
	connStart := time.Now()

	if tcping.userInput.networkInterface.use {
		// dialer already contains the timeout value
		conn, err = tcping.userInput.networkInterface.dialer.Dial("tcp", tcping.userInput.networkInterface.remoteAddr.String())
	} else {
		IPAndPort := netip.AddrPortFrom(tcping.userInput.ip, tcping.userInput.port)
		conn, err = net.DialTimeout("tcp", IPAndPort.String(), tcping.userInput.timeout)
	}

	connDuration := time.Since(connStart)
	rtt := nanoToMillisecond(connDuration.Nanoseconds())

	elapsed := maxDuration(connDuration, tcping.userInput.intervalBetweenProbes)

	if err != nil {
		tcping.handleConnError(connStart, elapsed)
	} else {
		tcping.handleConnSuccess(conn.LocalAddr().String(), rtt, connStart, elapsed)
		conn.Close()
	}
	<-tcping.ticker.C
}

func main() {
	tcping := &tcping{}
	processUserInput(tcping)
	tcping.ticker = time.NewTicker(tcping.userInput.intervalBetweenProbes)
	defer tcping.ticker.Stop()

	signalHandler(tcping)

	tcping.printStart(tcping.userInput.hostname, tcping.userInput.port)

	stdinchan := make(chan bool)
	go monitorSTDIN(stdinchan)

	var probeCount uint
	for {
		if tcping.userInput.shouldRetryResolve {
			retryResolveHostname(tcping)
		}

		tcpProbe(tcping)

		select {
		case pressedEnter := <-stdinchan:
			if pressedEnter {
				tcping.printStats()
			}
		default:
		}

		if tcping.userInput.probesBeforeQuit != 0 {
			probeCount++
			if probeCount == tcping.userInput.probesBeforeQuit {
				shutdown(tcping)
			}
		}
	}
}
