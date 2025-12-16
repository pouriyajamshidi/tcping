package app_test

import (
	"net"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3"
	"github.com/pouriyajamshidi/tcping/v3/internal/app"
)

func TestProberConfig_Defaults(t *testing.T) {
	config := app.ProberConfig{}

	if config.Hostname != "" {
		t.Errorf("Hostname should be empty, got %q", config.Hostname)
	}

	if config.Port != 0 {
		t.Errorf("Port should be 0, got %d", config.Port)
	}

	if config.Timeout != 0 {
		t.Errorf("Timeout should be 0, got %v", config.Timeout)
	}
}

func TestProberConfig_WithNetworkInterface(t *testing.T) {
	config := app.ProberConfig{
		InterfaceDialer: &net.Dialer{
			LocalAddr: &net.TCPAddr{
				IP: net.ParseIP("192.168.1.100"),
			},
			Timeout: 5 * time.Second,
		},
		InterfaceName: "eth0",
	}

	if config.InterfaceDialer == nil {
		t.Error("InterfaceDialer should not be nil")
	}

	if config.InterfaceName != "eth0" {
		t.Errorf("InterfaceName = %q, want %q", config.InterfaceName, "eth0")
	}
}

func TestProberConfig_TimingOptions(t *testing.T) {
	config := app.ProberConfig{
		Timeout:  2 * time.Second,
		Interval: 500 * time.Millisecond,
	}

	if config.Timeout != 2*time.Second {
		t.Errorf("Timeout = %v, want %v", config.Timeout, 2*time.Second)
	}

	if config.Interval != 500*time.Millisecond {
		t.Errorf("Interval = %v, want %v", config.Interval, 500*time.Millisecond)
	}
}

func TestProberConfig_ProbeControl(t *testing.T) {
	config := app.ProberConfig{
		ProbeCountLimit:  10,
		ShowFailuresOnly: true,
	}

	if config.ProbeCountLimit != 10 {
		t.Errorf("ProbeCountLimit = %d, want 10", config.ProbeCountLimit)
	}

	if !config.ShowFailuresOnly {
		t.Error("ShowFailuresOnly should be true")
	}
}

func TestProberConfig_DNSOptions(t *testing.T) {
	config := app.ProberConfig{
		RetryResolveAfter: 5,
		UseIPv4:           true,
		UseIPv6:           false,
	}

	if config.RetryResolveAfter != 5 {
		t.Errorf("RetryResolveAfter = %d, want 5", config.RetryResolveAfter)
	}

	if !config.UseIPv4 {
		t.Error("UseIPv4 should be true")
	}

	if config.UseIPv6 {
		t.Error("UseIPv6 should be false")
	}
}

func TestProberConfig_PrinterConfig(t *testing.T) {
	config := app.ProberConfig{
		PrinterConfig: tcping.PrinterConfig{
			OutputJSON:        true,
			PrettyJSON:        true,
			NoColor:           false,
			WithTimestamp:     true,
			WithSourceAddress: true,
			OutputDBPath:      "/tmp/test.db",
			OutputCSVPath:     "/tmp/test.csv",
			Target:            "example.com",
			Port:              "443",
		},
	}

	pc := config.PrinterConfig

	if !pc.OutputJSON {
		t.Error("OutputJSON should be true")
	}

	if !pc.PrettyJSON {
		t.Error("PrettyJSON should be true")
	}

	if pc.NoColor {
		t.Error("NoColor should be false")
	}

	if !pc.WithTimestamp {
		t.Error("WithTimestamp should be true")
	}

	if !pc.WithSourceAddress {
		t.Error("WithSourceAddress should be true")
	}

	if pc.OutputDBPath != "/tmp/test.db" {
		t.Errorf("OutputDBPath = %q, want %q", pc.OutputDBPath, "/tmp/test.db")
	}

	if pc.OutputCSVPath != "/tmp/test.csv" {
		t.Errorf("OutputCSVPath = %q, want %q", pc.OutputCSVPath, "/tmp/test.csv")
	}

	if pc.Target != "example.com" {
		t.Errorf("Target = %q, want %q", pc.Target, "example.com")
	}

	if pc.Port != "443" {
		t.Errorf("Port = %q, want %q", pc.Port, "443")
	}
}

func TestProberConfig_NetworkInterfaceBinding(t *testing.T) {
	localIP := net.ParseIP("192.168.1.100")

	config := app.ProberConfig{
		InterfaceDialer: &net.Dialer{
			LocalAddr: &net.TCPAddr{
				IP: localIP,
			},
		},
		InterfaceName: "custom-interface",
	}

	if config.InterfaceDialer == nil {
		t.Fatal("InterfaceDialer should not be nil")
	}

	tcpAddr, ok := config.InterfaceDialer.LocalAddr.(*net.TCPAddr)
	if !ok {
		t.Fatal("LocalAddr should be *net.TCPAddr")
	}

	if !tcpAddr.IP.Equal(localIP) {
		t.Errorf("LocalAddr IP = %v, want %v", tcpAddr.IP, localIP)
	}

	if config.InterfaceName != "custom-interface" {
		t.Errorf("InterfaceName = %q, want %q", config.InterfaceName, "custom-interface")
	}
}

func TestProberConfig_IPv4IPv6Selection(t *testing.T) {
	tests := []struct {
		name    string
		useIPv4 bool
		useIPv6 bool
	}{
		{name: "ipv4 only", useIPv4: true, useIPv6: false},
		{name: "ipv6 only", useIPv4: false, useIPv6: true},
		{name: "both false (auto)", useIPv4: false, useIPv6: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := app.ProberConfig{
				UseIPv4: tt.useIPv4,
				UseIPv6: tt.useIPv6,
			}

			if config.UseIPv4 != tt.useIPv4 {
				t.Errorf("UseIPv4 = %v, want %v", config.UseIPv4, tt.useIPv4)
			}

			if config.UseIPv6 != tt.useIPv6 {
				t.Errorf("UseIPv6 = %v, want %v", config.UseIPv6, tt.useIPv6)
			}
		})
	}
}

func TestProberConfig_HostnameAndPort(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		port     uint16
	}{
		{name: "http", hostname: "example.com", port: 80},
		{name: "https", hostname: "secure.example.com", port: 443},
		{name: "custom port", hostname: "api.example.com", port: 8080},
		{name: "ip address", hostname: "192.168.1.1", port: 22},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := app.ProberConfig{
				Hostname: tt.hostname,
				Port:     tt.port,
			}

			if config.Hostname != tt.hostname {
				t.Errorf("Hostname = %q, want %q", config.Hostname, tt.hostname)
			}

			if config.Port != tt.port {
				t.Errorf("Port = %d, want %d", config.Port, tt.port)
			}
		})
	}
}

func TestProberConfig_NonInteractive(t *testing.T) {
	config := app.ProberConfig{
		NonInteractive: true,
	}

	if !config.NonInteractive {
		t.Error("NonInteractive should be true")
	}
}

func TestProberConfig_ShowSourceAddress(t *testing.T) {
	config := app.ProberConfig{
		ShowSourceAddress: true,
	}

	if !config.ShowSourceAddress {
		t.Error("ShowSourceAddress should be true")
	}
}

func TestProberConfig_AllOptions(t *testing.T) {
	config := app.ProberConfig{
		Hostname:          "test.example.com",
		Port:              8080,
		UseIPv4:           true,
		UseIPv6:           false,
		InterfaceName:     "eth0",
		ShowSourceAddress: true,
		Timeout:           3 * time.Second,
		Interval:          200 * time.Millisecond,
		ProbeCountLimit:   20,
		ShowFailuresOnly:  false,
		RetryResolveAfter: 10,
		PrinterConfig: tcping.PrinterConfig{
			OutputJSON:        false,
			PrettyJSON:        false,
			NoColor:           true,
			WithTimestamp:     true,
			WithSourceAddress: true,
			Target:            "test.example.com",
			Port:              "8080",
		},
		NonInteractive: true,
	}

	if config.Hostname != "test.example.com" {
		t.Errorf("Hostname = %q, want %q", config.Hostname, "test.example.com")
	}

	if config.Port != 8080 {
		t.Errorf("Port = %d, want 8080", config.Port)
	}

	if config.Timeout != 3*time.Second {
		t.Errorf("Timeout = %v, want %v", config.Timeout, 3*time.Second)
	}

	if config.ProbeCountLimit != 20 {
		t.Errorf("ProbeCountLimit = %d, want 20", config.ProbeCountLimit)
	}
}
