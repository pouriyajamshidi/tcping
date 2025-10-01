// Package tcping provides TCP connectivity testing functionality with customizable probing and output formatting.
package tcping

import (
	"context"

	"github.com/pouriyajamshidi/tcping/v3/pingers"
)

var (
	// List of compile time checks for all pingers
	_ Pinger = (*pingers.TCPPinger)(nil)
)

// Pinger defines the interface for network connectivity testing implementations.
type Pinger interface {
	Ping(ctx context.Context) error
	IP() string
	Port() uint16
}
