package probers

import (
	"net"
	"net/netip"
	"os"
	"os/signal"
	"syscall"
	"time"
)

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

// Probe pings a target using TCP
func Probe(tcping *tcping) {
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
