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

type Conn struct {
	net.Conn

	HttpServer                  *http.Server
	HttpOnHttpsPortErrorHandler http.Handler

	maxHeaderBytes int

	isNotFirstRead      bool
	isReadingHttpHeader bool
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

func (c *Conn) resetMaxHeaderBytes() {
	if c.HttpServer != nil && c.HttpServer.MaxHeaderBytes != 0 {
		c.maxHeaderBytes = c.HttpServer.MaxHeaderBytes
	} else {
		c.maxHeaderBytes = http.DefaultMaxHeaderBytes
	}
}

func (c *Conn) setKeepAliveTimeout() error {
	if c.HttpServer != nil && c.HttpServer.IdleTimeout > 0 {
		return c.Conn.SetReadDeadline(time.Now().Add(c.HttpServer.IdleTimeout))
	}
	return c.setReadTimeout()
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

func (c *Conn) Read(b []byte) (int, error) {
	if c.isNotFirstRead {
		if !c.isReadingHttpHeader {
			return c.Conn.Read(b)
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
	rBuf := bufio.NewReader(c)

	// Read 1 Byte Header
	{
		rb1, err := rBuf.Peek(1)
		if err != nil {
			return 0, err
		}
		switch rb1[0] {
		case 'G', 'H', 'P', 'O', 'D', 'C', 'T':
			// Looks like HTTP.
		default:
			// Not looks like HTTP.
			// TLS handshake: 0x16
			return rBuf.Read(b)
		}
	}

	// HTTP/1.1
	defer c.Conn.Close()
	c.setReadHeaderTimeout()
	for {
		// Read HTTP header
		c.resetMaxHeaderBytes()
		c.isReadingHttpHeader = true
		r, err := http.ReadRequest(rBuf)
		c.isReadingHttpHeader = false
		if err != nil {
			return 0, fmt.Errorf("hlfhr read: %s", err)
		}

		// Response
		w := NewResponseWriter(c.Conn, nil)
		w.Resp.Request = r
		r.Response = w.Resp

		if r.Host == "" {
			// Missing Host header
			w.WriteHeader(400)
			io.WriteString(w, "Missing Host header.")
		} else if c.HttpOnHttpsPortErrorHandler != nil {
			// Handler
			c.setReadTimeout()
			c.HttpOnHttpsPortErrorHandler.ServeHTTP(w, r)
			if w.Hijacked {
				// Close
				return 0, ErrHttpOnHttpsPort
			}
			if c.HttpServer.IdleTimeout != 0 && w.Header().Get("Connection") == "keep-alive" && w.Header().Get("Keep-Alive") == "" {
				w.Header().Set("Keep-Alive", fmt.Sprint("timeout=", c.HttpServer.IdleTimeout.Seconds()))
			}
		} else {
			// Redirect
			RedirectToHttps(w, r, 302)
		}

		// Write
		c.setWriteTimeout()
		if err := w.Flush(); err != nil {
			return 0, fmt.Errorf("hlfhr write: %s", err)
		}
		if w.Resp.Close || w.Header().Get("Connection") == "close" {
			// Close
			return 0, ErrHttpOnHttpsPort
		}

		// Keep Alive
		c.setKeepAliveTimeout()
	}
}
