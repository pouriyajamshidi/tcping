package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
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

type stats struct {
	endTime               time.Time
	startOfUptime         time.Time
	startOfDowntime       time.Time
	lastSuccessfulProbe   time.Time
	lastUnsuccessfulProbe time.Time
	ip                    ipAddress
	startTime             time.Time
	statsPrinter
	retryHostnameResolveAfter uint // Retry resolving target's hostname after a certain number of failed requests
	hostname                  string
	rtt                       []float32
	ongoingUnsuccessfulProbes uint
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
}

type longestTime struct {
	start    time.Time
	end      time.Time
	duration float64
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

type ipAddress = netip.Addr
type cliArgs = []string
type calculatedTimeString = string

const (
	version             = "1.21.2"
	owner               = "pouriyajamshidi"
	repo                = "tcping"
	thousandMilliSecond = 1000 * time.Millisecond
	oneSecond           = 1 * time.Second
	timeFormat          = "2006-01-02 15:04:05"
	nullTimeFormat      = "0001-01-01 00:00:00"
	hourFormat          = "15:04:05"
)

/* Catch SIGINT and print tcping stats */
func signalHandler(tcpStats *stats) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		tcpStats.endTime = getSystemTime()
		tcpStats.printStatistics()
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
	retryHostnameResolveAfter := flag.Uint("r", 0, "retry resolving target's hostname after <n> number of failed requests. e.g. -r 10 for 10 failed probes.")
	shouldCheckUpdates := flag.Bool("u", false, "check for updates.")
	outputJson := flag.Bool("j", false, "output in JSON format.")
	prettyJson := flag.Bool("pretty", false, "use indentation when using json output format. No effect without the -j flag.")
	showVersion := flag.Bool("v", false, "show version.")
	useIPv4 := flag.Bool("4", false, "use IPv4 only.")
	useIPv6 := flag.Bool("6", false, "use IPv6 only.")

	flag.CommandLine.Usage = usage

	permuteArgs(os.Args[1:])
	flag.Parse()

	/* validation for flag and args */
	args := flag.Args()
	nFlag := flag.NFlag()

	if *retryHostnameResolveAfter > 0 {
		tcpStats.retryHostnameResolveAfter = *retryHostnameResolveAfter
	}

	/* -u works on its own. */
	if *shouldCheckUpdates {
		if len(args) == 0 && nFlag == 1 {
			checkLatestVersion()
		} else {
			usage()
		}
	}

	if *showVersion {
		colorGreen("TCPING version %s\n", version)
		os.Exit(0)
	}

	if *useIPv4 && *useIPv6 {
		colorRed("Only one IP version can be specified\n")
		usage()
	}

	if *useIPv4 {
		tcpStats.useIPv4 = true
	}

	if *useIPv6 {
		tcpStats.useIPv6 = true
	}

	if *prettyJson {
		if !*outputJson {
			colorRed("--pretty has no effect without the -j flag.\n")
			usage()
		}

		jsonEncoder.SetIndent("", "\t")
	}

	/* host and port must be specified　*/
	if len(args) != 2 {
		usage()
	}

	/* the non-flag command-line arguments */
	port, err := strconv.ParseUint(args[1], 10, 16)

	if err != nil {
		colorRed("Invalid port number: %s\n", args[1])
		os.Exit(1)
	}

	if port < 1 || port > 65535 {
		colorRed("Port should be in 1..65535 range\n")
		os.Exit(1)
	}

	tcpStats.hostname = args[0]
	tcpStats.port = uint16(port)
	tcpStats.ip = resolveHostname(tcpStats)
	tcpStats.startTime = getSystemTime()

	if tcpStats.hostname == tcpStats.ip.String() {
		tcpStats.isIP = true
	}

	if tcpStats.retryHostnameResolveAfter > 0 && !tcpStats.isIP {
		tcpStats.shouldRetryResolve = true
	}

	/* output format determination. */
	if *outputJson {
		tcpStats.statsPrinter = &statsJsonPrinter{stats: tcpStats}
	} else {
		tcpStats.statsPrinter = &statsPlanePrinter{stats: tcpStats}
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
func checkLatestVersion() {
	c := github.NewClient(nil)

	/* unauthenticated requests from the same IP are limited to 60 per hour. */
	latestRelease, _, err := c.Repositories.GetLatestRelease(context.Background(), owner, repo)
	if err != nil {
		colorRed("Failed to check for updates %s\n", err.Error())
		os.Exit(1)
	}

	reg := `^v?(\d+\.\d+\.\d+)$`
	latestTagName := latestRelease.GetTagName()
	latestVersion := regexp.MustCompile(reg).FindStringSubmatch(latestTagName)

	if len(latestVersion) == 0 {
		colorRed("Failed to check for updates. The version name does not match the rule: %s\n", latestTagName)
		os.Exit(1)
	}

	if latestVersion[1] != version {
		colorLightBlue("Found newer version %s\n", latestVersion[1])
		colorLightBlue("Please update TCPING from the URL below:\n")
		colorLightBlue("https://github.com/%s/%s/releases/tag/%s\n", owner, repo, latestTagName)
	} else {
		colorLightBlue("Newer version not found. %s is the latest version.\n", version)
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
		colorRed("Failed to resolve %s\n", tcpStats.hostname)
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
			colorRed("Failed to find IPv4 address for %s\n", tcpStats.hostname)
			os.Exit(1)
		}
		if len(ipAddrs) > 1 {
			index = rand.Intn(len(ipAddrs))
		} else {
			index = 0
		}
		ip, _ = netip.ParseAddr(ipAddrs[index].String())

	case tcpStats.useIPv6:
		for _, ip := range ipAddrs {
			if ip.To16() != nil {
				ipList = append(ipList, ip)
			}
		}
		if len(ipList) == 0 {
			colorRed("Failed to find IPv6 address for %s\n", tcpStats.hostname)
			os.Exit(1)
		}
		if len(ipAddrs) > 1 {
			index = rand.Intn(len(ipAddrs))
		} else {
			index = 0
		}
		ip, _ = netip.ParseAddr(ipAddrs[index].String())

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
	if tcpStats.ongoingUnsuccessfulProbes >= tcpStats.retryHostnameResolveAfter {
		tcpStats.printRetryingToResolve()
		tcpStats.ip = resolveHostname(tcpStats)
		tcpStats.ongoingUnsuccessfulProbes = 0
		tcpStats.retriedHostnameResolves += 1
	}
}

/* Create LongestTime structure */
func newLongestTime(startTime, endTime time.Time) longestTime {
	return longestTime{
		start:    startTime,
		end:      endTime,
		duration: endTime.Sub(startTime).Seconds(),
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

/* Calculate cumulative time */
func calcTime(time uint) calculatedTimeString {
	var timeStr string

	hours := time / (60 * 60)
	timeMod := time % (60 * 60)
	minutes := timeMod / (60)
	seconds := timeMod % (60)

	/* Calculate hours */
	if hours >= 2 {
		timeStr = fmt.Sprintf("%d hours %d minutes %d seconds", hours, minutes, seconds)
		return timeStr
	} else if hours == 1 && minutes == 0 && seconds == 0 {
		timeStr = fmt.Sprintf("%d hour", hours)
		return timeStr
	} else if hours == 1 {
		timeStr = fmt.Sprintf("%d hour %d minutes %d seconds", hours, minutes, seconds)
		return timeStr
	}

	/* Calculate minutes */
	if minutes >= 2 {
		timeStr = fmt.Sprintf("%d minutes %d seconds", minutes, seconds)
		return timeStr
	} else if minutes == 1 && seconds == 0 {
		timeStr = fmt.Sprintf("%d minute", minutes)
		return timeStr
	} else if minutes == 1 {
		timeStr = fmt.Sprintf("%d minute %d seconds", minutes, seconds)
		return timeStr
	}

	/* Calculate seconds */
	if seconds >= 2 {
		timeStr = fmt.Sprintf("%d seconds", seconds)
		return timeStr
	} else {
		timeStr = fmt.Sprintf("%d second", seconds)
		return timeStr
	}
}

/* Calculate the longest uptime */
func calcLongestUptime(tcpStats *stats, endOfUptime time.Time) {
	if tcpStats.startOfUptime.Format(timeFormat) == nullTimeFormat || endOfUptime.Format(timeFormat) == nullTimeFormat {
		return
	}

	longestUptime := newLongestTime(tcpStats.startOfUptime, endOfUptime)

	if tcpStats.longestUptime.end.Format(timeFormat) == nullTimeFormat {
		/* It means it is the first time we're calling this function */
		tcpStats.longestUptime = longestUptime
	} else if longestUptime.duration >= tcpStats.longestUptime.duration {
		tcpStats.longestUptime = longestUptime
	}
}

/* Calculate the longest downtime */
func calcLongestDowntime(tcpStats *stats, endOfDowntime time.Time) {
	if tcpStats.startOfDowntime.Format(timeFormat) == nullTimeFormat || endOfDowntime.Format(timeFormat) == nullTimeFormat {
		return
	}

	longestDowntime := newLongestTime(tcpStats.startOfDowntime, endOfDowntime)

	if tcpStats.longestDowntime.end.Format(timeFormat) == nullTimeFormat {
		/* It means it is the first time we're calling this function */
		tcpStats.longestDowntime = longestDowntime
	} else if longestDowntime.duration >= tcpStats.longestDowntime.duration {
		tcpStats.longestDowntime = longestDowntime
	}
}

/* Get current system time */
func getSystemTime() time.Time {
	return time.Now()
}

func nanoToMillisecond(nano int64) float32 {
	return float32(nano) / 1e6
}

func (tcpStats *stats) handleConnError(now time.Time) {
	if !tcpStats.wasDown {
		tcpStats.startOfDowntime = now
		calcLongestUptime(tcpStats, now)
		tcpStats.startOfUptime = time.Time{}
		tcpStats.wasDown = true
	}

	tcpStats.totalDowntime += time.Second
	tcpStats.lastUnsuccessfulProbe = now
	tcpStats.totalUnsuccessfulProbes += 1
	tcpStats.ongoingUnsuccessfulProbes += 1

	tcpStats.statsPrinter.printReply(replyMsg{msg: "No reply", rtt: 0})
}

func (tcpStats *stats) handleConnSuccess(rtt float32, now time.Time) {
	if tcpStats.wasDown {
		tcpStats.statsPrinter.printTotalDownTime(now)
		tcpStats.startOfUptime = now
		calcLongestDowntime(tcpStats, now)
		tcpStats.startOfDowntime = time.Time{}
		tcpStats.wasDown = false
		tcpStats.ongoingUnsuccessfulProbes = 0
	}

	if tcpStats.startOfUptime.Format(timeFormat) == nullTimeFormat {
		tcpStats.startOfUptime = now
	}

	tcpStats.totalUptime += time.Second
	tcpStats.lastSuccessfulProbe = now
	tcpStats.totalSuccessfulProbes += 1
	tcpStats.rtt = append(tcpStats.rtt, rtt)

	tcpStats.statsPrinter.printReply(replyMsg{msg: "Reply", rtt: rtt})
}

/* Ping host, TCP style */
func tcping(tcpStats *stats) {
	IPAndPort := netip.AddrPortFrom(tcpStats.ip, tcpStats.port)

	connStart := getSystemTime()
	conn, err := net.DialTimeout("tcp", IPAndPort.String(), oneSecond)
	connEnd := time.Since(connStart)

	rtt := nanoToMillisecond(connEnd.Nanoseconds())
	now := getSystemTime()

	if err != nil {
		tcpStats.handleConnError(now)
	} else {
		tcpStats.handleConnSuccess(rtt, now)
		conn.Close()
	}

	time.Sleep(thousandMilliSecond - connEnd)
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
	tcpStats := &stats{}
	processUserInput(tcpStats)
	signalHandler(tcpStats)
	tcpStats.printStart()

	stdinChan := make(chan string)
	go monitorStdin(stdinChan)

	for {
		if tcpStats.shouldRetryResolve {
			retryResolve(tcpStats)
		}

		tcping(tcpStats)

		/* print stats when the `enter` key is pressed */
		select {
		case stdin := <-stdinChan:
			if stdin == "\n" || stdin == "\r" || stdin == "\r\n" {
				tcpStats.printStatistics()
			}
		default:
			continue
		}
	}
}
