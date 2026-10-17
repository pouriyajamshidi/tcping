package probe

import (
	"bufio"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
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

// socks5Connect asks a SOCKS5 proxy to open a connection to host:port, the
// way RFC 1928 describes it, logging in first when the proxy URL has a user
// (RFC 1929). A hostname is handed to the proxy as it is, so the proxy looks
// it up on its own.
func socks5Connect(conn net.Conn, host string, port uint16, user *url.Userinfo) error {
	// The greeting is the version and the one login method we offer.
	method := byte(0x00) // no login
	if user != nil {
		method = 0x02 // a username and a password
	}

	if _, err := conn.Write([]byte{5, 1, method}); err != nil {
		return err
	}

	reply := make([]byte, 2)
	if _, err := io.ReadFull(conn, reply); err != nil {
		return err
	}
	if reply[0] != 5 || reply[1] != method {
		return errors.New("the SOCKS5 proxy does not accept our login method")
	}

	if user != nil {
		password, _ := user.Password()
		if len(user.Username()) > 255 || len(password) > 255 {
			return errors.New("the SOCKS5 username and password can be 255 bytes at most")
		}

		login := []byte{1, byte(len(user.Username()))}
		login = append(login, user.Username()...)
		login = append(login, byte(len(password)))
		login = append(login, password...)

		if _, err := conn.Write(login); err != nil {
			return err
		}
		if _, err := io.ReadFull(conn, reply); err != nil {
			return err
		}
		if reply[1] != 0 {
			return errors.New("the SOCKS5 proxy refused the login")
		}
	}

	// The request is the version, CONNECT, a reserved byte, then the address
	// with its type in front: 1 for IPv4, 4 for IPv6, 3 for a hostname.
	request := []byte{5, 1, 0}

	ip, err := netip.ParseAddr(host)
	switch {
	case err == nil && ip.Is4():
		request = append(request, 1)
		request = append(request, ip.AsSlice()...)
	case err == nil:
		request = append(request, 4)
		request = append(request, ip.AsSlice()...)
	case len(host) > 255:
		return errors.New("the hostname is too long for a SOCKS5 proxy")
	default:
		request = append(request, 3, byte(len(host)))
		request = append(request, host...)
	}

	request = binary.BigEndian.AppendUint16(request, port)

	if _, err := conn.Write(request); err != nil {
		return err
	}

	// The reply starts with the version and the result. The address the
	// proxy connected from follows, but we hang up straight away, so it is
	// left unread.
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return err
	}
	if header[1] != 0 {
		return fmt.Errorf("the SOCKS5 proxy could not connect, reply code %d", header[1])
	}

	return nil
}
