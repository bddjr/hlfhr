package hlfhr

import (
	"bufio"
	"errors"
	"log"
	"net"
	"net/http"
	"time"
)

var ErrHttpOnHttpsPort = errors.New("client sent an HTTP request to an HTTPS server")

type conn struct {
	isReadingTLS   bool
	isHandlingHttp bool
	net.Conn
	l *Listener
}

func IsMyConn(inner net.Conn) bool {
	_, ok := inner.(*conn)
	return ok
}

func (c *conn) logf(format string, args ...any) {
	srv := c.l.HttpServer
	if srv != nil {
		el := srv.ErrorLog
		if el != nil {
			el.Printf(format, args...)
			return
		}
	}
	log.Printf(format, args...)
}

func (c *conn) setReadHeaderTimeout() error {
	srv := c.l.HttpServer
	if srv != nil {
		t := srv.ReadHeaderTimeout
		if t > 0 {
			return c.Conn.SetReadDeadline(time.Now().Add(t))
		}
	}
	return c.setReadTimeout()
}

func (c *conn) setReadTimeout() error {
	srv := c.l.HttpServer
	if srv != nil {
		t := srv.ReadTimeout
		if t > 0 {
			return c.Conn.SetReadDeadline(time.Now().Add(t))
		}
	}
	return c.Conn.SetReadDeadline(time.Time{})
}

func (c *conn) setWriteTimeout() error {
	srv := c.l.HttpServer
	if srv != nil {
		t := srv.WriteTimeout
		if t > 0 {
			return c.Conn.SetWriteDeadline(time.Now().Add(t))
		}
	}
	return c.Conn.SetWriteDeadline(time.Time{})
}

func (c *conn) handleHttp(firstByte byte) {
	// HTTP/1.1
	defer c.Conn.Close()

	// Read HTTP header
	chhr := connHttpHeaderReader{c: c}
	chhr.setStatusFirstByte(firstByte)
	c.setReadHeaderTimeout()

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
	c.setReadTimeout()
	c.setWriteTimeout()

	if h := c.l.HttpOnHttpsPortErrorHandler; h != nil {
		// Handler
		h.ServeHTTP(w, r)
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
	if c.isHandlingHttp {
		return 0, net.ErrClosed
	}

	// Read 1 Byte Header
	n, err := c.Conn.Read(b[:1])
	if err != nil || n == 0 {
		return 0, err
	}

	if ConnFirstByteLooksLikeHttp(b[0]) {
		// Looks like HTTP.
		c.isHandlingHttp = true
		go c.handleHttp(b[0])
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

func (c *conn) Close() error {
	if c.isHandlingHttp {
		return nil
	}
	return c.Conn.Close()
}
