package probers

import (
	"net"
	"testing"
)

// testServerListen creates a new listener
// on port 12345 and automatically starts it.
//
// Use t.Cleanup with srv.Close() to close it after
// the test, so that other tests are not affected.
//
// It could fail if net.Listen or Accept has failed.
func testServerListen(t *testing.T) net.Listener {
	srv, err := net.Listen("tcp", ":12345")
	if err != nil {
		t.Errorf("test server: %v", err)
	}

	go func() {
		for {
			c, err := srv.Accept()
			if err != nil {
				return
			}

			c.Close()
		}
	}()

	return srv
}
