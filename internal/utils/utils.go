// Package utils hosts common utilities used throughout the code
package utils

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/pouriyajamshidi/tcping/v2/consts"
	"github.com/pouriyajamshidi/tcping/v2/types"
)

// MonitorSTDIN checks stdin to see whether the 'Enter' key was pressed
func MonitorSTDIN(stdinChan chan<- bool) {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')

		if input == "\n" || input == "\r" || input == "\r\n" {
			stdinChan <- true
		}
	}
}

// Usage prints how tcping should be run
func Usage() {
	executableName := os.Args[0]

	consts.ColorLightCyan("\nTCPING version %s\n\n", consts.Version)
	consts.ColorRed("Try running %s like:\n", executableName)
	consts.ColorRed("%s <hostname/ip> <port number>. For example:\n", executableName)
	consts.ColorRed("%s www.example.com 443\n", executableName)
	consts.ColorYellow("\n[optional flags]\n")

	flag.VisitAll(func(f *flag.Flag) {
		flagName := f.Name
		if len(f.Name) > 1 {
			flagName = "-" + flagName
		}

		consts.ColorYellow("  -%s : %s\n", flagName, f.Usage)
	})

	os.Exit(1)
}

// compareVersions is used to compare tcping versions
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		n1, _ := strconv.Atoi(parts1[i])
		n2, _ := strconv.Atoi(parts2[i])

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	// for cases in which version numbers differ in length
	if len(parts1) < len(parts2) {
		return -1
	}

	if len(parts1) > len(parts2) {
		return 1
	}

	return 0
}

// ShowVersion displays the version and exits
func ShowVersion() {
	consts.ColorGreen("TCPING version %s\n", consts.Version)
	os.Exit(0)
}

// CheckForUpdates checks for newer versions of tcping
func CheckForUpdates() {
	c := github.NewClient(nil)

	/* unauthenticated requests from the same IP are limited to 60 per hour. */
	latestRelease, _, err := c.Repositories.GetLatestRelease(context.Background(), consts.Owner, consts.Repo)
	if err != nil {
		consts.ColorRed("Failed to check for updates %s\n", err.Error())
		os.Exit(1)
	}

	reg := `^v?(\d+\.\d+\.\d+)$`
	latestTagName := latestRelease.GetTagName()
	latestVersion := regexp.MustCompile(reg).FindStringSubmatch(latestTagName)

	if len(latestVersion) == 0 {
		consts.ColorRed("Failed to check for updates. The version name does not match the rule: %s\n", latestTagName)
		os.Exit(1)
	}

	comparison := compareVersions(consts.Version, latestVersion[1])

	if comparison < 0 {
		consts.ColorCyan("Found newer version %s\n", latestVersion[1])
		consts.ColorCyan("Please update TCPING from the URL below:\n")
		consts.ColorCyan("https://github.com/%s/%s/releases/tag/%s\n",
			consts.Owner, consts.Repo, latestTagName)
	} else if comparison > 0 {
		consts.ColorCyan("Current version %s is newer than the latest release %s\n",
			consts.Version, latestVersion[1])
	} else {
		consts.ColorCyan("You have the latest version: %s\n", consts.Version)
	}

	os.Exit(0)
}

// NanoToMillisecond returns an amount of milliseconds from nanoseconds.
// Using duration.Milliseconds() is not an option, because it drops
// decimal points, returning an int.
func NanoToMillisecond(nano int64) float32 {
	return float32(nano) / float32(time.Millisecond)
}

// SecondsToDuration returns the corresponding duration from seconds expressed with a float.
func SecondsToDuration(seconds float64) time.Duration {
	return time.Duration(1000*seconds) * time.Millisecond
}

// MaxDuration is the implementation of the math.Max function for time.Duration types.
// returns the longest duration of x or y.
func MaxDuration(x, y time.Duration) time.Duration {
	if x > y {
		return x
	}
	return y
}

// DurationToString creates a human-readable string for a given duration
func DurationToString(duration time.Duration) string {
	hours := math.Floor(duration.Hours())
	if hours > 0 {
		duration -= time.Duration(hours * float64(time.Hour))
	}

	minutes := math.Floor(duration.Minutes())
	if minutes > 0 {
		duration -= time.Duration(minutes * float64(time.Minute))
	}

	seconds := duration.Seconds()

	switch {
	// Hours
	case hours >= 2:
		return fmt.Sprintf("%.0f hours %.0f minutes %.0f seconds", hours, minutes, seconds)
	case hours == 1 && minutes == 0 && seconds == 0:
		return fmt.Sprintf("%.0f hour", hours)
	case hours == 1:
		return fmt.Sprintf("%.0f hour %.0f minutes %.0f seconds", hours, minutes, seconds)

	// Minutes
	case minutes >= 2:
		return fmt.Sprintf("%.0f minutes %.0f seconds", minutes, seconds)
	case minutes == 1 && seconds == 0:
		return fmt.Sprintf("%.0f minute", minutes)
	case minutes == 1:
		return fmt.Sprintf("%.0f minute %.0f seconds", minutes, seconds)

	// Seconds
	case seconds == 0 || seconds == 1 || seconds >= 1 && seconds < 1.1:
		return fmt.Sprintf("%.0f second", seconds)
	case seconds < 1:
		return fmt.Sprintf("%.1f seconds", seconds)

	default:
		return fmt.Sprintf("%.0f seconds", seconds)
	}
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
