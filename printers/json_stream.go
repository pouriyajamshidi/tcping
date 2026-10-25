package printers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/version"
)

// Events are POSTed in the background, so a slow one does not hold up the
// probing, and a server that takes its time can be given it.
const jsonStreamTimeout = 15 * time.Second

// jsonStreamWriter is where the JSON printer writes when an address was
// given instead of a terminal: every event is POSTed to that address. The
// encoder writes one event per Write call, so each POST body is exactly one
// JSON event and the printer itself stays the same.
//
// The events are not POSTed by Write. They wait in a sendQueue, which POSTs
// them one at a time from a goroutine of its own as soon as they arrive.
type jsonStreamWriter struct {
	client *http.Client
	url    string
	queue  *sendQueue[[]byte]
}

// newJSONStreamWriter creates a writer pointed at the given address. The
// scheme can be left out, so both "localhost:8000/tcping" and
// "http://localhost:8000/tcping" work.
func newJSONStreamWriter(address string) (*jsonStreamWriter, error) {
	if !strings.Contains(address, "://") {
		address = "http://" + address
	}

	parsed, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("%q is not a valid address: %w", address, err)
	}

	if parsed.Host == "" {
		return nil, fmt.Errorf("%q has no host to send the events to", address)
	}

	w := &jsonStreamWriter{
		client: &http.Client{Timeout: jsonStreamTimeout},
		url:    address,
	}
	// A batch of one, since every event is a POST of its own, and no
	// interval, so a collector still sees each event as it happens.
	w.queue = newSendQueue(0, 1, w.send, w.printError)

	return w, nil
}

// Write puts one JSON event in line to be POSTed. The encoder reuses its
// buffer for the next event, so the event is copied first.
//
// It never reports an error back, because json.Encoder remembers the first
// one it is given and refuses to encode anything afterwards. A single failed
// POST would then end the stream for the rest of the run, and a server that
// was restarted mid-run would never be picked back up.
func (w *jsonStreamWriter) Write(event []byte) (int, error) {
	w.queue.add(bytes.Clone(event))
	return len(event), nil
}

// send POSTs one event and says whether it is worth trying again. A failure
// does not stop the probing, but every one of them is printed, so a server
// that stays down is never hidden.
func (w *jsonStreamWriter) send(events [][]byte) (retry bool) {
	req, err := http.NewRequest(http.MethodPost, w.url, bytes.NewReader(events[0]))
	if err != nil {
		w.printError("could not build the request: %v", err)
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", version.UserAgent)

	resp, err := w.client.Do(req)
	if err != nil {
		w.printError("could not reach %s: %v", w.url, err)
		return true
	}
	defer resp.Body.Close()

	// Reading the body to the end is what lets the connection be reused by
	// the next event instead of a new one being opened every time.
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode/100 != 2 {
		w.printError("%s rejected the event with %s", w.url, resp.Status)
	}

	// A server that is struggling, or asks us to slow down, may take the
	// same event later. Anything else it rejected it would reject again.
	return resp.StatusCode/100 == 5 || resp.StatusCode == http.StatusTooManyRequests
}

// printError goes to the terminal, since an error here usually means the
// server is the thing that is not working.
func (w *jsonStreamWriter) printError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "JSON stream Error: "+format+"\n", args...)
}

// close POSTs every event still waiting and closes the connection to the
// server.
func (w *jsonStreamWriter) close() {
	w.queue.close()
	w.client.CloseIdleConnections()
}
