package config

import (
	"testing"
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
