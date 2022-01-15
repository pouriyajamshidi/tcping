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
	ongoingDowntime       time.Time
	lastSuccessfulProbe   time.Time
	hostname              string
	IP                    string
	port                  string
	rtt                   []uint
	totalSuccessfulPkts   uint
	totalUptime           time.Duration
	totalDowntime         time.Duration
	totalUnsuccessfulPkts uint
	wasDown               bool // Used to determine the duration of a downtime
	isIP                  bool // When IP is provided instead of a hostname, suppresses printing the IP information twice
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
		return 0, 0, 0, true
	}

	sort.Slice(timeArr, func(i, j int) bool { return timeArr[i] < timeArr[j] })

	return timeArr[0], float32(avgTime) / float32(arrLen), timeArr[arrLen-1], false
}

/* Calculate the correct number of seconds in calcTime func */
func calcSeconds(time float64) string {
	_, float := math.Modf(time)
	secondStr := strconv.FormatFloat(float*60, 'f', 0, 32)

	return secondStr
}

/* Calculate the correct number of minutes in calcTime func */
func calcMinutes(time float64) string {
	return "TODO"
}

/* Calculate the correct number of hours in calcTime func */
func calcHours(time float64) string {
	return "TODO"
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

/* Print stattistics when program exits */
func printStatistics(tcpStats *stats) {
	min, avg, max, IsEmpty := findMinAvgMaxRttTime(tcpStats.rtt)

	if IsEmpty {
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

	/* uptime and downtime stats */
	cy("total uptime: ")
	cg("  %s\n", totalUptime)
	cy("total downtime: ")
	cr("%s\n", totalDowntime)

	/* latency stats.
	TODO: see if formatted string would suit better */
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
	cy("TCPing started at: %v\n", tcpStats.startTime.Format("2006-01-02 15:04:05"))

	/* If the program was not terminated, no need to show the end time */
	if tcpStats.endTime.Format("2006-01-02 15:04:05") == "0001-01-01 00:00:00" {
		durationDiff := time.Now().Sub(tcpStats.startTime)
		duration := time.Time{}.Add(durationDiff)
		cy("Duration Hours:Minutes:Seconds: %v\n", duration.Format("15:04:05"))
	} else {
		cy("TCPing ended at:   %v\n", tcpStats.endTime.Format("2006-01-02 15:04:05"))
		durationDiff := tcpStats.endTime.Sub(tcpStats.startTime)
		duration := time.Time{}.Add(durationDiff)
		cy("Duration Hours:Minutes:Seconds: %v\n", duration.Format("15:04:05"))
	}
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

/* Ping host, TCP style */
func tcping(tcpStats *stats) {

	IPAndPort := net.JoinHostPort(tcpStats.IP, tcpStats.port)

	startTime := time.Now()

	conn, err := net.DialTimeout("tcp", IPAndPort, 1*time.Second)
	defer conn.Close()

	endTime := time.Since(startTime)
	rtt := endTime.Milliseconds()

	if err != nil {
		/* if the previous probe was successful
		and the current one failed: */
		if !tcpStats.wasDown {
			tcpStats.ongoingDowntime = time.Now()
			tcpStats.wasDown = true
		}

		tcpStats.totalDowntime += time.Second
		tcpStats.totalUnsuccessfulPkts++

		printReply(tcpStats, "No reply", 0)
	} else {
		/* if the previous probe failed
		and the current one succeeded: */
		if tcpStats.wasDown {
			/* calculate the total downtime since
			the previous successful probe */
			currentDowntime := time.Since(tcpStats.ongoingDowntime).Seconds()
			calculatedDowntime := calcTime(uint(math.Ceil(currentDowntime)))
			color.Yellow.Printf("No response received for %s\n", calculatedDowntime)

			tcpStats.wasDown = false
		}

		tcpStats.lastSuccessfulProbe = time.Now()
		tcpStats.totalUptime += time.Second
		tcpStats.totalSuccessfulPkts++

		tcpStats.rtt = append(tcpStats.rtt, uint(rtt))
		printReply(tcpStats, "Reply", rtt)
	}

	time.Sleep((1000 * time.Millisecond) - endTime)
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
		case stdin, _ := <-stdinChan:
			if stdin == "\n" || stdin == "\r" || stdin == "\r\n" {
				printStatistics(&tcpStats)
			}
		default:
			continue
		}
	}
}
