// Package server holds the listeners tcping can run instead of probing, so
// that a probe has something on the other end to answer it.
package server

import (
	"context"
	"fmt"
	"net"
)

// maxDatagramSize is the buffer datagrams are read into. Probes send a few
// bytes, so anything larger than this gets truncated and echoed back short.
const maxDatagramSize = 1500

// ListenUDP echoes every datagram it receives on address back to its sender,
// so a `tcping udp://host port` on the other end can tell that its packet
// actually arrived. It returns when ctx is cancelled.
func ListenUDP(ctx context.Context, address string) error {
	conn, err := net.ListenPacket("udp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Closing the socket is what unblocks the read below.
	stop := context.AfterFunc(ctx, func() { conn.Close() })
	defer stop()

	fmt.Printf("Listening for UDP probes on %s\n", conn.LocalAddr())

	buf := make([]byte, maxDatagramSize)

	for {
		n, peer, err := conn.ReadFrom(buf)
		if err != nil {
			// We closed the socket ourselves on Ctrl+C, so this is not a
			// failure.
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		if _, err := conn.WriteTo(buf[:n], peer); err != nil {
			fmt.Printf("Failed to reply to %s: %s\n", peer, err)
			continue
		}

		fmt.Printf("Echoed %d bytes back to %s\n", n, peer)
	}
}
