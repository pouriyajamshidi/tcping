package cli

import (
	"fmt"
	"os"
	"strings"
	"testing"
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

// Go reads the proxy environment variables once per process and keeps them,
// so each case runs in its own process with its own environment.
func TestProxyForEnvironment(t *testing.T) {
	tests := []struct {
		name      string
		env       map[string]string
		proxyFlag string
		targetURL string
		want      string
	}{
		{
			name:      "HTTPS_PROXY for an https target",
			env:       map[string]string{"HTTPS_PROXY": "http://proxy.example.com:3128"},
			targetURL: "https://example.com/",
			want:      "http://proxy.example.com:3128",
		},
		{
			name:      "HTTP_PROXY for an http target",
			env:       map[string]string{"HTTP_PROXY": "http://proxy.example.com:3128"},
			targetURL: "http://example.com/",
			want:      "http://proxy.example.com:3128",
		},
		{
			name:      "NO_PROXY leaves the target alone",
			env:       map[string]string{"HTTPS_PROXY": "http://proxy.example.com:3128", "NO_PROXY": "example.com"},
			targetURL: "https://example.com/",
			want:      "<nil>",
		},
		{
			name:      "the flag wins over the environment",
			env:       map[string]string{"HTTPS_PROXY": "http://proxy.example.com:3128"},
			proxyFlag: "socks5://127.0.0.1:1080",
			targetURL: "https://example.com/",
			want:      "socks5://127.0.0.1:1080",
		},
	}

	if name := os.Getenv(subprocessCase); name != "" {
		for _, tt := range tests {
			if tt.name == name {
				proxy, err := proxyFor(tt.proxyFlag, tt.targetURL)
				fmt.Printf("proxy=%v err=%v\n", proxy, err)
			}
		}
		os.Exit(0)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, key := range []string{"HTTPS_PROXY", "HTTP_PROXY", "NO_PROXY", "https_proxy", "http_proxy", "no_proxy"} {
				t.Setenv(key, tt.env[key])
			}

			out, errOut, code := runCase(t, "TestProxyForEnvironment", tt.name)
			if code != 0 {
				t.Fatalf("exited with %d. Output:\n%s%s", code, out, errOut)
			}

			want := fmt.Sprintf("proxy=%s err=<nil>", tt.want)
			if !strings.Contains(out, want) {
				t.Errorf("want %s. Output:\n%s", want, out)
			}
		})
	}
}
