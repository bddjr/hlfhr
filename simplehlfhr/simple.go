// https://github.com/bddjr/hlfhr/tree/main/simplehlfhr
package simplehlfhr

import (
	"fmt"
	"io"
	"net"
	"net/http"

	// Only use:
	//  - IsHttpServerShuttingDown
	//  - FirstByteLooksLikeHttp
	//  - ErrHttpOnHttpsPort
	"github.com/bddjr/hlfhr"
)

const respBody = "<script>location.protocol='https:'</script>"

var resp = fmt.Append(nil,
	"HTTP/1.1 300 Multiple Choices\r\n",
	"Connection: close\r\n",
	"Content-Type: text/html\r\n",
	"Content-Length: ", len(respBody), "\r\n",
	"\r\n",
	respBody,
)

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
	l, ok := inner.(*Listener)
	if !ok {
		l = &Listener{Listener: inner}
	}
	l.HttpServer = srv
	return l
}

type Listener struct {
	net.Listener
	HttpServer *http.Server
}

func (l *Listener) Accept() (c net.Conn, err error) {
	c, err = l.Listener.Accept()
	if err == nil {
		c = &conn{
			Conn: c,
			srv:  l.HttpServer,
		}
	}
	return
}

type conn struct {
	isReadingTLS bool
	net.Conn
	srv *http.Server
}

func (c *conn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if c.isReadingTLS || err != nil || n <= 0 {
		return n, err
	}

	if !hlfhr.ConnFirstByteLooksLikeHttp(b[0]) || len(b) < 576 {
		// Not looks like HTTP.
		// TLS handshake: 0x16
		c.isReadingTLS = true
		return n, nil
	}

	// len(b) == 576

	// Looks like HTTP.
	defer c.Conn.Close()
	lastLF := -len("\r\n") - 1
	maxHeaderLen := http.DefaultMaxHeaderBytes
	if c.srv != nil && c.srv.MaxHeaderBytes != 0 {
		maxHeaderLen = c.srv.MaxHeaderBytes
	}

	for {
		// Fix "connection was reset" for method "GET"
		if n > 0 {
			for i, v := range b[:n] {
				if v == '\n' {
					if i-lastLF <= len("\r\n") {
						// Script Redirect
						c.Conn.Write(resp)
						return 0, hlfhr.ErrHttpOnHttpsPort
					}
					lastLF = i
				}
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
