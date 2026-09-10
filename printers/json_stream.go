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

// Probes are usually a second apart, so a POST that hangs longer than this
// would hold up the next probe.
const jsonStreamTimeout = 2 * time.Second

// jsonStreamWriter is where the JSON printer writes when an address was
// given instead of a terminal: every event is POSTed to that address. The
// encoder writes one event per Write call, so each POST body is exactly one
// JSON event and the printer itself stays the same.
type jsonStreamWriter struct {
	client *http.Client
	url    string
	warned bool // Whether we already complained about a POST that failed.
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

	return &jsonStreamWriter{
		client: &http.Client{Timeout: jsonStreamTimeout},
		url:    address,
	}, nil
}

// Write POSTs one JSON event. A failure does not stop the probing, and we
// say so only once, so a server that is down does not fill the terminal with
// the same error every second.
//
// It never reports an error back, because json.Encoder remembers the first
// one it is given and refuses to encode anything afterwards. A single failed
// POST would then end the stream for the rest of the run, and a server that
// was restarted mid-run would never be picked back up.
func (w *jsonStreamWriter) Write(event []byte) (int, error) {
	req, err := http.NewRequest(http.MethodPost, w.url, bytes.NewReader(event))
	if err != nil {
		w.warnOnce("could not build the request: %v", err)
		return len(event), nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", version.UserAgent)

	resp, err := w.client.Do(req)
	if err != nil {
		w.warnOnce("could not reach %s: %v", w.url, err)
		return len(event), nil
	}
	defer resp.Body.Close()

	// Reading the body to the end is what lets the connection be reused by
	// the next probe instead of a new one being opened every second.
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode/100 != 2 {
		w.warnOnce("%s rejected the event with %s", w.url, resp.Status)
	}

	return len(event), nil
}

func (w *jsonStreamWriter) warnOnce(format string, args ...any) {
	if w.warned {
		return
	}

	w.warned = true
	fmt.Fprintf(os.Stderr, "JSON stream Error: "+format+"\n", args...)
	fmt.Fprintln(os.Stderr, "Probing continues, but the events are being dropped.")
}
