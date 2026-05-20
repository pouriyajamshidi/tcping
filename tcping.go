// tcping.go probes a target using TCP
package main

import (
	"bufio"
	"net"
	"net/netip"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/printers"
)

var version = "" // set at compile time through the Makefile

const (
	owner      = "pouriyajamshidi"
	repo       = "tcping"
	dnsTimeout = 2 * time.Second
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
	printProbeSuccess(sourceAddr string, userInput userInput, streak uint, rtt float32)

	// printProbeFail should print a message after each failed probe.
	// hostname could be empty, meaning it's pinging an address.
	// streak is the number of successful consecutive probes.
	printProbeFail(userInput userInput, streak uint)

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
	printStatistics(s tcping)

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

type tcping struct {
	printer                   // printer holds the chosen printer implementation for outputting information and data.
	startTime                 time.Time
	endTime                   time.Time
	startOfUptime             time.Time
	startOfDowntime           time.Time
	lastSuccessfulProbe       time.Time
	lastUnsuccessfulProbe     time.Time
	ticker                    *time.Ticker // ticker is used to handle time between probes.
	longestUptime             longestTime
	longestDowntime           longestTime
	rtt                       []float32
	hostnameChanges           []hostnameChange
	userInput                 userInput
	ongoingSuccessfulProbes   uint
	ongoingUnsuccessfulProbes uint
	totalDowntime             time.Duration
	totalUptime               time.Duration
	totalSuccessfulProbes     uint
	totalUnsuccessfulProbes   uint
	retriedHostnameLookups    uint
	rttResults                rttResult
	destWasDown               bool // destWasDown is used to determine the duration of a downtime
	destIsIP                  bool // destIsIP suppresses printing the IP information twice when hostname is not provided
}

type longestTime struct {
	start    time.Time
	end      time.Time
	duration time.Duration
}

type rttResult struct {
	min        float32
	max        float32
	average    float32
	hasResults bool
}

type hostnameChange struct {
	Addr netip.Addr `json:"addr"`
	When time.Time  `json:"when"`
}

// signalHandler catches SIGINT and SIGTERM then prints tcping stats
func signalHandler(tcping *tcping) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		shutdown(tcping)
	}()
}

// monitorSTDIN checks stdin to see whether the 'Enter' key was pressed
func monitorSTDIN(stdinChan chan bool) {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')

		if input == "\n" || input == "\r" || input == "\r\n" {
			stdinChan <- true
		}
	}
}

// printStats is a helper method for printStatistics
// for the current printer.
//
// This should be used instead, as it makes
// all the necessary calculations beforehand.
func (t *tcping) printStats() {
	if t.destWasDown {
		calcLongestDowntime(t, time.Since(t.startOfDowntime))
	} else {
		calcLongestUptime(t, time.Since(t.startOfUptime))
	}
	t.rttResults = calcMinAvgMaxRttTime(t.rtt)

	t.printStatistics(*t)
}

// shutdown calculates endTime, prints statistics and calls os.Exit(0).
// This should be used as the main exit-point.
func shutdown(tcping *tcping) {
	tcping.endTime = time.Now()
	tcping.printStats()

	// if the printer type is `database`, close it before exiting
	if db, ok := tcping.printer.(*printers.Database); ok {
		db.conn.Close()
	}

	// if the printer type is `csvPrinter`, call the cleanup function before exiting
	if cp, ok := tcping.printer.(*printers.CSVPrinter); ok {
		cp.cleanup()
	}

	os.Exit(0)
}

// newLongestTime creates LongestTime structure
func newLongestTime(startTime time.Time, duration time.Duration) longestTime {
	return longestTime{
		start:    startTime,
		end:      startTime.Add(duration),
		duration: duration,
	}
}

// calcMinAvgMaxRttTime calculates min, avg and max RTT values
func calcMinAvgMaxRttTime(timeArr []float32) rttResult {
	var sum float32
	var result rttResult

	arrLen := len(timeArr)
	// rttResults.min = ^uint(0.0)
	if arrLen > 0 {
		result.min = timeArr[0]
	}

	for i := range arrLen {
		sum += timeArr[i]

		if timeArr[i] > result.max {
			result.max = timeArr[i]
		}

		if timeArr[i] < result.min {
			result.min = timeArr[i]
		}
	}

	if arrLen > 0 {
		result.hasResults = true
		result.average = sum / float32(arrLen)
	}

	return result
}

// calcLongestUptime calculates the longest uptime and sets it to tcpStats.
func calcLongestUptime(tcping *tcping, duration time.Duration) {
	if tcping.startOfUptime.IsZero() || duration == 0 {
		return
	}

	longestUptime := newLongestTime(tcping.startOfUptime, duration)

	// It means it is the first time we're calling this function
	if tcping.longestUptime.end.IsZero() {
		tcping.longestUptime = longestUptime
		return
	}

	if longestUptime.duration >= tcping.longestUptime.duration {
		tcping.longestUptime = longestUptime
	}
}

// calcLongestDowntime calculates the longest downtime and sets it to tcpStats.
func calcLongestDowntime(tcping *tcping, duration time.Duration) {
	if tcping.startOfDowntime.IsZero() || duration == 0 {
		return
	}

	longestDowntime := newLongestTime(tcping.startOfDowntime, duration)

	// It means it is the first time we're calling this function
	if tcping.longestDowntime.end.IsZero() {
		tcping.longestDowntime = longestDowntime
		return
	}

	if longestDowntime.duration >= tcping.longestDowntime.duration {
		tcping.longestDowntime = longestDowntime
	}
}

// nanoToMillisecond returns an amount of milliseconds from nanoseconds.
// Using duration.Milliseconds() is not an option, because it drops
// decimal points, returning an int.
func nanoToMillisecond(nano int64) float32 {
	return float32(nano) / float32(time.Millisecond)
}

// secondsToDuration returns the corresponding duration from seconds expressed with a float.
func secondsToDuration(seconds float64) time.Duration {
	return time.Duration(1000*seconds) * time.Millisecond
}

// maxDuration is the implementation of the math.Max function for time.Duration types.
// returns the longest duration of x or y.
func maxDuration(x, y time.Duration) time.Duration {
	if x > y {
		return x
	}
	return y
}

// handleConnError processes failed probes
func (t *tcping) handleConnError(connTime time.Time, elapsed time.Duration) {
	if !t.destWasDown {
		t.startOfDowntime = connTime
		uptime := t.startOfDowntime.Sub(t.startOfUptime)
		calcLongestUptime(t, uptime)
		t.startOfUptime = time.Time{}
		t.destWasDown = true
	}

	t.totalDowntime += elapsed
	t.lastUnsuccessfulProbe = connTime
	t.totalUnsuccessfulProbes++
	t.ongoingUnsuccessfulProbes++

	t.printProbeFail(
		t.userInput,
		t.ongoingUnsuccessfulProbes,
	)
}

// handleConnSuccess processes successful probes
func (t *tcping) handleConnSuccess(sourceAddr string, rtt float32, connTime time.Time, elapsed time.Duration) {
	if t.destWasDown {
		t.startOfUptime = connTime
		downtime := t.startOfUptime.Sub(t.startOfDowntime)
		calcLongestDowntime(t, downtime)
		t.printTotalDownTime(downtime)
		t.startOfDowntime = time.Time{}
		t.destWasDown = false
		t.ongoingUnsuccessfulProbes = 0
		t.ongoingSuccessfulProbes = 0
	}

	if t.startOfUptime.IsZero() {
		t.startOfUptime = connTime
	}

	t.totalUptime += elapsed
	t.lastSuccessfulProbe = connTime
	t.totalSuccessfulProbes++
	t.ongoingSuccessfulProbes++
	t.rtt = append(t.rtt, rtt)

	if !t.userInput.showFailuresOnly {
		t.printProbeSuccess(
			sourceAddr,
			t.userInput,
			t.ongoingSuccessfulProbes,
			rtt,
		)
	}
}

// tcpProbe pings a host, TCP style
func tcpProbe(tcping *tcping) {
	var err error
	var conn net.Conn
	connStart := time.Now()

	if tcping.userInput.networkInterface.use {
		// dialer already contains the timeout value
		conn, err = tcping.userInput.networkInterface.dialer.Dial("tcp", tcping.userInput.networkInterface.remoteAddr.String())
	} else {
		ipAndPort := netip.AddrPortFrom(tcping.userInput.ip, tcping.userInput.port)
		conn, err = net.DialTimeout("tcp", ipAndPort.String(), tcping.userInput.timeout)
	}

	connDuration := time.Since(connStart)
	rtt := nanoToMillisecond(connDuration.Nanoseconds())

	elapsed := maxDuration(connDuration, tcping.userInput.intervalBetweenProbes)

	if err != nil {
		tcping.handleConnError(connStart, elapsed)
	} else {
		tcping.handleConnSuccess(conn.LocalAddr().String(), rtt, connStart, elapsed)
		conn.Close()
	}
	<-tcping.ticker.C
}

func main() {
	tcping := &tcping{}
	processUserInput(tcping)
	tcping.ticker = time.NewTicker(tcping.userInput.intervalBetweenProbes)
	defer tcping.ticker.Stop()

	signalHandler(tcping)

	tcping.printStart(tcping.userInput.hostname, tcping.userInput.port)

	stdinchan := make(chan bool)
	if !tcping.userInput.nonInteractive {
		go monitorSTDIN(stdinchan)
	}

	var probeCount uint
	for {
		if tcping.userInput.shouldRetryResolve {
			retryResolveHostname(tcping)
		}

		tcpProbe(tcping)

		select {
		case pressedEnter := <-stdinchan:
			if pressedEnter {
				tcping.printStats()
			}
		default:
		}

		if tcping.userInput.probesBeforeQuit != 0 {
			probeCount++
			if probeCount == tcping.userInput.probesBeforeQuit {
				shutdown(tcping)
			}
		}
	}
}
