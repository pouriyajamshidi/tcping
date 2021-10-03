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
	hostname        string
	IP              string
	port            string
	totalSucPkts    uint
	totalUnsucPkts  uint
	totalUptime     time.Duration
	totalDowntime   time.Duration
	ongoingDowntime time.Time
	lastSucProbe    time.Time
	rtt             []uint
	wasDown         bool
	isIP            bool // used to not duplicate IP info when printing reply results
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
func findMinAvgMaxTime(timeArr []uint) (uint, float32, uint, bool) {

	var avgTime float32
	arrLen := len(timeArr)

	for i := 0; i < arrLen; i++ {
		if timeArr[i] == 0 {
			continue
		}
		avgTime += float32(timeArr[i])
	}

	if avgTime == 0 {
		/* prevents panics inside printStatistics func */
		return 0, 0, 0, true
	}

	sort.Slice(timeArr, func(i, j int) bool { return timeArr[i] < timeArr[j] })

	return timeArr[0], avgTime / float32(arrLen), timeArr[arrLen-1], false
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

	totalPackets := tcpStats.totalSucPkts + tcpStats.totalUnsucPkts
	totalUptime := calcTime(uint(tcpStats.totalUptime.Seconds()))
	totalDowntime := calcTime(uint(tcpStats.totalDowntime.Seconds()))

	min, avg, max, empty := findMinAvgMaxTime(tcpStats.rtt)

	if empty {
		/* There are no results to show */
		return
	}

	packetLoss := (float32(tcpStats.totalUnsucPkts) / float32(totalPackets)) * 100

	cy := color.Yellow.Printf
	cg := color.Green.Printf
	cr := color.Red.Printf
	cc := color.Cyan.Printf
	cly := color.LightYellow.Printf

	/* general stats */
	cy("\n--- %s TCPing statistics ---\n", tcpStats.hostname)
	cy("%d probes transmitted, ", totalPackets)
	cy("%d received, ", tcpStats.totalSucPkts)

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
	cg("%d\n", tcpStats.totalSucPkts)

	/* unsuccessful packet stats */
	cy("unsuccessful probes: ")
	cr("%d\n", tcpStats.totalUnsucPkts)

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
}

/* Print TCP probe replies according to our policies */
func printReply(tcpStats *stats, senderMsg string, rtt int64) {
	// TODO: Refactor

	cr := color.Red.Printf
	cg := color.LightGreen.Printf

	if tcpStats.isIP {
		if senderMsg == "No reply" {
			cr("%s from %s on port %s TCP_conn=%d\n",
				senderMsg, tcpStats.IP, tcpStats.port, tcpStats.totalUnsucPkts)
		} else {
			cg("%s from %s on port %s TCP_conn=%d time=%d ms\n",
				senderMsg, tcpStats.IP, tcpStats.port, tcpStats.totalSucPkts, rtt)
		}
	} else {
		if senderMsg == "No reply" {
			cr("%s from %s (%s) on port %s TCP_conn=%d\n",
				senderMsg, tcpStats.hostname, tcpStats.IP, tcpStats.port, tcpStats.totalUnsucPkts)
		} else {
			cg("%s from %s (%s) on port %s TCP_conn=%d time=%d ms\n",
				senderMsg, tcpStats.hostname, tcpStats.IP, tcpStats.port, tcpStats.totalSucPkts, rtt)
		}
	}
}

/* Ping host, TCP style */
func tcping(tcpStats *stats) {

	IPAndPort := net.JoinHostPort(tcpStats.IP, tcpStats.port)

	startTime := time.Now()

	conn, err := net.DialTimeout("tcp", IPAndPort, 1*time.Second)

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
		tcpStats.totalUnsucPkts++

		printReply(tcpStats, "No reply", rtt)

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

		tcpStats.lastSucProbe = time.Now()

		tcpStats.rtt = append(tcpStats.rtt, uint(rtt))

		tcpStats.totalUptime += time.Second
		tcpStats.totalSucPkts++

		printReply(tcpStats, "Reply", rtt)

		conn.Close()
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
