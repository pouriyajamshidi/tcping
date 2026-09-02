package config

import (
	"testing"

	"github.com/pouriyajamshidi/tcping/v3/internal/consts"
)

func TestParseHostPortArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantHost string
		wantPort string
	}{
		{
			name:     "traditional format: host port",
			args:     []string{"example.com", "8080"},
			wantHost: "example.com",
			wantPort: "8080",
		},
		{
			name:     "host:port format",
			args:     []string{"example.com:8080"},
			wantHost: "example.com",
			wantPort: "8080",
		},
		{
			name:     "IPv4:port format",
			args:     []string{"192.168.1.1:443"},
			wantHost: "192.168.1.1",
			wantPort: "443",
		},
		{
			name:     "IPv6 with brackets and port",
			args:     []string{"[2001:db8::1]:8080"},
			wantHost: "2001:db8::1",
			wantPort: "8080",
		},
		{
			name:     "IPv6 without brackets and port is ambiguous, returned as-is",
			args:     []string{"2001:db8::1:8080"},
			wantHost: "2001:db8::1:8080",
			wantPort: "",
		},
		{
			name:     "localhost:port format",
			args:     []string{"localhost:80"},
			wantHost: "localhost",
			wantPort: "80",
		},
		{
			name:     "IPv6 localhost with brackets",
			args:     []string{"[::1]:22"},
			wantHost: "::1",
			wantPort: "22",
		},
		{
			name:     "IPv6 localhost without brackets is ambiguous, returned as-is",
			args:     []string{"::1:22"},
			wantHost: "::1:22",
			wantPort: "",
		},
		{
			name:     "single argument without colon",
			args:     []string{"example.com"},
			wantHost: "example.com",
			wantPort: "",
		},
		{
			name:     "three arguments: only first two are used",
			args:     []string{"example.com", "8080", "extra"},
			wantHost: "example.com",
			wantPort: "8080",
		},
		{
			name:     "empty string argument",
			args:     []string{""},
			wantHost: "",
			wantPort: "",
		},
		{
			name:     "host:port with empty port",
			args:     []string{"example.com:"},
			wantHost: "example.com",
			wantPort: "",
		},
		{
			name:     "host:port with empty host",
			args:     []string{":8080"},
			wantHost: "",
			wantPort: "8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHost, gotPort := parseHostPortArgs(tt.args)
			if gotHost != tt.wantHost || gotPort != tt.wantPort {
				t.Errorf("parseHostPortArgs(%v) = (%q, %q), want (%q, %q)",
					tt.args, gotHost, gotPort, tt.wantHost, tt.wantPort)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want int
	}{
		{
			name: "equal versions",
			v1:   "1.2.3",
			v2:   "1.2.3",
			want: 0,
		},
		{
			name: "v1 major less than v2",
			v1:   "1.2.3",
			v2:   "2.0.0",
			want: -1,
		},
		{
			name: "v1 major greater than v2",
			v1:   "3.0.0",
			v2:   "2.9.9",
			want: 1,
		},
		{
			name: "v1 minor less than v2",
			v1:   "1.2.3",
			v2:   "1.3.0",
			want: -1,
		},
		{
			name: "v1 minor greater than v2",
			v1:   "1.4.0",
			v2:   "1.3.9",
			want: 1,
		},
		{
			name: "v1 patch less than v2",
			v1:   "1.2.3",
			v2:   "1.2.4",
			want: -1,
		},
		{
			name: "v1 patch greater than v2",
			v1:   "1.2.5",
			v2:   "1.2.4",
			want: 1,
		},
		{
			name: "v1 shorter but equal in common parts",
			v1:   "1.2",
			v2:   "1.2.0",
			want: -1,
		},
		{
			name: "v1 longer but equal in common parts",
			v1:   "1.2.0",
			v2:   "1.2",
			want: 1,
		},
		{
			name: "v1 shorter and less in common parts",
			v1:   "1.2",
			v2:   "1.3.0",
			want: -1,
		},
		{
			name: "v1 longer and greater in common parts",
			v1:   "2.0.0",
			v2:   "1.9",
			want: 1,
		},
		{
			name: "single-digit versions equal",
			v1:   "1",
			v2:   "1",
			want: 0,
		},
		{
			name: "non-numeric segment treated as zero",
			v1:   "1.x.0",
			v2:   "1.0.0",
			want: 0,
		},
		{
			name: "both empty strings",
			v1:   "",
			v2:   "",
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVersions(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

func TestParseTarget(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantHostname string
		wantPort     string
		wantProtocol consts.Protocol
		wantURL      string
		wantErr      bool
	}{
		{
			name:         "host and port stays TCP",
			args:         []string{"example.com", "8080"},
			wantHostname: "example.com",
			wantPort:     "8080",
			wantProtocol: consts.TCP,
		},
		{
			name:         "host:port stays TCP",
			args:         []string{"example.com:8080"},
			wantHostname: "example.com",
			wantPort:     "8080",
			wantProtocol: consts.TCP,
		},
		{
			name:         "https URL defaults to port 443",
			args:         []string{"https://example.com"},
			wantHostname: "example.com",
			wantPort:     "443",
			wantProtocol: consts.HTTPS,
			wantURL:      "https://example.com",
		},
		{
			name:         "http URL defaults to port 80",
			args:         []string{"http://example.com"},
			wantHostname: "example.com",
			wantPort:     "80",
			wantProtocol: consts.HTTP,
			wantURL:      "http://example.com",
		},
		{
			name:         "the path is kept",
			args:         []string{"https://example.com/health"},
			wantHostname: "example.com",
			wantPort:     "443",
			wantProtocol: consts.HTTPS,
			wantURL:      "https://example.com/health",
		},
		{
			name:         "a port in the URL wins over the default",
			args:         []string{"https://example.com:8443/live"},
			wantHostname: "example.com",
			wantPort:     "8443",
			wantProtocol: consts.HTTPS,
			wantURL:      "https://example.com:8443/live",
		},
		{
			name:         "a trailing port argument wins over the URL",
			args:         []string{"https://example.com:8443", "9443"},
			wantHostname: "example.com",
			wantPort:     "9443",
			wantProtocol: consts.HTTPS,
			wantURL:      "https://example.com:9443",
		},
		{
			name:         "the default port is left out of the Host header",
			args:         []string{"https://example.com", "443"},
			wantHostname: "example.com",
			wantPort:     "443",
			wantProtocol: consts.HTTPS,
			wantURL:      "https://example.com",
		},
		{
			name:         "IPv6 URL",
			args:         []string{"https://[2606:4700::1]:8443/"},
			wantHostname: "2606:4700::1",
			wantPort:     "8443",
			wantProtocol: consts.HTTPS,
			wantURL:      "https://[2606:4700::1]:8443/",
		},
		{
			name:         "udp URL with a port in it",
			args:         []string{"udp://example.com:53"},
			wantHostname: "example.com",
			wantPort:     "53",
			wantProtocol: consts.UDP,
		},
		{
			name:         "udp URL with a trailing port argument",
			args:         []string{"udp://example.com", "9999"},
			wantHostname: "example.com",
			wantPort:     "9999",
			wantProtocol: consts.UDP,
		},
		{
			name:         "udp has no default port",
			args:         []string{"udp://example.com"},
			wantHostname: "example.com",
			wantPort:     "",
			wantProtocol: consts.UDP,
		},
		{
			name:    "a URL without a host is rejected",
			args:    []string{"https://"},
			wantErr: true,
		},
		{
			name: "no arguments at all",
			args: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTarget(tt.args)

			if (err != nil) != tt.wantErr {
				t.Fatalf("parseTarget(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if got.hostname != tt.wantHostname || got.port != tt.wantPort ||
				got.protocol != tt.wantProtocol || got.url != tt.wantURL {
				t.Errorf("parseTarget(%v) = (%q, %q, %q, %q), want (%q, %q, %q, %q)",
					tt.args, got.hostname, got.port, got.protocol, got.url,
					tt.wantHostname, tt.wantPort, tt.wantProtocol, tt.wantURL)
			}
		})
	}
}
