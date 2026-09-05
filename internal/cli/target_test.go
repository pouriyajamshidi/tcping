package cli

import (
	"testing"

	"github.com/pouriyajamshidi/tcping/v3/config"
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

func TestParseTarget(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantHostname string
		wantPort     string
		wantProtocol config.Protocol
		wantURL      string
		wantErr      bool
	}{
		{
			name:         "host and port stays TCP",
			args:         []string{"example.com", "8080"},
			wantHostname: "example.com",
			wantPort:     "8080",
			wantProtocol: config.TCP,
		},
		{
			name:         "host:port stays TCP",
			args:         []string{"example.com:8080"},
			wantHostname: "example.com",
			wantPort:     "8080",
			wantProtocol: config.TCP,
		},
		{
			name:         "https URL defaults to port 443",
			args:         []string{"https://example.com"},
			wantHostname: "example.com",
			wantPort:     "443",
			wantProtocol: config.HTTPS,
			wantURL:      "https://example.com",
		},
		{
			name:         "http URL defaults to port 80",
			args:         []string{"http://example.com"},
			wantHostname: "example.com",
			wantPort:     "80",
			wantProtocol: config.HTTP,
			wantURL:      "http://example.com",
		},
		{
			name:         "the path is kept",
			args:         []string{"https://example.com/health"},
			wantHostname: "example.com",
			wantPort:     "443",
			wantProtocol: config.HTTPS,
			wantURL:      "https://example.com/health",
		},
		{
			name:         "a port in the URL wins over the default",
			args:         []string{"https://example.com:8443/live"},
			wantHostname: "example.com",
			wantPort:     "8443",
			wantProtocol: config.HTTPS,
			wantURL:      "https://example.com:8443/live",
		},
		{
			name:         "a trailing port argument wins over the URL",
			args:         []string{"https://example.com:8443", "9443"},
			wantHostname: "example.com",
			wantPort:     "9443",
			wantProtocol: config.HTTPS,
			wantURL:      "https://example.com:9443",
		},
		{
			name:         "the default port is left out of the Host header",
			args:         []string{"https://example.com", "443"},
			wantHostname: "example.com",
			wantPort:     "443",
			wantProtocol: config.HTTPS,
			wantURL:      "https://example.com",
		},
		{
			name:         "IPv6 URL",
			args:         []string{"https://[2606:4700::1]:8443/"},
			wantHostname: "2606:4700::1",
			wantPort:     "8443",
			wantProtocol: config.HTTPS,
			wantURL:      "https://[2606:4700::1]:8443/",
		},
		{
			name:         "udp URL with a port in it",
			args:         []string{"udp://example.com:53"},
			wantHostname: "example.com",
			wantPort:     "53",
			wantProtocol: config.UDP,
		},
		{
			name:         "udp URL with a trailing port argument",
			args:         []string{"udp://example.com", "9999"},
			wantHostname: "example.com",
			wantPort:     "9999",
			wantProtocol: config.UDP,
		},
		{
			name:         "udp has no default port",
			args:         []string{"udp://example.com"},
			wantHostname: "example.com",
			wantPort:     "",
			wantProtocol: config.UDP,
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

func TestConvertAndValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		want    uint16
		wantErr bool
	}{
		{
			name: "lowest valid port",
			port: "1",
			want: 1,
		},
		{
			name: "highest valid port",
			port: "65535",
			want: 65535,
		},
		{
			name: "a common port",
			port: "443",
			want: 443,
		},
		{
			name:    "zero is not a usable port",
			port:    "0",
			wantErr: true,
		},
		{
			name:    "one past the highest port",
			port:    "65536",
			wantErr: true,
		},
		{
			name:    "negative",
			port:    "-1",
			wantErr: true,
		},
		{
			name:    "not a number",
			port:    "http",
			wantErr: true,
		},
		{
			name:    "empty",
			port:    "",
			wantErr: true,
		},
		{
			name:    "decimal",
			port:    "443.0",
			wantErr: true,
		},
		{
			name:    "surrounding whitespace is not trimmed for us",
			port:    " 443",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertAndValidatePort(tt.port)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("convertAndValidatePort(%q) = %d, want an error", tt.port, got)
				}
				return
			}

			if err != nil {
				t.Fatalf("convertAndValidatePort(%q) returned an unexpected error: %v", tt.port, err)
			}

			if got != tt.want {
				t.Errorf("convertAndValidatePort(%q) = %d, want %d", tt.port, got, tt.want)
			}
		})
	}
}
