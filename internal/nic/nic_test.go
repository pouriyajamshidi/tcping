package nic

import (
	"net"
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
	if !ni.SourceIP.Equal(mustParseIP(t, "127.0.0.1")) {
		t.Errorf("SourceIP = %v, want 127.0.0.1", ni.SourceIP)
	}
	if ni.Dialer.LocalAddr == nil {
		t.Error("Dialer.LocalAddr is nil, want it set to SourceIP")
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
