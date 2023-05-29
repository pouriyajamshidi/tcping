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
	"syscall"
	"time"

	"github.com/google/go-github/v45/github"
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
	printProbeSuccess(hostname, ip string, port uint16, streak uint, rtt float32)

	// printProbeFail should print a message after each failed probe.
	// hostname could be empty, meaning it's pinging an address.
	// streak is the number of successful consecutive probes.
	printProbeFail(hostname, ip string, port uint16, streak uint)

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
	printStatistics(s stats)

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

type stats struct {
	endTime                   time.Time
	startOfUptime             time.Time
	startOfDowntime           time.Time
	lastSuccessfulProbe       time.Time
	lastUnsuccessfulProbe     time.Time
	ip                        ipAddress
	startTime                 time.Time
	retryHostnameResolveAfter uint // Retry resolving target's hostname after a certain number of failed requests
	hostname                  string
	hostnameChanges           []hostnameChange
	rtt                       []float32
	rttResults                rttResults
	ongoingUnsuccessfulProbes uint
	ongoingSuccessfulProbes   uint
	longestDowntime           longestTime
	totalSuccessfulProbes     uint
	totalUptime               time.Duration
	retriedHostnameResolves   uint
	longestUptime             longestTime
	totalDowntime             time.Duration
	totalUnsuccessfulProbes   uint
	port                      uint16
	wasDown                   bool // Used to determine the duration of a downtime
	isIP                      bool // If IP is provided instead of hostname, suppresses printing the IP information twice
	shouldRetryResolve        bool
	useIPv4                   bool
	useIPv6                   bool
	probesBeforeQuit          uint

	// ticker is used to handle time between probes.
	ticker *time.Ticker

	// printer holds the choosen printer implementation for
	// outputting information and data.
	printer printer
}

type longestTime struct {
	start    time.Time
	end      time.Time
	duration time.Duration
}

type rttResults struct {
	min        float32
	max        float32
	average    float32
	hasResults bool
}

type replyMsg struct {
	msg string
	rtt float32
}

type hostnameChange struct {
	Addr netip.Addr `json:"addr,omitempty"`
	When time.Time  `json:"when,omitempty"`
}

type (
	ipAddress = netip.Addr
	cliArgs   = []string
)

const (
	version = "1.22.1"
	owner   = "pouriyajamshidi"
	repo    = "tcping"
)

/* Catch SIGINT and print tcping stats */
func signalHandler(tcpStats *stats) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		totalRuntime := tcpStats.totalUnsuccessfulProbes + tcpStats.totalSuccessfulProbes
		tcpStats.endTime = tcpStats.startTime.Add(time.Duration(totalRuntime) * time.Second)
		tcpStats.printStats()
		os.Exit(0)
	}()
}

/* Print how program should be run */
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

/* Get and validate user input */
func processUserInput(tcpStats *stats) {
	useIPv4 := flag.Bool("4", false, "use IPv4 only.")
	useIPv6 := flag.Bool("6", false, "use IPv6 only.")
	retryHostnameResolveAfter := flag.Uint("r", 0, "retry resolving target's hostname after <n> number of failed requests. e.g. -r 10 for 10 failed probes.")
	probesBeforeQuit := flag.Uint("c", 0, "do n probes and quit, regardless of the result. If c is 0 (default), no limit will be applied")
	outputJson := flag.Bool("j", false, "output in JSON format.")
	prettyJson := flag.Bool("pretty", false, "use indentation when using json output format. No effect without the -j flag.")
	showVersion := flag.Bool("v", false, "show version.")
	shouldCheckUpdates := flag.Bool("u", false, "check for updates.")

	flag.CommandLine.Usage = usage

	permuteArgs(os.Args[1:])
	flag.Parse()

	/* validation for flag and args */
	args := flag.Args()
	nFlag := flag.NFlag()

	// we need to set printers first, because they're used for
	// errors reporting and other output.
	if *outputJson {
		tcpStats.printer = newJsonPrinter(*prettyJson)
	} else {
		tcpStats.printer = &planePrinter{}
	}

	if *prettyJson && !*outputJson {
		tcpStats.printer.printError("--pretty has no effect without the -j flag.")
		usage()
	}

	if *retryHostnameResolveAfter > 0 {
		tcpStats.retryHostnameResolveAfter = *retryHostnameResolveAfter
	}

	/* -u works on its own. */
	if *shouldCheckUpdates {
		if len(args) == 0 && nFlag == 1 {
			checkLatestVersion(tcpStats.printer)
		} else {
			usage()
		}
	}

	if *showVersion {
		tcpStats.printer.printVersion()
		os.Exit(0)
	}

	if *useIPv4 && *useIPv6 {
		tcpStats.printer.printError("Only one IP version can be specified")
		usage()
	}

	if *useIPv4 {
		tcpStats.useIPv4 = true
	}

	if *useIPv6 {
		tcpStats.useIPv6 = true
	}

	/* host and port must be specified　*/
	if len(args) != 2 {
		usage()
	}

	/* the non-flag command-line arguments */
	port, err := strconv.ParseUint(args[1], 10, 16)
	if err != nil {
		tcpStats.printer.printError("Invalid port number: %s", args[1])
		os.Exit(1)
	}

	if port < 1 || port > 65535 {
		tcpStats.printer.printError("Port should be in 1..65535 range")
		os.Exit(1)
	}

	tcpStats.hostname = args[0]
	tcpStats.port = uint16(port)
	tcpStats.ip = resolveHostname(tcpStats)
	tcpStats.startTime = time.Now()
	tcpStats.probesBeforeQuit = *probesBeforeQuit

	// this serves as a default starting value for tracking changes.
	tcpStats.hostnameChanges = []hostnameChange{
		{tcpStats.ip, time.Now()},
	}

	if tcpStats.hostname == tcpStats.ip.String() {
		tcpStats.isIP = true
	}

	if tcpStats.retryHostnameResolveAfter > 0 && !tcpStats.isIP {
		tcpStats.shouldRetryResolve = true
	}
}

/*
	Permute args for flag parsing stops just before the first non-flag argument.

see: https://pkg.go.dev/flag
*/
func permuteArgs(args cliArgs) {
	var flagArgs []string
	var nonFlagArgs []string

	for i := 0; i < len(args); i++ {
		v := args[i]
		if v[0] == '-' {
			optionName := v[1:]
			switch optionName {
			case "c":
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

/* Check for updates and print messages if there is a newer version */
func checkLatestVersion(p printer) {
	c := github.NewClient(nil)

	/* unauthenticated requests from the same IP are limited to 60 per hour. */
	latestRelease, _, err := c.Repositories.GetLatestRelease(context.Background(), owner, repo)
	if err != nil {
		p.printError("Failed to check for updates %s", err.Error())
		os.Exit(1)
	}

	reg := `^v?(\d+\.\d+\.\d+)$`
	latestTagName := latestRelease.GetTagName()
	latestVersion := regexp.MustCompile(reg).FindStringSubmatch(latestTagName)

	if len(latestVersion) == 0 {
		p.printError("Failed to check for updates. The version name does not match the rule: %s", latestTagName)
		os.Exit(1)
	}

	if latestVersion[1] != version {
		p.printInfo("Found newer version %s", latestVersion[1])
		p.printInfo("Please update TCPING from the URL below:")
		p.printInfo("https://github.com/%s/%s/releases/tag/%s",
			owner, repo, latestTagName)
	} else {
		p.printInfo("Newer version not found. %s is the latest version.",
			version)
	}
	os.Exit(0)
}

/* Hostname resolution */
func resolveHostname(tcpStats *stats) ipAddress {
	ip, err := netip.ParseAddr(tcpStats.hostname)
	if err == nil {
		return ip
	}

	ipAddrs, err := net.LookupIP(tcpStats.hostname)

	if err != nil && (tcpStats.totalSuccessfulProbes != 0 || tcpStats.totalUnsuccessfulProbes != 0) {
		/* Prevent exit if application has been running for a while */
		return tcpStats.ip
	} else if err != nil {
		tcpStats.printer.printError("Failed to resolve %s", tcpStats.hostname)
		os.Exit(1)
	}

	var index int
	var ipList []net.IP

	switch {
	case tcpStats.useIPv4:
		for _, ip := range ipAddrs {
			if ip.To4() != nil {
				ipList = append(ipList, ip)
			}
		}
		if len(ipList) == 0 {
			tcpStats.printer.printError("Failed to find IPv4 address for %s", tcpStats.hostname)
			os.Exit(1)
		}
		if len(ipList) > 1 {
			index = rand.Intn(len(ipAddrs))
		} else {
			index = 0
		}
		ip, _ = netip.ParseAddr(ipList[index].String())

	case tcpStats.useIPv6:
		for _, ip := range ipAddrs {
			if ip.To16() != nil {
				ipList = append(ipList, ip)
			}
		}
		if len(ipList) == 0 {
			tcpStats.printer.printError("Failed to find IPv6 address for %s", tcpStats.hostname)
			os.Exit(1)
		}
		if len(ipList) > 1 {
			index = rand.Intn(len(ipAddrs))
		} else {
			index = 0
		}
		ip, _ = netip.ParseAddr(ipList[index].String())

	default:
		if len(ipAddrs) > 1 {
			index = rand.Intn(len(ipAddrs))
		} else {
			index = 0
		}
		ip, _ = netip.ParseAddr(ipAddrs[index].String())
	}

	return ip
}

/* Retry resolve hostname after certain number of failures */
func retryResolve(tcpStats *stats) {
	// If we have a limit set (-r) and not reached it yet, just continue pinging.
	if tcpStats.retryHostnameResolveAfter != 0 &&
		tcpStats.ongoingUnsuccessfulProbes < tcpStats.retryHostnameResolveAfter {
		return
	}

	tcpStats.printer.printRetryingToResolve(tcpStats.hostname)
	tcpStats.ip = resolveHostname(tcpStats)
	tcpStats.ongoingUnsuccessfulProbes = 0
	tcpStats.retriedHostnameResolves += 1

	// At this point hostnameChanges should have len > 0, but just in case
	if len(tcpStats.hostnameChanges) == 0 {
		return
	}

	lastAddr := tcpStats.hostnameChanges[len(tcpStats.hostnameChanges)-1].Addr
	if lastAddr != tcpStats.ip {
		tcpStats.hostnameChanges = append(tcpStats.hostnameChanges, hostnameChange{
			Addr: tcpStats.ip,
			When: time.Now(),
		})
	}
}

/* Create LongestTime structure */
func newLongestTime(startTime time.Time, duration time.Duration) longestTime {
	return longestTime{
		start:    startTime,
		end:      startTime.Add(duration),
		duration: duration,
	}
}

/* Find min/avg/max RTT values. The last int acts as err code */
func findMinAvgMaxRttTime(timeArr []float32) rttResults {
	var accum float32
	var rttResults rttResults

	arrLen := len(timeArr)
	// rttResults.min = ^uint(0.0)
	if arrLen > 0 {
		rttResults.min = timeArr[0]
	}

	for i := 0; i < arrLen; i++ {
		accum += timeArr[i]

		if timeArr[i] > rttResults.max {
			rttResults.max = timeArr[i]
		}

		if timeArr[i] < rttResults.min {
			rttResults.min = timeArr[i]
		}
	}

	if arrLen > 0 {
		rttResults.hasResults = true
		rttResults.average = accum / float32(arrLen)
	}

	return rttResults
}

// calcLongestUptime calculates the longest uptime and sets it to tcpStats.
func calcLongestUptime(tcpStats *stats, duration time.Duration) {
	if tcpStats.startOfUptime.IsZero() || duration == 0 {
		return
	}

	longestUptime := newLongestTime(tcpStats.startOfUptime, duration)

	// It means it is the first time we're calling this function
	if tcpStats.longestUptime.end.IsZero() {
		tcpStats.longestUptime = longestUptime
		return
	}

	if longestUptime.duration >= tcpStats.longestUptime.duration {
		tcpStats.longestUptime = longestUptime
	}
}

// calcLongestDowntime calculates the longest downtime and sets it to tcpStats.
func calcLongestDowntime(tcpStats *stats, duration time.Duration) {
	if tcpStats.startOfDowntime.IsZero() || duration == 0 {
		return
	}

	longestDowntime := newLongestTime(tcpStats.startOfDowntime, duration)

	// It means it is the first time we're calling this function
	if tcpStats.longestDowntime.end.IsZero() {
		tcpStats.longestDowntime = longestDowntime
		return
	}

	if longestDowntime.duration >= tcpStats.longestDowntime.duration {
		tcpStats.longestDowntime = longestDowntime
	}
}

// nanoToMillisecond returns an amount of milliseconds from nanoseconds.
// Using duration.Milliseconds() is not an option, because it drops
// decimal points, returning an int.
func nanoToMillisecond(nano int64) float32 {
	return float32(nano) / float32(time.Millisecond)
}

func (tcpStats *stats) handleConnError(now time.Time) {
	if !tcpStats.wasDown {
		tcpStats.startOfDowntime = now
		calcLongestUptime(tcpStats,
			time.Duration(tcpStats.ongoingSuccessfulProbes)*time.Second)
		tcpStats.startOfUptime = time.Time{}
		tcpStats.wasDown = true
	}

	tcpStats.totalDowntime += time.Second
	tcpStats.lastUnsuccessfulProbe = now
	tcpStats.totalUnsuccessfulProbes += 1
	tcpStats.ongoingUnsuccessfulProbes += 1

	tcpStats.printer.printProbeFail(
		tcpStats.hostname,
		tcpStats.ip.String(),
		tcpStats.port,
		tcpStats.ongoingUnsuccessfulProbes,
	)
}

func (tcpStats *stats) handleConnSuccess(rtt float32, now time.Time) {
	if tcpStats.wasDown {
		tcpStats.startOfUptime = now
		downtime := time.Since(tcpStats.startOfDowntime).Truncate(time.Second)
		calcLongestDowntime(tcpStats, downtime)
		tcpStats.printer.printTotalDownTime(downtime)
		tcpStats.startOfDowntime = time.Time{}
		tcpStats.wasDown = false
		tcpStats.ongoingUnsuccessfulProbes = 0
		tcpStats.ongoingSuccessfulProbes = 0
	}

	if tcpStats.startOfUptime.IsZero() {
		tcpStats.startOfUptime = now
	}

	tcpStats.totalUptime += time.Second
	tcpStats.lastSuccessfulProbe = now
	tcpStats.totalSuccessfulProbes += 1
	tcpStats.ongoingSuccessfulProbes += 1
	tcpStats.rtt = append(tcpStats.rtt, rtt)

	tcpStats.printer.printProbeSuccess(
		tcpStats.hostname,
		tcpStats.ip.String(),
		tcpStats.port,
		tcpStats.ongoingSuccessfulProbes,
		rtt,
	)
}

// printStats is a helper method for printStatistics
// for the current printer.
//
// This should be used instead, as it makes
// all the nescessary calculations beforehand.
func (tcpStats *stats) printStats() {
	calcLongestUptime(tcpStats,
		time.Duration(tcpStats.ongoingSuccessfulProbes)*time.Second)
	calcLongestDowntime(tcpStats,
		time.Duration(tcpStats.ongoingUnsuccessfulProbes)*time.Second)

	tcpStats.rttResults = findMinAvgMaxRttTime(tcpStats.rtt)

	tcpStats.printer.printStatistics(*tcpStats)
}

/* Ping host, TCP style */
func tcping(tcpStats *stats) {
	IPAndPort := netip.AddrPortFrom(tcpStats.ip, tcpStats.port)

	connStart := time.Now()
	conn, err := net.DialTimeout("tcp", IPAndPort.String(), time.Second)
	connEnd := time.Since(connStart)
	rtt := nanoToMillisecond(connEnd.Nanoseconds())
	now := time.Now()

	if err != nil {
		tcpStats.handleConnError(now)
	} else {
		tcpStats.handleConnSuccess(rtt, now)
		conn.Close()
	}

	<-tcpStats.ticker.C
}

/* Capture keystrokes from stdin */
func monitorStdin(stdinChan chan string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		key, _ := reader.ReadString('\n')
		stdinChan <- key
	}
}

func main() {
	tcpStats := &stats{
		ticker: time.NewTicker(time.Second),
	}
	defer tcpStats.ticker.Stop()
	processUserInput(tcpStats)
	signalHandler(tcpStats)
	tcpStats.printer.printStart(tcpStats.hostname, tcpStats.port)

	stdinChan := make(chan string)
	go monitorStdin(stdinChan)

	var probeCount uint = 0
	for {
		if tcpStats.shouldRetryResolve {
			retryResolve(tcpStats)
		}

		tcping(tcpStats)

		/* print stats when the `enter` key is pressed */
		select {
		case stdin := <-stdinChan:
			if stdin == "\n" || stdin == "\r" || stdin == "\r\n" {
				tcpStats.printStats()
			}
		default:
		}

		if tcpStats.probesBeforeQuit == 0 {
			continue
		}

		probeCount++
		if probeCount == tcpStats.probesBeforeQuit {
			tcpStats.printStats()
			return
		}
	}
}
