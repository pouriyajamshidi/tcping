package server

import (
	"context"
	"net"
	"testing"
	"time"
)

// freeUDPAddress asks the OS for a free UDP port and releases it, so the
// server under test can bind it.
func freeUDPAddress(t *testing.T) string {
	t.Helper()

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatalf("failed to reserve a UDP port: %v", err)
	}
	address := conn.LocalAddr().String()
	conn.Close()

	return address
}

func TestListenUDP_EchoesBackWhatItReceives(t *testing.T) {
	address := freeUDPAddress(t)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go ListenUDP(ctx, address)

	conn, err := net.Dial("udp", address)
	if err != nil {
		t.Fatalf("failed to dial the server: %v", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(2 * time.Second))

	// The server may not have bound the port yet, and a datagram sent
	// before it does is simply lost, so keep sending until one comes back.
	sent := []byte("42")
	reply := make([]byte, len(sent))

	for range 20 {
		if _, err := conn.Write(sent); err != nil {
			t.Fatalf("failed to send to the server: %v", err)
		}

		conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

		n, err := conn.Read(reply)
		if err != nil {
			continue
		}

		if string(reply[:n]) != string(sent) {
			t.Fatalf("reply = %q, want %q", reply[:n], sent)
		}

		return
	}

	t.Fatal("the server never echoed anything back")
}

func TestListenUDP_StopsWhenTheContextIsCancelled(t *testing.T) {
	address := freeUDPAddress(t)

	ctx, cancel := context.WithCancel(context.Background())

	stopped := make(chan error, 1)
	go func() { stopped <- ListenUDP(ctx, address) }()

	// Give it a moment to bind before asking it to stop.
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-stopped:
		if err != nil {
			t.Errorf("ListenUDP() error = %v, want nil since we asked it to stop", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ListenUDP() did not return after the context was cancelled")
	}
}

func TestListenUDP_FailsOnAnAddressItCannotBind(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatalf("failed to occupy a UDP port: %v", err)
	}
	defer conn.Close()

	if err := ListenUDP(context.Background(), conn.LocalAddr().String()); err == nil {
		t.Fatal("ListenUDP() error = nil, want an error since the port is already in use")
	}
}
