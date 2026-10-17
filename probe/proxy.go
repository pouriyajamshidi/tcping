package probe

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/pouriyajamshidi/tcping/v3/internal/version"
)

// httpConnect asks an HTTP proxy to open a connection to target, which is
// what a CONNECT request is for. Any answer but a 2xx means it did not.
func httpConnect(conn net.Conn, target string, user *url.Userinfo) error {
	req := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Opaque: target},
		Host:   target,
		Header: http.Header{"User-Agent": {version.UserAgent}},
	}

	if user != nil {
		password, _ := user.Password()
		credentials := base64.StdEncoding.EncodeToString([]byte(user.Username() + ":" + password))
		req.Header.Set("Proxy-Authorization", "Basic "+credentials)
	}

	if err := req.Write(conn); err != nil {
		return err
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("the proxy answered %s", resp.Status)
	}

	return nil
}
