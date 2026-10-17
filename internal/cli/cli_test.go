package cli

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

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

func TestProxyForFlag(t *testing.T) {
	tests := []struct {
		name      string
		proxyFlag string
		targetURL string
		want      string
		wantErr   bool
	}{
		{name: "HTTP proxy", proxyFlag: "http://proxy.example.com:3128", targetURL: "https://example.com/", want: "http://proxy.example.com:3128"},
		{name: "SOCKS proxy", proxyFlag: "socks5://127.0.0.1:1080", targetURL: "http://example.com/", want: "socks5://127.0.0.1:1080"},
		{name: "no scheme", proxyFlag: "proxy.example.com:3128", targetURL: "https://example.com/", wantErr: true},
		{name: "unsupported scheme", proxyFlag: "ftp://proxy.example.com", targetURL: "https://example.com/", wantErr: true},
		{name: "TCP target", proxyFlag: "http://proxy.example.com:3128", targetURL: "", wantErr: true},
		{name: "no proxy for a TCP target", proxyFlag: "", targetURL: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := proxyFor(tt.proxyFlag, tt.targetURL)
			if (err != nil) != tt.wantErr {
				t.Fatalf("proxyFor(%q, %q) error = %v, wantErr %v", tt.proxyFlag, tt.targetURL, err, tt.wantErr)
			}

			got := ""
			if proxy != nil {
				got = proxy.String()
			}
			if got != tt.want {
				t.Errorf("proxyFor(%q, %q) = %q, want %q", tt.proxyFlag, tt.targetURL, got, tt.want)
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

	for _, targetURL := range []string{"https://example.com/", "http://example.com/"} {
		proxy, err := proxyFor("", targetURL)
		if err != nil || proxy != nil {
			t.Errorf("proxyFor(\"\", %q) = %v, %v, want no proxy", targetURL, proxy, err)
		}
	}
}
