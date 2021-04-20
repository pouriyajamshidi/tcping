package main

import (
	"bufio"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gookit/color"
)

var (
	minTime           uint
	aveTime           uint
	maxTime           uint
	totalSucPackets   uint
	totalUnsucPackets uint
	totalDowntime     uint
	totalUptime       uint
	timeArray         []uint
	successCounter    uint = 0
	failureCounter    uint = 0
	wasDown           bool = false
	downTime          time.Time
)

/* Print how program should be run */
func usage() {
	color.Red.Printf("Invalid input. Try running %s like:\n", os.Args[0])
	color.Red.Printf("%s HOSTNAME/IP PORT-NUMBER... example:\n", os.Args[0])
	color.Red.Printf("%s www.example.com 443\n", os.Args[0])
	os.Exit(1)
}

/* Catch SIGINT and print a summary of stats */
func signalHandler(host string) {
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		printResults(host)
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
	var placeHolder uint

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

	for i := 0; i < arrLen-1; i++ {

		for j := 0; j < arrLen-i-1; j++ {

			if x[j] > x[j+1] {
				placeHolder = x[j+1]
				x[j+1] = x[j]
				x[j] = placeHolder
			}
		}
	}

	return x[0], avgTime / float32(arrLen), x[arrLen-1], 0
}

/* Calculate cumulative time */
func formatTime(time uint) string {
	var timestr string

	if time == 1 {
		timestr = strconv.FormatUint(uint64(time), 10) + " second"
		return timestr
	} else if time < 60 {
		timestr = strconv.FormatUint(uint64(time), 10) + " seconds"
		return timestr
	} else {
		timeFloat := float64(time) / 60

		if timeFloat < 2 {
			timestr = strconv.FormatFloat(timeFloat, 'f', 2, 32) + " second"
			return timestr
		} else if timeFloat < 60 {
			timestr = strconv.FormatFloat(timeFloat, 'f', 2, 32) + " seconds"
			return timestr
		}
		timestr = strconv.FormatFloat(timeFloat, 'f', 2, 32) + " minutes.seconds"
		return timestr
	}
}

/* Print stattistics when program exits */
func printResults(host string) {

	totaluptime := formatTime(totalSucPackets)
	totaldowntime := formatTime(totalUnsucPackets)
	min, avg, max, err := findMinAvgMaxTime(timeArray)

	if err == -1 {
		return
	}

	totalPackets := totalSucPackets + totalUnsucPackets
	packetLoss := (float32(totalUnsucPackets) / float32(totalPackets)) * 100

	color.Yellow.Printf("\n--- %s TCPing statistics ---\n", host)

	color.Yellow.Printf("%d probes transmitted, ", totalPackets)
	color.Yellow.Printf("%d received, ", totalSucPackets)

	if packetLoss == 0 {
		color.Green.Printf("%.2f%%", packetLoss)
	} else if packetLoss > 0 && packetLoss <= 30 {
		color.LightYellow.Printf("%.2f%%", packetLoss)
	} else {
		color.Red.Printf("%.2f%%", packetLoss)
	}
	color.Yellow.Printf(" packet loss\n")

	color.Yellow.Printf("successful probes:   ")
	color.Green.Printf("%d\n", totalSucPackets)
	color.Yellow.Printf("unsuccessful probes: ")
	color.Red.Printf("%d\n", totalUnsucPackets)

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
func tcping(host string, port string, IP string) {

	hostAndPort := net.JoinHostPort(IP, port)

	startTime := time.Now()

	conn, err := net.DialTimeout("tcp", hostAndPort, 1*time.Second)

	timeDiff := time.Since(startTime)
	timeDiffInMilSec := timeDiff.Milliseconds()

	if err != nil {

		if !wasDown {
			downTime = time.Now()
			wasDown = true
		}

		downtime := uint(time.Since(startTime) / time.Second)
		totalDowntime += downtime

		failureCounter++
		totalUnsucPackets++

		color.Red.Printf("No reply from %s (%s) on port %s TCP_seq=%d\n",
			host, IP, port, failureCounter)

	} else {

		if wasDown {

			downtime := uint(time.Since(downTime) / time.Second)
			totalDowntime += downtime

			if downtime == 1 {
				color.Yellow.Printf("No response received for %d second\n", downtime)
			} else if downtime < 60 {
				color.Yellow.Printf("No response received for %d seconds\n", downtime)
			} else {
				color.Yellow.Printf("No response received for %d minutes\n", downtime/60)
			}

			wasDown = false
		}

		timeArray = append(timeArray, uint(timeDiffInMilSec))

		successCounter++
		totalSucPackets++

		color.LightGreen.Printf("Reply from %s (%s) on port %s TCP_seq=%d time=%d ms\n",
			host, IP, port, successCounter, timeDiff.Milliseconds())

		conn.Close()
	}
	time.Sleep(1 * time.Second)
}

func main() {

	host, port, IP := validateInput()

	signalHandler(host)

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
		tcping(host, port, IP)

		select {
		case stdin, _ := <-channel:
			if stdin == "\n" || stdin == "\r" || stdin == "\r\n" {
				printResults(host)
			}
		default:
			continue
		}
	}
}
