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

type Conn struct {
	net.Conn

	HttpServer                  *http.Server
	HttpOnHttpsPortErrorHandler http.Handler

	isNotFirstRead bool
	isHandlingHttp bool
}

func IsMyConn(inner net.Conn) bool {
	_, ok := inner.(*Conn)
	return ok
}

func NewConn(
	inner net.Conn,
	httpServer *http.Server,
	httpOnHttpsPortErrorHandler http.Handler,
) net.Conn {
	c, ok := inner.(*Conn)
	if !ok {
		c = &Conn{Conn: inner}
	}
	c.HttpServer = httpServer
	c.HttpOnHttpsPortErrorHandler = httpOnHttpsPortErrorHandler
	return c
}

func (c *Conn) logf(format string, args ...any) {
	if c.HttpServer.ErrorLog != nil {
		c.HttpServer.ErrorLog.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

func (c *Conn) setReadHeaderTimeout() error {
	if c.HttpServer != nil && c.HttpServer.ReadHeaderTimeout > 0 {
		return c.Conn.SetReadDeadline(time.Now().Add(c.HttpServer.ReadHeaderTimeout))
	}
	return c.setReadTimeout()
}

func (c *Conn) setReadTimeout() error {
	if c.HttpServer != nil && c.HttpServer.ReadTimeout > 0 {
		return c.Conn.SetReadDeadline(time.Now().Add(c.HttpServer.ReadTimeout))
	}
	return c.Conn.SetReadDeadline(time.Time{})
}

func (c *Conn) setWriteTimeout() error {
	if c.HttpServer != nil && c.HttpServer.WriteTimeout > 0 {
		return c.Conn.SetWriteDeadline(time.Now().Add(c.HttpServer.WriteTimeout))
	}
	return c.Conn.SetWriteDeadline(time.Time{})
}

func (c *Conn) handleHttp(rBuf *bufio.Reader, chhr *connHttpHeaderReader) {
	// HTTP/1.1
	defer c.Conn.Close()

	// Read HTTP header
	c.setReadHeaderTimeout()
	chhr.isReadingHttpHeader = true
	r, err := http.ReadRequest(rBuf)
	chhr.isReadingHttpHeader = false
	if err != nil {
		c.logf("hlfhr: Read error from %s: %v", c.Conn.RemoteAddr(), err)
		return
	}
	if r.Host == "" {
		// Missing Host header
		return
	}

	// Response
	w := NewResponseWriter(c.Conn, nil)

	if c.HttpOnHttpsPortErrorHandler == nil {
		// Redirect
		RedirectToHttps(w, r, 302)
	} else {
		// Handler
		w.Resp.Request = r
		r.Response = w.Resp
		c.setReadTimeout()
		c.HttpOnHttpsPortErrorHandler.ServeHTTP(w, r)
		delete(w.Resp.Header, "Connection")
		delete(w.Resp.Header, "Keep-Alive")
		w.Resp.Close = true
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

func (c *Conn) Read(b []byte) (int, error) {
	if c.isNotFirstRead {
		if c.isHandlingHttp {
			c.logf("hlfhr isHandlingHttp read")
		}
		return c.Conn.Read(b)
	}
	c.isNotFirstRead = true

	chhr := &connHttpHeaderReader{
		c:          c.Conn,
		httpServer: c.HttpServer,
	}

	rBuf := bufio.NewReaderSize(chhr, len(b)) // Size: 576
	if len(b) < rBuf.Size() {
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

func (c *Conn) Close() error {
	if c.isHandlingHttp {
		return nil
	}
	return c.Conn.Close()
}
