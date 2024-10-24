package hlfhr

import (
	"bufio"
	"errors"
	"log"
	"net"
	"net/http"
)

var ErrHttpOnHttpsPort = errors.New("client sent an HTTP request to an HTTPS server")

type conn struct {
	isReadingTLS bool
	net.Conn
	l *Listener
}

func IsMyConn(inner net.Conn) bool {
	_, ok := inner.(*conn)
	return ok
}

func (c *conn) logf(format string, args ...any) {
	if c.l.HttpServer != nil && c.l.HttpServer.ErrorLog != nil {
		c.l.HttpServer.ErrorLog.Printf(format, args...)
		return
	}
	log.Printf(format, args...)
}

func (c *conn) serve(firstByte byte) {
	// Read HTTP header
	chhr := connHttpHeaderReader{c: c}
	chhr.setStatusFirstByte(firstByte)

	r, err := http.ReadRequest(bufio.NewReader(&chhr))
	if err != nil {
		c.logf("hlfhr: Read request error from %s: %v", c.Conn.RemoteAddr(), err)
		return
	}
	if r.Host == "" {
		c.logf("hlfhr: error form %s: missing required Host header", c.Conn.RemoteAddr())
		return
	}

	// Response
	w := newResponse(c.Conn)
	chhr.setStatusReadingBody()

	if c.l.HttpOnHttpsPortErrorHandler != nil {
		// Handler
		c.l.HttpOnHttpsPortErrorHandler.ServeHTTP(w, r)
	} else {
		// Redirect
		RedirectToHttps(w, r, 302)
	}

	// Write
	err = w.flush()
	if err != nil {
		c.logf("hlfhr: Write error for %s: %v", c.Conn.RemoteAddr(), err)
	}
}

func (c *conn) Read(b []byte) (int, error) {
	if c.isReadingTLS || len(b) == 0 {
		return c.Conn.Read(b)
	}

	// Read 1 Byte Header
	n, err := c.Conn.Read(b[:1])
	if err != nil || n == 0 {
		return 0, err
	}

	if ConnFirstByteLooksLikeHttp(b[0]) {
		// Looks like HTTP.
		c.serve(b[0])
		return 0, ErrHttpOnHttpsPort
	}

	// Not looks like HTTP.
	// TLS handshake: 0x16
	c.isReadingTLS = true
	if len(b) > 1 {
		n, err = c.Conn.Read(b[1:])
		return n + 1, err
	}
	return 1, nil
}
