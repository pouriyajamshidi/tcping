package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
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
	startTime             time.Time
	endTime               time.Time
	startOfUptime         time.Time
	startOfDowntime       time.Time
	lastSuccessfulProbe   time.Time
	lastUnsuccessfulProbe time.Time
	statsPrinter
	retryHostnameResolveAfter *uint // Retry resolving target's hostname after a certain number of failed requests
	ip                        ipAddress
	port                      string
	hostname                  string
	rtt                       []uint
	totalUnsuccessfulPkts     uint
	longestDowntime           longestTime
	totalSuccessfulPkts       uint
	totalUptime               time.Duration
	ongoingUnsuccessfulPkts   uint
	retriedHostnameResolves   uint
	longestUptime             longestTime
	totalDowntime             time.Duration
	wasDown                   bool // Used to determine the duration of a downtime
	isIP                      bool // If IP is provided instead of hostname, suppresses printing the IP information twice
	shouldRetryResolve        bool
}

type longestTime struct {
	start    time.Time
	end      time.Time
	duration float64
}

type rttResults struct {
	min        uint
	max        uint
	average    float32
	hasResults bool
}

type replyMsg struct {
	msg string
	rtt int64
}

type ipAddress = netip.Addr
type cliArgs = []string
type calculatedTimeString = string

const (
	version             = "1.12.1"
	owner               = "pouriyajamshidi"
	repo                = "tcping"
	thousandMilliSecond = 1000 * time.Millisecond
	oneSecond           = 1 * time.Second
	timeFormat          = "2006-01-02 15:04:05"
	nullTimeFormat      = "0001-01-01 00:00:00"
	hourFormat          = "15:04:05"
)

/* Print how program should be run */
func usage() {
	executableName := os.Args[0]

	colorLightCyan("\nTCPING version %s\n\n", version)
	colorRed("Try running %s like:\n", executableName)
	colorRed("%s <hostname/ip> <port number>. For example:\n", executableName)
	colorRed("%s www.example.com 443\n", executableName)
	colorYellow("\n[optional flags]\n")

	flag.VisitAll(func(f *flag.Flag) {
		colorYellow("  -%s : %s\n", f.Name, f.Usage)
	})

	os.Exit(1)
}

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

/* Get and validate user input */
func processUserInput(tcpStats *stats) {
	tcpStats.retryHostnameResolveAfter = flag.Uint("r", 0, "retry resolving target's hostname after <n> number of failed requests. e.g. -r 10 for 10 failed probes.")
	shouldCheckUpdates := flag.Bool("u", false, "check for updates.")
	outputJson := flag.Bool("j", false, "output in JSON format.")
	showVersion := flag.Bool("v", false, "show version.")

	flag.CommandLine.Usage = usage

	permuteArgs(os.Args[1:])
	flag.Parse()

	/* validation for flag and args */
	args := flag.Args()
	nFlag := flag.NFlag()

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

	/* host and port must be specified　*/
	if len(args) != 2 {
		usage()
	}

	/* the non-flag command-line arguments */
	port, _ := strconv.Atoi(args[1])
	if port < 1 || port > 65535 {
		print("Port should be in 1..65535 range\n")
		os.Exit(1)
	}

	tcpStats.hostname = args[0]
	tcpStats.port = strconv.Itoa(port)
    var err error
    tcpStats.ip, err = resolveHostname(tcpStats)
    if err != nil {
        colorRed(err.Error())
        os.Exit(1)
    }
	tcpStats.startTime = getSystemTime()

	if tcpStats.hostname == tcpStats.ip.String() {
		tcpStats.isIP = true
	}

	if *tcpStats.retryHostnameResolveAfter > 0 && !tcpStats.isIP {
		tcpStats.shouldRetryResolve = true
	}

	/* output format determination. */
	if *outputJson {
		tcpStats.statsPrinter = &statsJsonPrinter{stats: tcpStats}
	} else {
		tcpStats.statsPrinter = &statsPlanePrinter{stats: tcpStats}
	}
}

/* Permute args for flag parsing stops just before the first non-flag argument.
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
		colorLightBlue("Newer version not found . Your version %s is the latest.\n", version)
	}
	os.Exit(0)
}

/* Hostname resolution */
func resolveHostname(tcpStats *stats) (ipAddress, error) {
    ip, err := netip.ParseAddr(tcpStats.hostname)
    if err == nil {
        return ip, nil
    }

	ipAddr, err := net.LookupIP(tcpStats.hostname)

	if err != nil && (tcpStats.totalSuccessfulPkts != 0 || tcpStats.totalUnsuccessfulPkts != 0) {
		/* Prevent exit if application has been running for a while */
		return tcpStats.ip, nil
	} else if err != nil {
		return ip, fmt.Errorf("failed to resolve %s", tcpStats.hostname)
	}

	return netip.ParseAddr(ipAddr[0].String()) 
}

/* Retry resolve hostname after certain number of failures */
func retryResolve(tcpStats *stats) {
	if tcpStats.ongoingUnsuccessfulPkts > *tcpStats.retryHostnameResolveAfter {
		tcpStats.printRetryingToResolve()
		tcpStats.ip, _ = resolveHostname(tcpStats)
		tcpStats.ongoingUnsuccessfulPkts = 0
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
func findMinAvgMaxRttTime(timeArr []uint) rttResults {
	arrLen := len(timeArr)
	var accum uint

	var rttResults rttResults
	rttResults.min = ^uint(0)

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
		rttResults.average = float32(accum) / float32(arrLen)
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

/* Ping host, TCP style */
func tcping(tcpStats *stats) {

	IPAndPort := net.JoinHostPort(tcpStats.ip.String(), tcpStats.port)

	connStart := getSystemTime()
	conn, err := net.DialTimeout("tcp", IPAndPort, oneSecond)
	connEnd := time.Since(connStart)

	rtt := connEnd.Milliseconds()
	now := getSystemTime()

	if err != nil {
		/* if the previous probe was successful
		and the current one failed: */
		if !tcpStats.wasDown {
			/* Update startOfDowntime */
			tcpStats.startOfDowntime = now

			/* Calculate the longest uptime */
			endOfUptime := now
			calcLongestUptime(tcpStats, endOfUptime)
			tcpStats.startOfUptime = time.Time{}

			tcpStats.wasDown = true
		}

		tcpStats.totalDowntime += time.Second
		tcpStats.totalUnsuccessfulPkts += 1
		tcpStats.lastUnsuccessfulProbe = now
		tcpStats.ongoingUnsuccessfulPkts += 1

		tcpStats.printReply(replyMsg{msg: "No reply", rtt: 0})
	} else {
		/* if the previous probe failed
		and the current one succeeded: */
		if tcpStats.wasDown {
			/* calculate the total downtime since
			the previous successful probe */
			tcpStats.printTotalDownTime(now)

			/* Update startOfUptime */
			tcpStats.startOfUptime = now

			/* Calculate the longest downtime */
			endOfDowntime := now
			calcLongestDowntime(tcpStats, endOfDowntime)
			tcpStats.startOfDowntime = time.Time{}

			tcpStats.wasDown = false
			tcpStats.ongoingUnsuccessfulPkts = 0
		}

		/* It means it is the first time to get a response*/
		if tcpStats.startOfUptime.Format(timeFormat) == nullTimeFormat {
			tcpStats.startOfUptime = now
		}

		tcpStats.totalUptime += time.Second
		tcpStats.totalSuccessfulPkts += 1
		tcpStats.lastSuccessfulProbe = now

		tcpStats.rtt = append(tcpStats.rtt, uint(rtt))
		tcpStats.printReply(replyMsg{msg: "Reply", rtt: rtt})

		defer conn.Close()
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
