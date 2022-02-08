package main

import (
	"bufio"
	"math"
	"net"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"syscall"
	"time"

	"github.com/gookit/color"
)

type stats struct {
	startTime             time.Time
	endTime               time.Time
	startOfDowntime       time.Time
	endOfDowntime         time.Time
	lastSuccessfulProbe   time.Time
	lastUnsuccessfulProbe time.Time
	hostname              string
	IP                    string
	port                  string
	rtt                   []uint
	longestDowntime       longestDowntime
	totalUptime           time.Duration
	totalDowntime         time.Duration
	totalSuccessfulPkts   uint
	totalUnsuccessfulPkts uint
	wasDown               bool // Used to determine the duration of a downtime
	isIP                  bool // If IP is provided instead of hostname, suppresses printing the IP information twice
}

type longestDowntime struct {
	start time.Time
	end   time.Time
	time  float64
}

/* Print how program should be run */
func usage() {
	color.Red.Printf("Try running %s like:\n", os.Args[0])
	color.Red.Printf("%s <hostname/ip> <port number> | for example:\n", os.Args[0])
	color.Red.Printf("%s www.example.com 443\n", os.Args[0])
	os.Exit(1)
}

/* Catch SIGINT and print tcping stats */
func signalHandler(tcpStats *stats) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		tcpStats.endTime = time.Now()
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

/* Find min/avg/max RTT values. The last int acts as err code */
func findMinAvgMaxRttTime(timeArr []uint) (uint, float32, uint, bool) {
	isEmpty := true
	var avgTime uint
	arrLen := len(timeArr)

	for i := 0; i < arrLen; i++ {
		if timeArr[i] == 0 {
			continue
		}
		avgTime += timeArr[i]
	}

	if avgTime == 0 {
		/* prevents panics inside printStatistics func */
		return 0, 0, 0, isEmpty
	}

	sort.Slice(timeArr, func(i, j int) bool { return timeArr[i] < timeArr[j] })

	return timeArr[0], float32(avgTime) / float32(arrLen), timeArr[arrLen-1], !isEmpty
}

/* Calculate the correct number of seconds in calcTime func */
func calcSeconds(time float64) string {
	_, float := math.Modf(time)
	secondStr := strconv.FormatFloat(float*60, 'f', 0, 32)

	return secondStr
}

/* Calculate cumulative time */
func calcTime(time uint) string {
	var timeStr string

	if time == 1 {
		timeStr = strconv.FormatUint(uint64(time), 10) + " second"
		return timeStr
	} else if time < 60 {
		timeStr = strconv.FormatUint(uint64(time), 10) + " seconds"
		return timeStr
	} else {
		timeFloat := float64(time) / 60

		if timeFloat == 1.00 {
			return "1 minute"
		} else if timeFloat < 2.00 {
			timeMnt := int(timeFloat)
			timeSec := calcSeconds(timeFloat)
			timeStr := strconv.Itoa(timeMnt) + "." + timeSec + " minute.seconds"
			return timeStr
		}

		timeMnt := int(timeFloat)
		timeSec := calcSeconds(timeFloat)
		timeStr := strconv.Itoa(timeMnt) + "." + timeSec + " minutes.seconds"

		return timeStr
	}
}

/* Print the last successful and unsuccessful probes */
func printLastSucUnsucProbes(successfulProbe, unsuccessfulProbe time.Time) {
	cy := color.Yellow.Printf
	cg := color.Green.Printf
	cr := color.Red.Printf

	lastSuccessfulProbe := successfulProbe.Format("2006-01-02 15:04:05")
	lastUnsuccessfulProbe := unsuccessfulProbe.Format("2006-01-02 15:04:05")

	cy("last successful probe:   ")
	if lastSuccessfulProbe == "0001-01-01 00:00:00" {
		cr("Never succeeded\n")
	} else {
		cg("%v\n", lastSuccessfulProbe)
	}

	cy("last unsuccessful probe: ")
	if lastUnsuccessfulProbe == "0001-01-01 00:00:00" {
		cg("Never failed\n")
	} else {
		cr("%v\n", lastUnsuccessfulProbe)
	}
}

/* Print the start and end time of the program */
func printDurationStats(startTime, endTime time.Time) {
	var duration time.Time
	var durationDiff time.Duration

	cy := color.Yellow.Printf

	cy("--------------------------------------\n")
	cy("TCPing started at: %v\n", startTime.Format("2006-01-02 15:04:05"))

	/* If the program was not terminated, no need to show the end time */
	if endTime.Format("2006-01-02 15:04:05") == "0001-01-01 00:00:00" {
		durationDiff = time.Since(startTime)
	} else {
		cy("TCPing ended at:   %v\n", endTime.Format("2006-01-02 15:04:05"))
		durationDiff = endTime.Sub(startTime)
	}

	duration = time.Time{}.Add(durationDiff)
	cy("duration (HH:MM:SS): %v\n\n", duration.Format("15:04:05"))
}

/* Print stattistics when program exits */
func printStatistics(tcpStats *stats) {
	min, avg, max, isEmpty := findMinAvgMaxRttTime(tcpStats.rtt)

	if isEmpty {
		/* There are no results to show */
		return
	}

	totalPackets := tcpStats.totalSuccessfulPkts + tcpStats.totalUnsuccessfulPkts
	totalUptime := calcTime(uint(tcpStats.totalUptime.Seconds()))
	totalDowntime := calcTime(uint(tcpStats.totalDowntime.Seconds()))
	packetLoss := (float32(tcpStats.totalUnsuccessfulPkts) / float32(totalPackets)) * 100

	/* shortened color functions */
	cy := color.Yellow.Printf
	cg := color.Green.Printf
	cr := color.Red.Printf
	cc := color.Cyan.Printf
	cly := color.LightYellow.Printf

	/* general stats */
	cy("\n--- %s TCPing statistics ---\n", tcpStats.hostname)
	cy("%d probes transmitted, ", totalPackets)
	cy("%d received, ", tcpStats.totalSuccessfulPkts)

	/* packet loss stats */
	if packetLoss == 0 {
		cg("%.2f%%", packetLoss)
	} else if packetLoss > 0 && packetLoss <= 30 {
		cly("%.2f%%", packetLoss)
	} else {
		cr("%.2f%%", packetLoss)
	}
	cy(" packet loss\n")

	/* successful packet stats */
	cy("successful probes:   ")
	cg("%d\n", tcpStats.totalSuccessfulPkts)

	/* unsuccessful packet stats */
	cy("unsuccessful probes: ")
	cr("%d\n", tcpStats.totalUnsuccessfulPkts)

	printLastSucUnsucProbes(tcpStats.lastSuccessfulProbe, tcpStats.lastUnsuccessfulProbe)

	/* uptime and downtime stats */
	cy("total uptime: ")
	cg("  %s\n", totalUptime)
	cy("total downtime: ")
	cr("%s\n", totalDowntime)

	/* longest downtime stats */
	printLongestDowntime(tcpStats.longestDowntime.time, tcpStats.longestDowntime.start, tcpStats.longestDowntime.end)

	/*TODO: see if formatted string would suit better */
	/* latency stats.*/
	cy("rtt ")
	cg("min")
	cy("/")
	cc("avg")
	cy("/")
	cr("max: ")
	cg("%d", min)
	cy("/")
	cc("%.2f", avg)
	cy("/")
	cr("%d", max)
	cy(" ms\n")

	/* duration stats */
	printDurationStats(tcpStats.startTime, tcpStats.endTime)
}

/* Print TCP probe replies according to our policies */
func printReply(tcpStats *stats, senderMsg string, rtt int64) {
	// TODO: Refactor

	cr := color.Red.Printf
	cg := color.LightGreen.Printf

	if tcpStats.isIP {
		if senderMsg == "No reply" {
			cr("%s from %s on port %s TCP_conn=%d\n",
				senderMsg, tcpStats.IP, tcpStats.port, tcpStats.totalUnsuccessfulPkts)
		} else {
			cg("%s from %s on port %s TCP_conn=%d time=%d ms\n",
				senderMsg, tcpStats.IP, tcpStats.port, tcpStats.totalSuccessfulPkts, rtt)
		}
	} else {
		if senderMsg == "No reply" {
			cr("%s from %s (%s) on port %s TCP_conn=%d\n",
				senderMsg, tcpStats.hostname, tcpStats.IP, tcpStats.port, tcpStats.totalUnsuccessfulPkts)
		} else {
			cg("%s from %s (%s) on port %s TCP_conn=%d time=%d ms\n",
				senderMsg, tcpStats.hostname, tcpStats.IP, tcpStats.port, tcpStats.totalSuccessfulPkts, rtt)
		}
	}
}

/* Print the longest downtime */
func printLongestDowntime(longestDowntime float64, startTime, endTime time.Time) {
	cy := color.Yellow.Printf
	clb := color.FgLightBlue.Printf
	cr := color.Red.Printf

	if longestDowntime == 0 {
		return
	}

	downtime := calcTime(uint(math.Ceil(longestDowntime)))

	cy("longest downtime: ")
	cr("%v ", downtime)
	cy("from ")
	clb("%v ", startTime.Format("2006-01-02 15:04:05"))
	cy("to ")
	clb("%v\n", endTime.Format("2006-01-02 15:04:05"))
}

/* Calculate the longest downtime */
func calcLongestDowntime(tcpStats *stats) {

	latestStartOfDowntime := tcpStats.startOfDowntime
	latestEndOfDowntime := tcpStats.endOfDowntime

	if tcpStats.longestDowntime.end.Format("2006-01-02 15:04:05") == "0001-01-01 00:00:00" {
		/* It means it is the first time we're calling this function */
		tcpStats.longestDowntime.start = latestStartOfDowntime
		tcpStats.longestDowntime.end = latestEndOfDowntime
		tcpStats.longestDowntime.time = latestEndOfDowntime.Sub(latestStartOfDowntime).Seconds()
	} else {
		downtimeDuration := latestEndOfDowntime.Sub(latestStartOfDowntime).Seconds()

		if downtimeDuration >= tcpStats.longestDowntime.time {
			tcpStats.longestDowntime.start = latestStartOfDowntime
			tcpStats.longestDowntime.end = latestEndOfDowntime
			tcpStats.longestDowntime.time = downtimeDuration
		}
	}
}

/* Ping host, TCP style */
func tcping(tcpStats *stats) {

	IPAndPort := net.JoinHostPort(tcpStats.IP, tcpStats.port)

	connAttempt := time.Now()
	conn, err := net.DialTimeout("tcp", IPAndPort, 1*time.Second)
	connEnd := time.Since(connAttempt)
	rtt := connEnd.Milliseconds()

	if err != nil {
		/* if the previous probe was successful
		and the current one failed: */
		if !tcpStats.wasDown {
			tcpStats.startOfDowntime = time.Now()
			tcpStats.wasDown = true
		}

		tcpStats.totalDowntime += time.Second
		tcpStats.totalUnsuccessfulPkts++
		tcpStats.lastUnsuccessfulProbe = time.Now()

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

			tcpStats.endOfDowntime = time.Now()
			calcLongestDowntime(tcpStats)

			tcpStats.wasDown = false
		}

		tcpStats.totalUptime += time.Second
		tcpStats.totalSuccessfulPkts++
		tcpStats.lastSuccessfulProbe = time.Now()

		tcpStats.rtt = append(tcpStats.rtt, uint(rtt))
		printReply(tcpStats, "Reply", rtt)

		defer conn.Close()
	}

	time.Sleep((1000 * time.Millisecond) - connEnd)
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
	tcpStats.startTime = time.Now()

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
