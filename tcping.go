package main

import (
	"bufio"
	"fmt"
	"math"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gookit/color"
)

type stats struct {
	startTime             time.Time
	endTime               time.Time
	startOfUptime         time.Time
	startOfDowntime       time.Time
	lastSuccessfulProbe   time.Time
	lastUnsuccessfulProbe time.Time
	hostname              string
	IP                    string
	port                  string
	rtt                   []uint
	longestUptime         longestTime
	longestDowntime       longestTime
	totalUptime           time.Duration
	totalDowntime         time.Duration
	totalSuccessfulPkts   uint
	totalUnsuccessfulPkts uint
	wasDown               bool // Used to determine the duration of a downtime
	isIP                  bool // If IP is provided instead of hostname, suppresses printing the IP information twice
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
	ThousandMilliSecond = 1000 * time.Millisecond
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
	color.Red.Printf("Try running %s like:\n", os.Args[0])
	color.Red.Printf("%s <hostname/ip> <port number> | for example:\n", os.Args[0])
	color.Red.Printf("%s www.example.com 443\n", os.Args[0])
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
func getInput() (string, string, string) {
	args := os.Args[1:]

	if len(args) != 2 {
		usage()
	}

	host := args[0]
	port := args[1]
	portInt, _ := strconv.Atoi(port)

	if portInt < 1 || portInt > 65535 {
		print("Port should be in 1..65535 range\n")
		os.Exit(1)
	}

	IP := resolveHostname(host)

	return host, port, IP
}

/* Hostname resolution */
func resolveHostname(host string) string {
	var IP string

	IPRaw := net.ParseIP(host)

	if IPRaw != nil {
		IP = IPRaw.String()
		return IP
	}

	IPaddr, err := net.LookupIP(host)

	if err != nil {
		color.Red.Printf("Failed to resolve %s\n", host)
		os.Exit(1)
	}

	IP = IPaddr[0].String()

	return IP
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
		timeStr = fmt.Sprintf("%d.%d.%d hours.minutes.seconds", hours, minutes, seconds)
		return timeStr
	} else if hours == 1 && minutes == 0 && seconds == 0 {
		timeStr = fmt.Sprintf("%d hour", hours)
		return timeStr
	} else if hours == 1 {
		timeStr = fmt.Sprintf("%d.%d.%d hour.minutes.seconds", hours, minutes, seconds)
		return timeStr
	}

	/* Calculate minutes */
	if minutes >= 2 {
		timeStr = fmt.Sprintf("%d.%d minutes.seconds", minutes, seconds)
		return timeStr
	} else if minutes == 1 && seconds == 0 {
		timeStr = fmt.Sprintf("%d minute", minutes)
		return timeStr
	} else if minutes == 1 {
		timeStr = fmt.Sprintf("%d.%d minute.seconds", minutes, seconds)
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

/* Print stattistics when program exits */
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
				senderMsg, tcpStats.IP, tcpStats.port, tcpStats.totalUnsuccessfulPkts)
		} else {
			colorLightGreen("%s from %s on port %s TCP_conn=%d time=%d ms\n",
				senderMsg, tcpStats.IP, tcpStats.port, tcpStats.totalSuccessfulPkts, rtt)
		}
	} else {
		if senderMsg == "No reply" {
			colorRed("%s from %s (%s) on port %s TCP_conn=%d\n",
				senderMsg, tcpStats.hostname, tcpStats.IP, tcpStats.port, tcpStats.totalUnsuccessfulPkts)
		} else {
			colorLightGreen("%s from %s (%s) on port %s TCP_conn=%d time=%d ms\n",
				senderMsg, tcpStats.hostname, tcpStats.IP, tcpStats.port, tcpStats.totalSuccessfulPkts, rtt)
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

	IPAndPort := net.JoinHostPort(tcpStats.IP, tcpStats.port)

	connStart := getSystemTime()
	conn, err := net.DialTimeout("tcp", IPAndPort, oneSecond)
	connEnd := time.Since(connStart)

	rtt := connEnd.Milliseconds()

	if err != nil {
		/* if the previous probe was successful
		and the current one failed: */
		if !tcpStats.wasDown {
			/* Update startOfDowntime */
			tcpStats.startOfDowntime = getSystemTime()

			/* Calculate the longest uptime */
			endOfUptime := getSystemTime()
			calcLongestUptime(tcpStats, endOfUptime)
			tcpStats.startOfUptime = time.Time{}

			tcpStats.wasDown = true
		}

		tcpStats.totalDowntime += time.Second
		tcpStats.totalUnsuccessfulPkts += 1
		tcpStats.lastUnsuccessfulProbe = getSystemTime()

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
			tcpStats.startOfUptime = getSystemTime()

			/* Calculate the longest downtime */
			endOfDowntime := getSystemTime()
			calcLongestDowntime(tcpStats, endOfDowntime)
			tcpStats.startOfDowntime = time.Time{}

			tcpStats.wasDown = false
		}

		/* It means it is the first time to get a response*/
		if tcpStats.startOfUptime.Format(timeFormat) == nullTimeFormat {
			tcpStats.startOfUptime = getSystemTime()
		}

		tcpStats.totalUptime += time.Second
		tcpStats.totalSuccessfulPkts += 1
		tcpStats.lastSuccessfulProbe = getSystemTime()

		tcpStats.rtt = append(tcpStats.rtt, uint(rtt))
		printReply(tcpStats, "Reply", rtt)

		defer conn.Close()
	}

	time.Sleep(ThousandMilliSecond - connEnd)
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

	host, port, IP := getInput()

	var tcpStats stats
	tcpStats.hostname = host
	tcpStats.IP = IP
	tcpStats.port = port
	tcpStats.startTime = getSystemTime()

	if host == IP {
		tcpStats.isIP = true
	}

	signalHandler(&tcpStats)

	color.LightCyan.Printf("TCPinging %s on port %s\n", host, port)

	stdinChan := make(chan string)
	go monitorStdin(stdinChan)

	for {
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
