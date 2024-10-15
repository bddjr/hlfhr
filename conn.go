package hlfhr

import (
	"bufio"
	"errors"
	"log"
	"net"
	"net/http"
	"time"
)

var ErrHttpOnHttpsPort = errors.New("hlfhr: Client sent an HTTP request to an HTTPS server")
var ErrMissingRequiredHostHeader = errors.New("missing required Host header")

// var ErrUnsupportedProtocolVersion = errors.New("unsupported protocol version")

type conn struct {
	isNotFirstRead bool
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
	chhr := connHttpHeaderReader{
		firstByte: firstByte,
		c:         c,
	}
	chhr.resetMaxHeaderBytes()
	c.setReadHeaderTimeout()

	r, err := http.ReadRequest(bufio.NewReader(&chhr))
	if err == nil && r.Host == "" {
		err = ErrMissingRequiredHostHeader
	}
	if err != nil {
		c.logf("hlfhr: Read request error from %s: %v", c.Conn.RemoteAddr(), err)
		return
	}

	// Response
	w := NewResponseWriter(c.Conn, nil)

	if h := c.l.HttpOnHttpsPortErrorHandler; h != nil {
		// Handler
		w.Resp.Request = r
		r.Response = w.Resp
		chhr.isReadingBody = true
		c.setReadTimeout()
		h.ServeHTTP(w, r)
		w.Resp.Close = true
	} else {
		// Redirect
		RedirectToHttps(w, r, 302)
	}

	// Write
	c.setWriteTimeout()
	err = w.Flush()
	if err != nil {
		c.logf("hlfhr: Write error for %s: %v", c.Conn.RemoteAddr(), err)
	}
}

func (c *conn) Read(b []byte) (int, error) {
	if c.isNotFirstRead || len(b) == 0 {
		return c.Conn.Read(b)
	}

	// Read 1 Byte Header
	n, err := c.Conn.Read(b[:1])
	if err != nil || n == 0 {
		return 0, err
	}

	c.isNotFirstRead = true

	switch b[0] {
	case 'G', 'H', 'P', 'O', 'D', 'C', 'T':
		// Looks like HTTP.
		// GET, HEAD, POST PUT PATCH, OPTIONS, DELETE, CONNECT, TRACE
		c.isHandlingHttp = true
		go c.handleHttp(b[0])
		return 0, ErrHttpOnHttpsPort
	}

	// Not looks like HTTP.
	// TLS handshake: 0x16
	if len(b) > 1 {
		n, err = c.Conn.Read(b[1:])
		n++
	}
	return n, err
}

func (c *conn) Close() error {
	if c.isHandlingHttp {
		return nil
	}
	return c.Conn.Close()
}
