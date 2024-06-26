package hlfhr

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

var ErrHttpOnHttpsPort = errors.New("hlfhr: Client sent an HTTP request to an HTTPS server")

type conn struct {
	net.Conn

	httpServer                  *http.Server
	httpOnHttpsPortErrorHandler http.Handler

	maxHeaderBytes int

	isNotFirstRead      bool
	isReadingHttpHeader bool
	httpHeaderByteBuf   byte
}

func IsMyConn(inner net.Conn) bool {
	_, ok := inner.(*conn)
	return ok
}

func NewConn(
	inner net.Conn,
	httpServer *http.Server,
	httpOnHttpsPortErrorHandler http.Handler,
) net.Conn {
	c, ok := inner.(*conn)
	if !ok {
		c = &conn{Conn: inner}
	}
	c.httpServer = httpServer
	c.httpOnHttpsPortErrorHandler = httpOnHttpsPortErrorHandler
	return c
}

func (c *conn) resetMaxHeaderBytes() {
	if c.httpServer != nil && c.httpServer.MaxHeaderBytes != 0 {
		c.maxHeaderBytes = c.httpServer.MaxHeaderBytes
	} else {
		c.maxHeaderBytes = http.DefaultMaxHeaderBytes
	}
}

func (c *conn) setKeepAliveTimeout() error {
	if c.httpServer != nil && c.httpServer.IdleTimeout > 0 {
		return c.Conn.SetReadDeadline(time.Now().Add(c.httpServer.IdleTimeout))
	}
	return c.setReadHeaderTimeout()
}

func (c *conn) setReadHeaderTimeout() error {
	if c.httpServer != nil && c.httpServer.ReadHeaderTimeout > 0 {
		return c.Conn.SetReadDeadline(time.Now().Add(c.httpServer.ReadHeaderTimeout))
	}
	return c.setReadTimeout()
}

func (c *conn) setReadTimeout() error {
	if c.httpServer != nil && c.httpServer.ReadTimeout > 0 {
		return c.Conn.SetReadDeadline(time.Now().Add(c.httpServer.ReadTimeout))
	}
	return c.Conn.SetReadDeadline(time.Time{})
}

func (c *conn) setWriteTimeout() error {
	if c.httpServer != nil && c.httpServer.WriteTimeout > 0 {
		return c.Conn.SetWriteDeadline(time.Now().Add(c.httpServer.WriteTimeout))
	}
	return c.Conn.SetWriteDeadline(time.Time{})
}

func (c *conn) Read(b []byte) (int, error) {
	if c.isNotFirstRead {
		if !c.isReadingHttpHeader {
			return c.Conn.Read(b)
		}
		if c.httpHeaderByteBuf > 0 {
			b[0] = c.httpHeaderByteBuf
			c.httpHeaderByteBuf = 0
			return 1, nil
		}
		if c.maxHeaderBytes <= 0 {
			return 0, http.ErrHeaderTooLong
		}
		if len(b) > c.maxHeaderBytes {
			b = b[:c.maxHeaderBytes]
		}
		n, err := c.Conn.Read(b)
		c.maxHeaderBytes -= n
		return n, err
	}
	c.isNotFirstRead = true
	c.resetMaxHeaderBytes()

	// Read 1 Byte Header
	rb1n, err := c.Conn.Read(b[:1])
	if err != nil || rb1n == 0 {
		return rb1n, err
	}
	c.maxHeaderBytes -= rb1n

	switch b[0] {
	case 'G', 'H', 'P', 'O', 'D', 'C', 'T':
		// Looks like HTTP.
	default:
		// Not looks like HTTP.
		// TLS handshake: 0x16
		return rb1n, nil
	}

	// HTTP/1.1
	defer c.Conn.Close()
	c.httpHeaderByteBuf = b[0]
	c.setReadHeaderTimeout()
	for {
		// Read HTTP header
		c.isReadingHttpHeader = true
		r, err := http.ReadRequest(bufio.NewReader(c))
		c.isReadingHttpHeader = false
		if err != nil {
			return 0, fmt.Errorf("hlfhr read: %s", err)
		}

		// Response
		w := NewResponseWriter(c.Conn, nil)
		w.Resp.Request = r
		r.Response = w.Resp
		c.setReadTimeout()

		if r.Host == "" {
			// Missing Host header
			w.WriteHeader(400)
			io.WriteString(w, "Missing Host header.")
		} else if c.httpOnHttpsPortErrorHandler != nil {
			// Handler
			c.httpOnHttpsPortErrorHandler.ServeHTTP(w, r)
		} else {
			// Redirect
			http.Redirect(w, r, "https://"+r.Host+r.URL.RequestURI(), http.StatusFound)
		}

		// Write
		c.setWriteTimeout()
		if err := w.Finish(); err != nil {
			return 0, fmt.Errorf("hlfhr write: %s", err)
		}
		if w.Resp.Close || w.Header().Get("Connection") == "close" {
			// Close.
			return 0, ErrHttpOnHttpsPort
		}

		// Keep Alive
		c.resetMaxHeaderBytes()
		c.setKeepAliveTimeout()
	}
}
