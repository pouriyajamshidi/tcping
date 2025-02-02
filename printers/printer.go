// Package printers contains the logic for printing information
package printers

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/types"
)

// Shutdown calculates endTime, prints statistics and calls os.Exit(0).
// This should be used as the main exit-point.
func Shutdown(t *types.Tcping) {
	t.EndTime = time.Now()
	PrintStats(t)

	// if the printer type is `database`, close it before exiting
	if db, ok := t.Printer.(*Database); ok {
		db.Conn.Close()
	}

	// if the printer type is `csvPrinter`, call the cleanup function before exiting
	if cp, ok := t.Printer.(*CsvPrinter); ok {
		cp.Cleanup()
	}

	os.Exit(0)
}

// SignalHandler catches SIGINT and SIGTERM then prints tcping stats
func SignalHandler(tcping *types.Tcping) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		Shutdown(tcping)
	}()
}

// PrintStats is a helper method for printStatistics
// for the current printer.
//
// This should be used instead, as it makes
// all the necessary calculations beforehand.
func PrintStats(t *types.Tcping) {
	if t.DestWasDown {
		CalcLongestDowntime(t, time.Since(t.StartOfDowntime))
	} else {
		CalcLongestUptime(t, time.Since(t.StartOfUptime))
	}
	t.RttResults = calcMinAvgMaxRttTime(t.Rtt)

	t.Printer.PrintStatistics(*t)
}

// CalcLongestUptime calculates the longest uptime and sets it to tcpStats.
func CalcLongestUptime(tcping *types.Tcping, duration time.Duration) {
	if tcping.StartOfUptime.IsZero() || duration == 0 {
		return
	}

	longestUptime := types.NewLongestTime(tcping.StartOfUptime, duration)

	// It means it is the first time we're calling this function
	if tcping.LongestUptime.End.IsZero() {
		tcping.LongestUptime = longestUptime
		return
	}

	if longestUptime.Duration >= tcping.LongestUptime.Duration {
		tcping.LongestUptime = longestUptime
	}
}

// CalcLongestDowntime calculates the longest downtime and sets it to tcpStats.
func CalcLongestDowntime(tcping *types.Tcping, duration time.Duration) {
	if tcping.StartOfDowntime.IsZero() || duration == 0 {
		return
	}

	longestDowntime := types.NewLongestTime(tcping.StartOfDowntime, duration)

	// It means it is the first time we're calling this function
	if tcping.LongestDowntime.End.IsZero() {
		tcping.LongestDowntime = longestDowntime
		return
	}

	if longestDowntime.Duration >= tcping.LongestDowntime.Duration {
		tcping.LongestDowntime = longestDowntime
	}
}

// calcMinAvgMaxRttTime calculates min, avg and max RTT values
func calcMinAvgMaxRttTime(timeArr []float32) types.RttResult {
	var sum float32
	var result types.RttResult

	arrLen := len(timeArr)
	if arrLen > 0 {
		result.Min = timeArr[0]
	}

	for i := 0; i < arrLen; i++ {
		sum += timeArr[i]

		if timeArr[i] > result.Max {
			result.Max = timeArr[i]
		}

		if timeArr[i] < result.Min {
			result.Min = timeArr[i]
		}
	}

	if arrLen > 0 {
		result.HasResults = true
		result.Average = sum / float32(arrLen)
	}

	return result
}
