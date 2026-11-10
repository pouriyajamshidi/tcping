package cli

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/config"
	"github.com/pouriyajamshidi/tcping/v3/dns"
	"github.com/pouriyajamshidi/tcping/v3/nic"
)

func TestProcessUserInputSpotsIPLiterals(t *testing.T) {
	tests := []struct {
		target string
		want   bool
	}{
		{target: "127.0.0.1", want: true},
		{target: "::1", want: true},
		{target: "0:0:0:0:0:0:0:1", want: true},
		{target: "2001:0db8::1", want: true},
		{target: "[::1]", want: true},
		{target: "localhost", want: false},
	}

	// ProcessUserInput registers its flags on the global flag set, which
	// panics when done twice, so each case runs in its own process.
	if target := os.Getenv(subprocessCase); target != "" {
		os.Args = []string{"tcping", target, "443"}
		cfg, _ := ProcessUserInput()
		fmt.Printf("TargetIsIP=%v\n", cfg.TargetIsIP)
		os.Exit(0)
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			out, errOut, code := runCase(t, "TestProcessUserInputSpotsIPLiterals", tt.target)
			if code != 0 {
				t.Fatalf("exited with %d. Output:\n%s%s", code, out, errOut)
			}

			want := fmt.Sprintf("TargetIsIP=%v", tt.want)
			if !strings.Contains(out, want) {
				t.Errorf("%s: want %s. Output:\n%s", tt.target, want, out)
			}
		})
	}
}

// Asking for help is not a mistake, so it is answered on stdout and with a
// zero exit status, which is what makes "tcping --help | less" work.
func TestProcessUserInputHelp(t *testing.T) {
	if arg := os.Getenv(subprocessCase); arg != "" {
		os.Args = []string{"tcping", "example.com", "443", arg}
		ProcessUserInput()
		t.Fatalf("ProcessUserInput returned instead of exiting for %q", arg)
	}

	for _, arg := range []string{"-h", "--h", "-help", "--help"} {
		t.Run(arg, func(t *testing.T) {
			out, errOut, code := runCase(t, "TestProcessUserInputHelp", arg)
			if code != 0 {
				t.Errorf("exited with %d, want 0. Output:\n%s%s", code, out, errOut)
			}

			if !strings.Contains(out, "[optional flags]") || !strings.Contains(out, "-resolve-every-probe") {
				t.Errorf("the flags were not printed on stdout. Output:\n%s", out)
			}

			if errOut != "" {
				t.Errorf("wrote to stderr, want it all on stdout:\n%s", errOut)
			}
		})
	}
}

func TestProcessUserInputTimeout(t *testing.T) {
	tests := []struct {
		args string
		want time.Duration
	}{
		{args: "127.0.0.1 443", want: time.Second},
		{args: "127.0.0.1 443 -t 5", want: 5 * time.Second},
		{args: "http://127.0.0.1/", want: 3 * time.Second},
		{args: "https://127.0.0.1/", want: 3 * time.Second},
		{args: "https://127.0.0.1/ -t 1", want: time.Second},
		{args: "https://127.0.0.1/ -t 0", want: 0},
	}

	if args := os.Getenv(subprocessCase); args != "" {
		os.Args = append([]string{"tcping"}, strings.Fields(args)...)
		cfg, _ := ProcessUserInput()
		fmt.Printf("Timeout=%v\n", cfg.Timeout)
		os.Exit(0)
	}

	for _, tt := range tests {
		t.Run(tt.args, func(t *testing.T) {
			out, errOut, code := runCase(t, "TestProcessUserInputTimeout", tt.args)
			if code != 0 {
				t.Fatalf("exited with %d. Output:\n%s%s", code, out, errOut)
			}

			want := fmt.Sprintf("Timeout=%v", tt.want)
			if !strings.Contains(out, want) {
				t.Errorf("%s: want %s. Output:\n%s", tt.args, want, out)
			}
		})
	}
}

// A target tcping cannot make sense of is refused up front with what is
// wrong, not probed in part or left to fail as a lookup.
func TestProcessUserInputRejectsBadTargets(t *testing.T) {
	tests := []struct {
		args string
		want string
	}{
		{args: "example.com 443 444", want: `unexpected argument "444", give only the host and the port`},
		{args: "example.com 443 444 -c 1", want: `unexpected argument "444", give only the host and the port`},
		{args: "https://example.com 443 444", want: `unexpected argument "444", give only the host and the port`},
		{args: "example.com:443 80", want: `"example.com:443" already has a port, give either host:port or host port`},
		{args: "[::1]:443 80", want: `"[::1]:443" already has a port, give either host:port or host port`},
	}

	if args := os.Getenv(subprocessCase); args != "" {
		os.Args = append([]string{"tcping"}, strings.Fields(args)...)
		ProcessUserInput()
		t.Fatalf("ProcessUserInput returned instead of exiting for %q", args)
	}

	for _, tt := range tests {
		t.Run(tt.args, func(t *testing.T) {
			out, errOut, code := runCase(t, "TestProcessUserInputRejectsBadTargets", tt.args)
			if code != 1 {
				t.Errorf("exited with %d, want 1. Output:\n%s%s", code, out, errOut)
			}

			if errOut != tt.want+"\n" {
				t.Errorf("stderr = %q, want %q", errOut, tt.want+"\n")
			}

			if out != "" {
				t.Errorf("wrote to stdout, want nothing there:\n%s", out)
			}
		})
	}
}

func TestProxyForFlag(t *testing.T) {
	tests := []struct {
		name      string
		proxyFlag string
		protocol  config.Protocol
		want      string
		wantErr   bool
	}{
		{name: "HTTP proxy", proxyFlag: "http://proxy.example.com:3128", protocol: config.HTTPS, want: "http://proxy.example.com:3128"},
		{name: "SOCKS proxy", proxyFlag: "socks5://127.0.0.1:1080", protocol: config.HTTP, want: "socks5://127.0.0.1:1080"},
		{name: "HTTP proxy for a TCP target", proxyFlag: "http://proxy.example.com:3128", protocol: config.TCP, want: "http://proxy.example.com:3128"},
		{name: "HTTPS proxy for a TCP target", proxyFlag: "https://proxy.example.com", protocol: config.TCP, want: "https://proxy.example.com"},
		{name: "SOCKS proxy for a TCP target", proxyFlag: "socks5://127.0.0.1:1080", protocol: config.TCP, want: "socks5://127.0.0.1:1080"},
		{name: "UDP target", proxyFlag: "http://proxy.example.com:3128", protocol: config.UDP, wantErr: true},
		{name: "no scheme", proxyFlag: "proxy.example.com:3128", protocol: config.HTTPS, wantErr: true},
		{name: "unsupported scheme", proxyFlag: "ftp://proxy.example.com", protocol: config.HTTPS, wantErr: true},
		{name: "no proxy", proxyFlag: "", protocol: config.UDP, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := proxyFor(tt.proxyFlag, tt.protocol)
			if (err != nil) != tt.wantErr {
				t.Fatalf("proxyFor(%q, %s) error = %v, wantErr %v", tt.proxyFlag, tt.protocol, err, tt.wantErr)
			}

			got := ""
			if proxy != nil {
				got = proxy.String()
			}
			if got != tt.want {
				t.Errorf("proxyFor(%q, %s) = %q, want %q", tt.proxyFlag, tt.protocol, got, tt.want)
			}
		})
	}
}

func TestProxyAddress(t *testing.T) {
	resolver, err := dns.NewResolver("", time.Second, false, false, nic.NetworkInterface{})
	if err != nil {
		t.Fatalf("NewResolver() failed: %v", err)
	}

	tests := []struct {
		proxy   string
		want    string
		wantErr bool
	}{
		{proxy: "http://127.0.0.1:3128", want: "127.0.0.1:3128"},
		{proxy: "http://127.0.0.1", want: "127.0.0.1:80"},
		{proxy: "https://127.0.0.1", want: "127.0.0.1:443"},
		{proxy: "socks5://[::1]", want: "[::1]:1080"},
		{proxy: "socks5h://[::1]", want: "[::1]:1080"},
		{proxy: "http://127.0.0.1:99999", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.proxy, func(t *testing.T) {
			proxy, _ := url.Parse(tt.proxy)

			got, err := proxyAddress(resolver, proxy)
			if (err != nil) != tt.wantErr {
				t.Fatalf("proxyAddress(%q) error = %v, wantErr %v", tt.proxy, err, tt.wantErr)
			}

			if !tt.wantErr && got.String() != tt.want {
				t.Errorf("proxyAddress(%q) = %s, want %s", tt.proxy, got, tt.want)
			}
		})
	}
}

// A proxy is only used when --proxy asks for one, so the environment
// variables other tools read must not change where a probe goes.
func TestProxyForIgnoresEnvironment(t *testing.T) {
	for _, key := range []string{"HTTPS_PROXY", "HTTP_PROXY", "ALL_PROXY", "https_proxy", "http_proxy", "all_proxy"} {
		t.Setenv(key, "http://proxy.example.com:3128")
	}

	for _, protocol := range []config.Protocol{config.TCP, config.HTTP, config.HTTPS} {
		proxy, err := proxyFor("", protocol)
		if err != nil || proxy != nil {
			t.Errorf("proxyFor(\"\", %s) = %v, %v, want no proxy", protocol, proxy, err)
		}
	}
}
