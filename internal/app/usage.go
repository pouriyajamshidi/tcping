package app

import (
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v45/github"
)

// Version is set at compile time
var Version = ""

const (
	Owner = "pouriyajamshidi"
	Repo  = "tcping"
)

// PrintUsage prints how tcping should be run
func PrintUsage() {
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
}

func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := range min(len(parts1), len(parts2)) {
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

// PrintVersion displays the version
func PrintVersion() {
	fmt.Printf("TCPING version %s\n", Version)
}

// CheckForUpdates checks for newer versions of tcping and returns update message
func CheckForUpdates() (string, error) {
	c := github.NewClient(nil)

	// unauthenticated requests from the same IP are limited to 60 per hour
	latestRelease, _, err := c.Repositories.GetLatestRelease(context.Background(), Owner, Repo)
	if err != nil {
		return "", fmt.Errorf("check for updates: %w", err)
	}

	reg := `^v?(\d+\.\d+\.\d+)$`
	latestTagName := latestRelease.GetTagName()
	latestVersion := regexp.MustCompile(reg).FindStringSubmatch(latestTagName)

	if len(latestVersion) == 0 {
		return "", fmt.Errorf("version name does not match expected format: %s", latestTagName)
	}

	comparison := compareVersions(Version, latestVersion[1])

	switch comparison {
	case -1:
		return fmt.Sprintf("Found newer version %s\nPlease update TCPING from the URL below:\nhttps://github.com/%s/%s/releases/tag/%s",
			latestVersion[1], Owner, Repo, latestTagName), nil
	case 1:
		return fmt.Sprintf("Current version %s is newer than the latest release %s",
			Version, latestVersion[1]), nil
	case 0:
		return fmt.Sprintf("TCPING is on the latest version: %s", Version), nil
	}

	return "", fmt.Errorf("unexpected version comparison result")
}
