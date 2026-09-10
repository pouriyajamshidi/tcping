package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	// maxEventSize is the largest body we accept. tcping's own events are a
	// few hundred bytes, so anything near this did not come from one.
	maxEventSize = 1 << 20

	// A client that connects and then says nothing should not hold a
	// goroutine forever.
	readHeaderTimeout = 5 * time.Second
)

// ListenJSON prints every JSON event POSTed to address, so that a
// `tcping --json-url` on the other end has somewhere to send its run without
// a server having to be written first. It returns when ctx is cancelled.
//
// Any path is accepted, so the address given to --json-url can carry whatever
// path suits the sending side.
//
// The events go to stdout exactly as they arrived and everything else goes to
// stderr, so the output can be piped into another program or redirected into
// a file without anything of ours in the way.
func ListenJSON(ctx context.Context, address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	server := &http.Server{
		Handler:           http.HandlerFunc(printEvent),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	// Closing the server is what unblocks Serve below.
	stop := context.AfterFunc(ctx, func() { _ = server.Close() })
	defer stop()

	fmt.Fprintf(os.Stderr, "Listening for JSON events on %s\n", listener.Addr())

	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

// printEvent prints one event and answers the sender. What is not JSON is
// refused rather than printed, so whatever reads our stdout can count on
// getting only events.
func printEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST is accepted", http.StatusMethodNotAllowed)
		return
	}

	event, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxEventSize))
	if err != nil {
		http.Error(w, "could not read the event", http.StatusBadRequest)
		return
	}

	if !json.Valid(event) {
		http.Error(w, "the body is not JSON", http.StatusBadRequest)
		return
	}

	fmt.Printf("%s\n", bytes.TrimSpace(event))

	w.WriteHeader(http.StatusNoContent)
}
