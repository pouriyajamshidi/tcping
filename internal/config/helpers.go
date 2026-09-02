package config

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/consts"
)

// version is set at compile time via the Makefile
var version = "beta"

// Used when checking for updates
const (
	owner = "pouriyajamshidi"
	repo  = "tcping"
)

// convertAndValidatePort validates and returns the TCP/UDP port
func convertAndValidatePort(port string) (uint16, error) {
	parsedPort, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid port number %q", port)
	}

	if parsedPort == 0 {
		return 0, fmt.Errorf("port should be in 1..65535 range")
	}

	return uint16(parsedPort), nil
}

// parseHostPortArgs handles both "host port" and "host:port" formats
func parseHostPortArgs(args []string) (host string, port string) {
	if len(args) == 1 {
		// We have `host:port`
		if h, p, err := net.SplitHostPort(args[0]); err == nil {
			return h, p
		}
		return args[0], ""
	}

	// We were given `host port`
	return args[0], args[1]
}

// probeTarget is what the user asked us to probe, extracted from the
// positional arguments.
type probeTarget struct {
	hostname string
	port     string
	protocol consts.Protocol
	url      string // Full URL, HTTP(S) targets only.
}

// parseTarget picks the protocol from the user input: an
// http:// or https:// prefix means an HTTP probe, anything else is the plain
// TCP probe tcping has always done. default port is
// (80 or 443) but can still be overridden by the URL's own port or
// a trailing port argument.
func parseTarget(args []string) (probeTarget, error) {
	if len(args) == 0 {
		return probeTarget{}, nil
	}

	protocol, isURL := schemeProtocol(args[0])
	if !isURL {
		host, port := parseHostPortArgs(args)
		return probeTarget{hostname: host, port: port, protocol: consts.TCP}, nil
	}

	u, err := url.Parse(args[0])
	if err != nil || u.Hostname() == "" {
		return probeTarget{}, fmt.Errorf("invalid URL %q", args[0])
	}

	target := probeTarget{
		hostname: u.Hostname(),
		port:     u.Port(),
		protocol: protocol,
	}

	// A trailing port argument wins over the one embedded in the URL.
	if len(args) > 1 {
		target.port = args[1]
	}

	if target.port == "" {
		target.port = defaultPort(protocol)
	}

	// Rebuild the host so the Host header matches the port we will dial,
	// while still omitting the port when it is the scheme's default - some
	// virtual hosts only match the bare name.
	if target.port == defaultPort(protocol) {
		u.Host = target.hostname
	} else {
		u.Host = net.JoinHostPort(target.hostname, target.port)
	}

	target.url = u.String()

	return target, nil
}

// schemeProtocol reports the protocol implied by target's URL scheme, and
// whether it had one we handle at all.
func schemeProtocol(target string) (consts.Protocol, bool) {
	switch {
	case strings.HasPrefix(target, "http://"):
		return consts.HTTP, true
	case strings.HasPrefix(target, "https://"):
		return consts.HTTPS, true
	default:
		return "", false
	}
}

func defaultPort(protocol consts.Protocol) string {
	if protocol == consts.HTTPS {
		return "443"
	}
	return "80"
}

// usage prints how tcping should be run
func usage() {
	fmt.Printf("\nTCPING version %s\n\n", version)
	fmt.Println("Try running tcping like:")
	fmt.Println("tcping www.example.com 443")
	fmt.Println("Or use the <hostname/ip:port> format:")
	fmt.Println("tcping www.example.com:443")
	fmt.Println("Or probe over HTTP(S) by giving a URL:")
	fmt.Println("tcping https://www.example.com/health")
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

// showVersion displays the version and exits
func showVersion() {
	fmt.Printf("TCPING version %s\n", version)
	os.Exit(0)
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

// checkForUpdates checks for newer versions of tcping
func checkForUpdates() {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		fmt.Printf("Could not create request: %s", err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: 5 * time.Second}

	// optional (GitHub recommends)
	req.Header.Set("User-Agent", "pouriyajamshidi-tcping-update-check")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to check for updates %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to check for updates: HTTP %d\n", resp.StatusCode)
		os.Exit(1)
	}

	release := struct {
		TagName string `json:"tag_name"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fmt.Printf("Failed to parse release info: %s\n", err)
		os.Exit(1)
	}

	reg := `^v?(\d+\.\d+\.\d+)$`
	latestTagName := release.TagName
	re := regexp.MustCompile(reg)
	m := re.FindStringSubmatch(latestTagName)
	if len(m) == 0 {
		fmt.Printf("Failed to check for updates. The version name does not match the rule: %s\n", latestTagName)
		os.Exit(1)
	}

	latestVer := m[1]

	comparison := compareVersions(version, latestVer)

	if comparison < 0 {
		fmt.Printf("Found newer version: %s\n", latestVer)
		fmt.Printf("Please update TCPING from the URL below:\n")
		fmt.Printf("https://github.com/%s/%s/releases/tag/%s\n",
			owner, repo, latestTagName)
	} else if comparison > 0 {
		fmt.Printf("Current version %s is newer than the latest release %s\n",
			version, latestVer)
	} else {
		fmt.Printf("You have the latest version: %s\n", version)
	}

	os.Exit(0)
}

// SecondsToDuration returns the corresponding duration from seconds expressed with a float.
func SecondsToDuration(seconds float64) time.Duration {
	return time.Duration(1000*seconds) * time.Millisecond
}
