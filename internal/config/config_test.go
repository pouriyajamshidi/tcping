package config

import (
	"slices"
	"testing"
)

func TestParseHostPortArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "traditional format: host port",
			args: []string{"example.com", "8080"},
			want: []string{"example.com", "8080"},
		},
		{
			name: "host:port format",
			args: []string{"example.com:8080"},
			want: []string{"example.com", "8080"},
		},
		{
			name: "IPv4:port format",
			args: []string{"192.168.1.1:443"},
			want: []string{"192.168.1.1", "443"},
		},
		{
			name: "IPv6 with brackets and port",
			args: []string{"[2001:db8::1]:8080"},
			want: []string{"2001:db8::1", "8080"},
		},
		{
			name: "IPv6 without brackets and port",
			args: []string{"2001:db8::1:8080"},
			want: []string{"2001:db8::1", "8080"},
		},
		{
			name: "localhost:port format",
			args: []string{"localhost:80"},
			want: []string{"localhost", "80"},
		},
		{
			name: "IPv6 localhost with brackets",
			args: []string{"[::1]:22"},
			want: []string{"::1", "22"},
		},
		{
			name: "IPv6 localhost without brackets",
			args: []string{"::1:22"},
			want: []string{"::1", "22"},
		},
		{
			name: "single argument without colon",
			args: []string{"example.com"},
			want: []string{"example.com"},
		},
		{
			name: "three arguments unchanged",
			args: []string{"example.com", "8080", "extra"},
			want: []string{"example.com", "8080", "extra"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseHostPortArgs(tt.args)
			if !slices.Equal(got, tt.want) {
				t.Errorf("parseHostPortArgs(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
