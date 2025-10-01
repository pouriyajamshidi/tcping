// Package utils hosts common utilities used throughout the code
package usage

import (
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v45/github"
)

// Version is set at compile time
var Version = ""

// Used when checking for updates
const (
	Owner = "pouriyajamshidi"
	Repo  = "tcping"
)

// Usage prints how tcping should be run
func Usage() {
	executableName := os.Args[0]

	fmt.Printf("\nTCPING version %s\n\n", Version)
	fmt.Printf("Try running %s like:\n", executableName)
	fmt.Printf("%s <hostname/ip> <port number>. For example:\n", executableName)
	fmt.Printf("%s www.example.com 443\n", executableName)
	fmt.Printf("\n[optional flags]\n")

	flag.VisitAll(func(f *flag.Flag) {
		flagName := f.Name
		if len(f.Name) > 1 {
			flagName = "-" + flagName
		}

		fmt.Printf("  -%s : %s\n", flagName, f.Usage)
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
	fmt.Printf("TCPING version %s\n", Version)
	os.Exit(0)
}

// CheckForUpdates checks for newer versions of tcping
func CheckForUpdates() {
	c := github.NewClient(nil)

	/* unauthenticated requests from the same IP are limited to 60 per hour. */
	latestRelease, _, err := c.Repositories.GetLatestRelease(context.Background(), Owner, Repo)
	if err != nil {
		fmt.Printf("Failed to check for updates %s\n", err.Error())
		os.Exit(1)
	}

	reg := `^v?(\d+\.\d+\.\d+)$`
	latestTagName := latestRelease.GetTagName()
	latestVersion := regexp.MustCompile(reg).FindStringSubmatch(latestTagName)

	if len(latestVersion) == 0 {
		fmt.Printf("Failed to check for updates. The version name does not match the rule: %s\n", latestTagName)
		os.Exit(1)
	}

	comparison := compareVersions(Version, latestVersion[1])

	switch comparison {
	case -1:
		fmt.Printf("Found newer version %s\n", latestVersion[1])
		fmt.Printf("Please update TCPING from the URL below:\n")
		fmt.Printf("https://github.com/%s/%s/releases/tag/%s\n",
			Owner, Repo, latestTagName)
	case 1:
		fmt.Printf("Current version %s is newer than the latest release %s\n",
			Version, latestVersion[1])
	case 0:
		fmt.Printf("TCPING is on the latest version: %s\n", Version)
	}

	os.Exit(0)
}

// MaxDuration is the implementation of the math.Max function for time.Duration statistics.
// returns the longest duration of x or y.
func MaxDuration(x, y time.Duration) time.Duration {
	if x > y {
		return x
	}
	return y
}
