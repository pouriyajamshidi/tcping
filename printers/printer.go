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
	if cp, ok := t.Printer.(*CSVPrinter); ok {
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

// SetLongestDuration updates the longest uptime or downtime based on the given type.
func SetLongestDuration(start time.Time, duration time.Duration, longest *types.LongestTime) {
	if start.IsZero() || duration == 0 {
		return
	}

	newLongest := types.NewLongestTime(start, duration)

	if longest.End.IsZero() || newLongest.Duration >= longest.Duration {
		*longest = newLongest
	}
}

// PrintStats is a helper method for PrintStatistics
// for the current printer.
// This should be used instead, as it makes
// all the necessary calculations beforehand.
func PrintStats(t *types.Tcping) {
	if t.DestWasDown {
		SetLongestDuration(t.StartOfDowntime, time.Since(t.StartOfDowntime), &t.LongestDowntime)
	} else {
		SetLongestDuration(t.StartOfUptime, time.Since(t.StartOfUptime), &t.LongestUptime)
	}

	t.RttResults = calcMinAvgMaxRttTime(t.Rtt)

	t.PrintStatistics(*t)
}
