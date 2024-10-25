package hlfhr

import (
	"errors"
	"log"
	"net"
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

func (c *conn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if c.isReadingTLS || err != nil || n <= 0 {
		return n, err
	}

	if !ConnFirstByteLooksLikeHttp(b[0]) || len(b) < 576 {
		// Not looks like HTTP.
		// TLS handshake: 0x16
		c.isReadingTLS = true
		return n, nil
	}

	// Looks like HTTP.
	// len(b) == 576
	defer c.Conn.Close()

	// Read request
	r, err := c.readRequest(b, n)
	if err != nil {
		c.logf("hlfhr: Read request error from %s: %v", c.Conn.RemoteAddr(), err)
		return 0, ErrHttpOnHttpsPort
	}

	// Response
	w := newResponse(c.Conn)

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

	return 0, ErrHttpOnHttpsPort
}
