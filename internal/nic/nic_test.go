package nic

import (
	"net"
	"net/netip"
	"testing"
)

func TestNewNetworkInterface_UnknownNameOrIP(t *testing.T) {
	_, err := NewNetworkInterface("this-interface-does-not-exist", false, false)
	if err == nil {
		t.Fatal("expected an error for an unknown interface name, got nil")
	}
}

func TestNewNetworkInterface_IPNotAssignedToAnyInterface(t *testing.T) {
	_, err := NewNetworkInterface("203.0.113.1", false, false)
	if err == nil {
		t.Fatal("expected an error for an IP not assigned to any local interface, got nil")
	}
}

func TestNewNetworkInterface_Loopback(t *testing.T) {
	ni, err := NewNetworkInterface("127.0.0.1", false, false)
	if err != nil {
		t.Fatalf("NewNetworkInterface(127.0.0.1) error = %v", err)
	}

	if !ni.Use {
		t.Error("Use = false, want true")
	}
	if !ni.SourceIPv4.Equal(mustParseIP(t, "127.0.0.1")) {
		t.Errorf("SourceIPv4 = %v, want 127.0.0.1", ni.SourceIPv4)
	}
	if ni.SourceIPv6 != nil {
		t.Errorf("SourceIPv6 = %v, want nil for an IPv4 literal", ni.SourceIPv6)
	}
}

func TestNetworkInterface_LocalIPFor(t *testing.T) {
	ni := NetworkInterface{
		Use:        true,
		SourceIPv4: mustParseIP(t, "192.168.1.10"),
		SourceIPv6: mustParseIP(t, "::1"),
	}

	if got := ni.LocalIPFor(netip.MustParseAddr("93.184.216.34")); !got.Equal(ni.SourceIPv4) {
		t.Errorf("LocalIPFor(v4 target) = %v, want %v", got, ni.SourceIPv4)
	}
	if got := ni.LocalIPFor(netip.MustParseAddr("2606:2800:220:1:248:1893:25c8:1946")); !got.Equal(ni.SourceIPv6) {
		t.Errorf("LocalIPFor(v6 target) = %v, want %v", got, ni.SourceIPv6)
	}
}

func TestNetworkInterface_LocalIPFor_MissingFamilyReturnsNil(t *testing.T) {
	ni := NetworkInterface{Use: true, SourceIPv4: mustParseIP(t, "192.168.1.10")}

	if got := ni.LocalIPFor(netip.MustParseAddr("2606:2800:220:1:248:1893:25c8:1946")); got != nil {
		t.Errorf("LocalIPFor(v6 target) = %v, want nil (interface has no IPv6 address)", got)
	}
}

func mustParseIP(t *testing.T, s string) net.IP {
	t.Helper()
	ip := net.ParseIP(s)
	if ip == nil {
		t.Fatalf("net.ParseIP(%q) returned nil", s)
	}
	return ip
}
