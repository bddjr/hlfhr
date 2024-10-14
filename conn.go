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
	if el := c.l.HttpServer.ErrorLog; el != nil {
		el.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

func (c *conn) setReadHeaderTimeout() error {
	if srv := c.l.HttpServer; srv != nil {
		t := srv.ReadHeaderTimeout
		if t > 0 {
			return c.Conn.SetReadDeadline(time.Now().Add(t))
		}
	}
	return c.setReadTimeout()
}

func (c *conn) setReadTimeout() error {
	if srv := c.l.HttpServer; srv != nil {
		t := srv.ReadTimeout
		if t > 0 {
			return c.Conn.SetReadDeadline(time.Now().Add(t))
		}
	}
	return c.Conn.SetReadDeadline(time.Time{})
}

func (c *conn) setWriteTimeout() error {
	if srv := c.l.HttpServer; srv != nil {
		t := srv.WriteTimeout
		if t > 0 {
			return c.Conn.SetWriteDeadline(time.Now().Add(t))
		}
	}
	return c.Conn.SetWriteDeadline(time.Time{})
}

func (c *conn) handleHttp(rBuf *bufio.Reader, chhr *connHttpHeaderReader) {
	// HTTP/1.1
	defer c.Conn.Close()

	// Read HTTP header
	c.setReadHeaderTimeout()
	chhr.isReadingHttpHeader = true
	r, err := http.ReadRequest(rBuf)
	chhr.isReadingHttpHeader = false
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
		c.setReadTimeout()
		h.ServeHTTP(w, r)
		w.Resp.Close = true
	} else {
		// Redirect
		RedirectToHttps(w, r, 302)
	}

	// Write
	c.setWriteTimeout()
	if err := w.Flush(); err != nil {
		c.logf("hlfhr: Write error for %s: %v", c.Conn.RemoteAddr(), err)
		return
	}
	if w.Resp.Close {
		// Close
		return
	}
}

func (c *conn) Read(b []byte) (int, error) {
	if c.isNotFirstRead {
		return c.Conn.Read(b)
	}
	c.isNotFirstRead = true

	chhr := &connHttpHeaderReader{c: c}

	rBuf := bufio.NewReaderSize(chhr, len(b)) // Size: 576
	if len(b) != rBuf.Size() {
		// HTTPS should work even if the standard library is modified
		return c.Conn.Read(b)
	}

	// Read 1 Byte Header
	chhr.resetMaxHeaderBytes(len(b))
	chhr.isReadingHttpHeader = true
	rb1, err := rBuf.Peek(1)
	chhr.isReadingHttpHeader = false
	if err != nil {
		return 0, err
	}
	if len(rb1) >= 1 {
		switch rb1[0] {
		case 'G', 'H', 'P', 'O', 'D', 'C', 'T':
			// Looks like HTTP.
			c.isHandlingHttp = true
			go c.handleHttp(rBuf, chhr)
			return 0, ErrHttpOnHttpsPort
		}
	}

	// Not looks like HTTP.
	// TLS handshake: 0x16
	return rBuf.Read(b)
}

func (c *conn) Close() error {
	if c.isHandlingHttp {
		return nil
	}
	return c.Conn.Close()
}
