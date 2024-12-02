// Simple HTTPS Listener For HTTP Redirect
//
// https://github.com/bddjr/hlfhr/tree/main/simplehlfhr
package simplehlfhr

import (
	"io"
	"net"
	"net/http"

	// Only use:
	//  - IsHttpServerShuttingDown
	//  - ConnFirstByteLooksLikeHttp
	//  - ErrHttpOnHttpsPort
	"github.com/bddjr/hlfhr"
)

const resp = "HTTP/1.1 300 Multiple Choices\r\n" +
	"Connection: close\r\n" +
	"Content-Type: text/html\r\n" +
	"\r\n" +
	"<script>location.protocol='https:'</script>"

func ListenAndServeTLS(srv *http.Server, certFile, keyFile string) error {
	if hlfhr.IsHttpServerShuttingDown(srv) {
		return http.ErrServerClosed
	}
	addr := srv.Addr
	if addr == "" {
		addr = ":https"
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()

	l = NewListener(l, srv)

	return srv.ServeTLS(l, certFile, keyFile)
}

func NewListener(inner net.Listener, srv *http.Server) net.Listener {
	return &Listener{
		Listener: inner,
		srv:      srv,
	}
}

type Listener struct {
	net.Listener
	srv *http.Server
}

func (l *Listener) Accept() (c net.Conn, err error) {
	c, err = l.Listener.Accept()
	if err == nil {
		c = &conn{
			Conn: c,
			l:    l,
		}
	}
	return
}

type conn struct {
	net.Conn
	// Reading TLS if nil
	l *Listener
}

func (c *conn) maxHeaderLen() int {
	if c.l.srv != nil && c.l.srv.MaxHeaderBytes != 0 {
		return c.l.srv.MaxHeaderBytes
	}
	return http.DefaultMaxHeaderBytes
}

func (c *conn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if c.l == nil || err != nil || n <= 0 {
		return n, err
	}

	if !hlfhr.ConnFirstByteLooksLikeHttp(b[0]) {
		// Not looks like HTTP.
		// TLS handshake: 0x16
		c.l = nil
		return n, nil
	}

	// Looks like HTTP.
	// len(b) == 576
	defer c.Conn.Close()
	lastLF := -len("\r\n") - 1
	maxHeaderLen := c.maxHeaderLen()

	for {
		// Fix "connection was reset" for method "GET"
		if n > 0 {
			for i, v := range b[:n] {
				if v != '\n' {
					continue
				}
				if i-lastLF <= len("\r\n") {
					// Script Redirect
					io.WriteString(c.Conn, resp)
					return 0, hlfhr.ErrHttpOnHttpsPort
				}
				lastLF = i
			}
			if n >= maxHeaderLen {
				return 0, io.EOF
			}
			maxHeaderLen -= n
			lastLF -= n
		}

		if len(b) > maxHeaderLen {
			b = b[:maxHeaderLen]
		}
		n, err = c.Conn.Read(b)
		if err != nil {
			return 0, err
		}
	}
}
