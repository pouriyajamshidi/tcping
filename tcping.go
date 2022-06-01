package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"math"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/gookit/color"
)

type stats struct {
	startTime                 time.Time
	endTime                   time.Time
	startOfUptime             time.Time
	startOfDowntime           time.Time
	lastSuccessfulProbe       time.Time
	lastUnsuccessfulProbe     time.Time
	retryHostnameResolveAfter *uint // Retry resolving target's hostname after a certain number of failed requests
	ip                        string
	port                      string
	hostname                  string
	rtt                       []uint
	totalDowntime             time.Duration
	totalUptime               time.Duration
	longestDowntime           longestTime
	totalSuccessfulPkts       uint
	totalUnsuccessfulPkts     uint
	ongoingUnsuccessfulPkts   uint
	retriedHostnameResolves   uint
	longestUptime             longestTime
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
	slowest    uint
	fastest    uint
	average    float32
	hasResults bool
}

const (
	version             = "1.9.0"
	owner               = "pouriyajamshidi"
	repo                = "tcping"
	thousandMilliSecond = 1000 * time.Millisecond
	oneSecond           = 1 * time.Second
	timeFormat          = "2006-01-02 15:04:05"
	nullTimeFormat      = "0001-01-01 00:00:00"
	hourFormat          = "15:04:05"
)

var (
	colorYellow      = color.Yellow.Printf
	colorGreen       = color.Green.Printf
	colorRed         = color.Red.Printf
	colorCyan        = color.Cyan.Printf
	colorLightYellow = color.LightYellow.Printf
	colorLightBlue   = color.FgLightBlue.Printf
	colorLightGreen  = color.LightGreen.Printf
)

/* Print how program should be run */
func usage() {
	commandName := os.Args[0]

	colorRed("TCPING verion %s\n\n", version)
	colorRed("Try running %s like:\n", commandName)
	colorRed("%s <hostname/ip> <port number> | for example:\n", commandName)
	colorRed("%s www.example.com 443\n", commandName)
	colorYellow("[optional]\n")

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
		printStatistics(tcpStats)
		os.Exit(0)
	}()
}

/* Get and validate user input */
func processUserInput(tcpStats *stats) {
	tcpStats.retryHostnameResolveAfter = flag.Uint("r", 0, "retry resolving target's hostname after <n> number of failed requests. e.g. -r 10 for 10 failed probes")
	shouldCheckUpdates := flag.Bool("u", false, "check for updates and display a message if there is a newer version.")
	flag.CommandLine.Usage = usage
	err := permuteArgs(os.Args[1:])
	if err != nil {
		usage()
	}
	flag.Parse()

	if *shouldCheckUpdates {
		checkLatestVersion()
	}

	/* the non-flag command-line arguments */
	args := flag.Args()
	port, _ := strconv.Atoi(args[1])

	if port < 1 || port > 65535 {
		print("Port should be in 1..65535 range\n")
		os.Exit(1)
	}

	tcpStats.hostname = args[0]
	tcpStats.port = strconv.Itoa(port)
	tcpStats.ip = resolveHostname(tcpStats)
	tcpStats.startTime = getSystemTime()

	if tcpStats.hostname == tcpStats.ip {
		tcpStats.isIP = true
	}

	if *tcpStats.retryHostnameResolveAfter > 0 && !tcpStats.isIP {
		tcpStats.shouldRetryResolve = true
	}
}

var errArgsFlagHasNoValue = errors.New("error: flag has no value")
var errArgsNotEnough = errors.New("error: not enough args")

/* Permute args for flag parsing stops just before the first non-flag argument.
see: https://pkg.go.dev/flag
*/
func permuteArgs(args []string) error {
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
					return errArgsFlagHasNoValue
				}
				/* the next flag has come */
				optionVal := args[i+1]
				if optionVal[0] == '-' {
					return errArgsFlagHasNoValue
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
	/* host and port must be specified　*/
	if len(nonFlagArgs) != 2 {
		return errArgsNotEnough
	}
	permutedArgs := append(flagArgs, nonFlagArgs...)

	/* replace args */
	for i := 0; i < len(args); i++ {
		args[i] = permutedArgs[i]
	}
	return nil
}

/* Check for updates and print messages if there is a newer version */
func checkLatestVersion() {
	c := github.NewClient(nil)

	/* unauthenticated requests from the same IP are limited to 60 per hour. */
	latestRelease, _, err := c.Repositories.GetLatestRelease(context.Background(), owner, repo)
	if err != nil {
		colorRed("Failed to check for updates %s\n", err.Error())
		return
	}

	reg := `^v?(\d+\.\d+\.\d+)$`
	latestTagName := latestRelease.GetTagName()
	latestVersion := regexp.MustCompile(reg).FindStringSubmatch(latestTagName)

	if len(latestVersion) == 0 {
		colorRed("Failed to check for updates. The version name does not match the rule: %s\n", latestTagName)
		return
	}

	if latestVersion[1] != version {
		colorLightBlue("Found newer version %s\n", latestVersion[1])
		colorLightBlue("Please update TCPING from the URL below: \n")
		colorLightBlue("https://github.com/%s/%s/releases/tag/%s \n\n", owner, repo, latestTagName)
	}
}

/* Hostname resolution */
func resolveHostname(tcpStats *stats) string {
	ipRaw := net.ParseIP(tcpStats.hostname)

	if ipRaw != nil {
		return ipRaw.String()
	}

	ipAddr, err := net.LookupIP(tcpStats.hostname)

	if err != nil && (tcpStats.totalSuccessfulPkts != 0 || tcpStats.totalUnsuccessfulPkts != 0) {
		/* Prevent exit if application has been running for a while */
		return tcpStats.ip
	} else if err != nil {
		color.Red.Printf("Failed to resolve %s\n", tcpStats.hostname)
		os.Exit(1)
	}

	return ipAddr[0].String()
}

/* Retry resolve hostname after certain number of failures */
func retryResolve(tcpStats *stats) {
	if tcpStats.ongoingUnsuccessfulPkts > *tcpStats.retryHostnameResolveAfter {
		colorLightYellow("Retrying to resolve %s\n", tcpStats.hostname)
		tcpStats.ip = resolveHostname(tcpStats)
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
	rttResults.fastest = ^uint(0)

	for i := 0; i < arrLen; i++ {
		accum += timeArr[i]

		if timeArr[i] > rttResults.slowest {
			rttResults.slowest = timeArr[i]
		}

		if timeArr[i] < rttResults.fastest {
			rttResults.fastest = timeArr[i]
		}
	}

	if arrLen > 0 {
		rttResults.hasResults = true
		rttResults.average = float32(accum) / float32(arrLen)
	}

	return rttResults
}

/* Calculate cumulative time */
func calcTime(time uint) string {
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

/* Print the last successful and unsuccessful probes */
func printLastSucUnsucProbes(lastSuccessfulProbe, lastUnsuccessfulProbe time.Time) {
	formattedLastSuccessfulProbe := lastSuccessfulProbe.Format(timeFormat)
	formattedLastUnsuccessfulProbe := lastUnsuccessfulProbe.Format(timeFormat)

	colorYellow("last successful probe:   ")
	if formattedLastSuccessfulProbe == nullTimeFormat {
		colorRed("Never succeeded\n")
	} else {
		colorGreen("%v\n", formattedLastSuccessfulProbe)
	}

	colorYellow("last unsuccessful probe: ")
	if formattedLastUnsuccessfulProbe == nullTimeFormat {
		colorGreen("Never failed\n")
	} else {
		colorRed("%v\n", formattedLastUnsuccessfulProbe)
	}
}

/* Print the start and end time of the program */
func printDurationStats(startTime, endTime time.Time) {
	var duration time.Time
	var durationDiff time.Duration

	colorYellow("--------------------------------------\n")
	colorYellow("TCPing started at: %v\n", startTime.Format(timeFormat))

	/* If the program was not terminated, no need to show the end time */
	if endTime.Format(timeFormat) == nullTimeFormat {
		durationDiff = time.Since(startTime)
	} else {
		colorYellow("TCPing ended at:   %v\n", endTime.Format(timeFormat))
		durationDiff = endTime.Sub(startTime)
	}

	duration = time.Time{}.Add(durationDiff)
	colorYellow("duration (HH:MM:SS): %v\n\n", duration.Format(hourFormat))
}

/* Print statistics when program exits */
func printStatistics(tcpStats *stats) {
	rttResults := findMinAvgMaxRttTime(tcpStats.rtt)

	if rttResults.hasResults {

		totalPackets := tcpStats.totalSuccessfulPkts + tcpStats.totalUnsuccessfulPkts
		totalUptime := calcTime(uint(tcpStats.totalUptime.Seconds()))
		totalDowntime := calcTime(uint(tcpStats.totalDowntime.Seconds()))
		packetLoss := (float32(tcpStats.totalUnsuccessfulPkts) / float32(totalPackets)) * 100

		/* general stats */
		colorYellow("\n--- %s TCPing statistics ---\n", tcpStats.hostname)
		colorYellow("%d probes transmitted, ", totalPackets)
		colorYellow("%d received, ", tcpStats.totalSuccessfulPkts)

		/* packet loss stats */
		if packetLoss == 0 {
			colorGreen("%.2f%%", packetLoss)
		} else if packetLoss > 0 && packetLoss <= 30 {
			colorLightYellow("%.2f%%", packetLoss)
		} else {
			colorRed("%.2f%%", packetLoss)
		}

		colorYellow(" packet loss\n")

		/* successful packet stats */
		colorYellow("successful probes:   ")
		colorGreen("%d\n", tcpStats.totalSuccessfulPkts)

		/* unsuccessful packet stats */
		colorYellow("unsuccessful probes: ")
		colorRed("%d\n", tcpStats.totalUnsuccessfulPkts)

		printLastSucUnsucProbes(tcpStats.lastSuccessfulProbe, tcpStats.lastUnsuccessfulProbe)

		/* uptime and downtime stats */
		colorYellow("total uptime: ")
		colorGreen("  %s\n", totalUptime)
		colorYellow("total downtime: ")
		colorRed("%s\n", totalDowntime)

		/* calculate the last longest time */
		if !tcpStats.wasDown {
			calcLongestUptime(tcpStats, tcpStats.lastSuccessfulProbe)
		} else {
			calcLongestDowntime(tcpStats, tcpStats.lastUnsuccessfulProbe)
		}

		/* longest uptime stats */
		printLongestUptime(tcpStats.longestUptime)

		/* longest downtime stats */
		printLongestDowntime(tcpStats.longestDowntime)

		/* resolve retry stats */
		if !tcpStats.isIP {
			printRetryResolveStats(tcpStats.retriedHostnameResolves)
		}

		/*TODO: see if formatted string would suit better */
		/* latency stats.*/
		colorYellow("rtt ")
		colorGreen("min")
		colorYellow("/")
		colorCyan("avg")
		colorYellow("/")
		colorRed("max: ")
		colorGreen("%d", rttResults.fastest)
		colorYellow("/")
		colorCyan("%.2f", rttResults.average)
		colorYellow("/")
		colorRed("%d", rttResults.slowest)
		colorYellow(" ms\n")

		/* duration stats */
		printDurationStats(tcpStats.startTime, tcpStats.endTime)
	}
}

/* Print TCP probe replies according to our policies */
func printReply(tcpStats *stats, senderMsg string, rtt int64) {
	if tcpStats.isIP {
		if senderMsg == "No reply" {
			colorRed("%s from %s on port %s TCP_conn=%d\n",
				senderMsg, tcpStats.ip, tcpStats.port, tcpStats.totalUnsuccessfulPkts)
		} else {
			colorLightGreen("%s from %s on port %s TCP_conn=%d time=%d ms\n",
				senderMsg, tcpStats.ip, tcpStats.port, tcpStats.totalSuccessfulPkts, rtt)
		}
	} else {
		if senderMsg == "No reply" {
			colorRed("%s from %s (%s) on port %s TCP_conn=%d\n",
				senderMsg, tcpStats.hostname, tcpStats.ip, tcpStats.port, tcpStats.totalUnsuccessfulPkts)
		} else {
			colorLightGreen("%s from %s (%s) on port %s TCP_conn=%d time=%d ms\n",
				senderMsg, tcpStats.hostname, tcpStats.ip, tcpStats.port, tcpStats.totalSuccessfulPkts, rtt)
		}
	}
}

/* Print the longest uptime */
func printLongestUptime(longestUptime longestTime) {
	if longestUptime.duration == 0 {
		return
	}

	uptime := calcTime(uint(math.Ceil(longestUptime.duration)))

	colorYellow("longest uptime:   ")
	colorGreen("%v ", uptime)
	colorYellow("from ")
	colorLightBlue("%v ", longestUptime.start.Format(timeFormat))
	colorYellow("to ")
	colorLightBlue("%v\n", longestUptime.end.Format(timeFormat))
}

/* Print the longest downtime */
func printLongestDowntime(longestDowntime longestTime) {
	if longestDowntime.duration == 0 {
		return
	}

	downtime := calcTime(uint(math.Ceil(longestDowntime.duration)))

	colorYellow("longest downtime: ")
	colorRed("%v ", downtime)
	colorYellow("from ")
	colorLightBlue("%v ", longestDowntime.start.Format(timeFormat))
	colorYellow("to ")
	colorLightBlue("%v\n", longestDowntime.end.Format(timeFormat))
}

/* Print the number of times that we tried resolving a hostname after a failure */
func printRetryResolveStats(retries uint) {
	colorYellow("Retried to resolve hostname ")
	colorRed("%d ", retries)
	colorYellow("times\n")
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

/* get current system time */
func getSystemTime() time.Time {
	return time.Now()
}

/* Ping host, TCP style */
func tcping(tcpStats *stats) {

	IPAndPort := net.JoinHostPort(tcpStats.ip, tcpStats.port)

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

		printReply(tcpStats, "No reply", 0)
	} else {
		/* if the previous probe failed
		and the current one succeeded: */
		if tcpStats.wasDown {
			/* calculate the total downtime since
			the previous successful probe */
			latestDowntimeDuration := time.Since(tcpStats.startOfDowntime).Seconds()
			calculatedDowntime := calcTime(uint(math.Ceil(latestDowntimeDuration)))
			color.Yellow.Printf("No response received for %s\n", calculatedDowntime)

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
		printReply(tcpStats, "Reply", rtt)

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
	var tcpStats stats
	processUserInput(&tcpStats)
	signalHandler(&tcpStats)

	color.LightCyan.Printf("TCPinging %s on port %s\n", tcpStats.hostname, tcpStats.port)

	stdinChan := make(chan string)
	go monitorStdin(stdinChan)

	for {
		if tcpStats.shouldRetryResolve {
			retryResolve(&tcpStats)
		}

		tcping(&tcpStats)

		/* print stats when the `enter` key is pressed */
		select {
		case stdin := <-stdinChan:
			if stdin == "\n" || stdin == "\r" || stdin == "\r\n" {
				printStatistics(&tcpStats)
			}
		default:
			continue
		}
	}
}
