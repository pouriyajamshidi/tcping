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
	totalSucPackets   uint
	totalUnsucPackets uint
	successCounter    uint
	failureCounter    uint
	totalUptime       uint
	totalDowntime     uint
	timeArray         []uint
	wasDown           bool
	downTime          time.Time
}

/* Print how program should be run */
func usage() {
	color.Red.Printf("Try running %s like:\n", os.Args[0])
	color.Red.Printf("%s HOSTNAME/IP PORT-NUMBER... example:\n", os.Args[0])
	color.Red.Printf("%s www.example.com 443\n", os.Args[0])
	os.Exit(1)
}

/* Catch SIGINT and print a summary of stats */
func signalHandler(host string, tcpStats *stats) {
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		printResults(host, tcpStats)
		os.Exit(0)
	}()
}

/* Validate user input */
func validateInput() (string, string, string) {
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

	IPaddr := resolveHostname(host)
	IP := IPaddr[0].String()

	return host, port, IP
}

/* Hostname resolution */
func resolveHostname(host string) []net.IP {
	IPaddr, err := net.LookupIP(host)

	if err != nil {
		color.Red.Printf("Failed to resolve %s\n", host)
		os.Exit(1)
	}
	return IPaddr
}

/* find min/avg/max RTT values */
func findMinAvgMaxTime(x []uint) (uint, float32, uint, int) {

	var avgTime float32 = 0
	// var placeHolder uint

	arrLen := len(x)

	for i := 0; i < arrLen; i++ {
		if x[i] == 0 {
			continue
		}
		avgTime += float32(x[i])
	}

	if avgTime == 0 {
		return 0, 0, 0, -1
	}

	sort.Slice(x, func(i, j int) bool { return x[i] < x[j] })

	return x[0], avgTime / float32(arrLen), x[arrLen-1], 0
}

/* Calculate the correct number of seconds in formatTime func */
func calcSeconds(time float64) string {
	_, float := math.Modf(time)
	seconds := float * 60
	timeSec := strconv.FormatFloat(seconds, 'f', 0, 32)

	return timeSec
}

/* Calculate cumulative time */
func formatTime(time uint) string {
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
			timeStr = strconv.FormatFloat(timeFloat, 'f', 0, 32) + " minute"
			return timeStr
		} else if timeFloat < 2.00 {
			timeMin := int(timeFloat)
			timeSec := calcSeconds(timeFloat)
			timeStr := strconv.Itoa(timeMin) + "." + timeSec + " minute.seconds"
			return timeStr
		}

		timeMin := int(timeFloat)
		timeSec := calcSeconds(timeFloat)
		timeStr := strconv.Itoa(timeMin) + "." + timeSec + " minute.seconds"

		return timeStr
	}
}

/* Print stattistics when program exits */
func printResults(host string, tcpStats *stats) {

	totalPackets := tcpStats.totalSucPackets + tcpStats.totalUnsucPackets
	totaluptime := formatTime(tcpStats.totalSucPackets)
	totaldowntime := formatTime(tcpStats.totalUnsucPackets)
	min, avg, max, err := findMinAvgMaxTime(tcpStats.timeArray)

	if err == -1 {
		return
	}

	packetLoss := (float32(tcpStats.totalUnsucPackets) / float32(totalPackets)) * 100

	color.Yellow.Printf("\n--- %s TCPing statistics ---\n", host)

	color.Yellow.Printf("%d probes transmitted, ", totalPackets)
	color.Yellow.Printf("%d received, ", tcpStats.totalSucPackets)

	if packetLoss == 0 {
		color.Green.Printf("%.2f%%", packetLoss)
	} else if packetLoss > 0 && packetLoss <= 30 {
		color.LightYellow.Printf("%.2f%%", packetLoss)
	} else {
		color.Red.Printf("%.2f%%", packetLoss)
	}
	color.Yellow.Printf(" packet loss\n")

	color.Yellow.Printf("successful probes:   ")
	color.Green.Printf("%d\n", tcpStats.totalSucPackets)
	color.Yellow.Printf("unsuccessful probes: ")
	color.Red.Printf("%d\n", tcpStats.totalUnsucPackets)

	color.Yellow.Printf("total uptime: ")
	color.Green.Printf("  %s\n", totaluptime)
	color.Yellow.Printf("total downtime: ")
	color.Red.Printf("%s\n", totaldowntime)

	/* otherwise will be buggy on Windows */
	color.Yellow.Printf("rtt ")
	color.Green.Printf("min")
	color.Yellow.Printf("/")
	color.Cyan.Printf("avg")
	color.Yellow.Printf("/")
	color.Red.Printf("max: ")
	color.Green.Printf("%d", min)
	color.Yellow.Printf("/")
	color.Cyan.Printf("%.2f", avg)
	color.Yellow.Printf("/")
	color.Red.Printf("%d", max)
	color.Yellow.Printf(" ms\n")
}

/* Ping host, TCP style */
func tcping(host string, port string, IP string, tcpStats *stats) {

	hostAndPort := net.JoinHostPort(IP, port)

	startTime := time.Now()

	conn, err := net.DialTimeout("tcp", hostAndPort, 1*time.Second)

	timeDiff := time.Since(startTime)
	timeDiffInMilSec := timeDiff.Milliseconds()

	if err != nil {

		if !tcpStats.wasDown {
			tcpStats.downTime = time.Now()
			tcpStats.wasDown = true
		}

		downtime := uint(time.Since(startTime) / time.Second)
		tcpStats.totalDowntime += downtime

		tcpStats.failureCounter++
		tcpStats.totalUnsucPackets++

		color.Red.Printf("No reply from %s (%s) on port %s TCP_seq=%d\n",
			host, IP, port, tcpStats.failureCounter)

	} else {

		if tcpStats.wasDown {

			downtime := uint(time.Since(tcpStats.downTime) / time.Second)
			tcpStats.totalDowntime += downtime

			if downtime == 1 {
				color.Yellow.Printf("No response received for %d second\n", downtime)
			} else if downtime < 60 {
				color.Yellow.Printf("No response received for %d seconds\n", downtime)
			} else {
				color.Yellow.Printf("No response received for %d minutes\n", downtime/60)
			}

			tcpStats.wasDown = false
		}

		tcpStats.timeArray = append(tcpStats.timeArray, uint(timeDiffInMilSec))

		tcpStats.successCounter++
		tcpStats.totalSucPackets++

		color.LightGreen.Printf("Reply from %s (%s) on port %s TCP_seq=%d time=%d ms\n",
			host, IP, port, tcpStats.successCounter, timeDiff.Milliseconds())

		conn.Close()
	}
	time.Sleep(1 * time.Second)
}

func main() {

	host, port, IP := validateInput()

	var tcpStats stats
	tcpStatsAdrr := &tcpStats

	signalHandler(host, tcpStatsAdrr)

	color.LightCyan.Printf("TCPinging %s on port %s\n", host, port)

	channel := make(chan string)

	go func(channel chan string) {
		reader := bufio.NewReader(os.Stdin)
		for {
			key, _ := reader.ReadString('\n')
			channel <- key
		}
	}(channel)

	for {
		tcping(host, port, IP, tcpStatsAdrr)

		select {
		case stdin, _ := <-channel:
			if stdin == "\n" || stdin == "\r" || stdin == "\r\n" {
				printResults(host, tcpStatsAdrr)
			}
		default:
			continue
		}
	}
}
