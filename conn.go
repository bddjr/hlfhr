package hlfhr

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
)

var ErrHttpOnHttpsPort = errors.New("hlfhr: Client sent an HTTP request to an HTTPS server")
var ErrMissingHostHeader = errors.New("hlfhr: Missing Host header")

type HttpOnHttpsPortErrorHandler func(rb []byte, conn net.Conn)

type conn struct {
	net.Conn

	// Default http.DefaultMaxHeaderBytes
	maxHeaderBytes int

	// HttpOnHttpsPortErrorHandler handles HTTP requests sent to an HTTPS port.
	httpOnHttpsPortErrorHandler HttpOnHttpsPortErrorHandler

	isNotFirstRead      bool
	isReadingHttpHeader bool
	byteBuf             byte
}

func IsMyConn(inner net.Conn) bool {
	_, ok := inner.(*conn)
	return ok
}

func NewConn(inner net.Conn, maxHeaderBytes int, httpOnHttpsPortErrorHandler HttpOnHttpsPortErrorHandler) net.Conn {
	c, ok := inner.(*conn)
	if !ok {
		c = &conn{Conn: inner}
	}
	c.maxHeaderBytes = maxHeaderBytes
	if c.maxHeaderBytes == 0 {
		c.maxHeaderBytes = http.DefaultMaxHeaderBytes
	}
	c.httpOnHttpsPortErrorHandler = httpOnHttpsPortErrorHandler
	return c
}

func (c *conn) Read(b []byte) (n int, err error) {
	if c.isNotFirstRead {
		if !c.isReadingHttpHeader {
			return c.Conn.Read(b)
		}
		if c.byteBuf > 0 {
			b[0] = c.byteBuf
			c.byteBuf = 0
			return 1, nil
		}
		if c.maxHeaderBytes <= 0 {
			return 0, http.ErrHeaderTooLong
		}
		if len(b) > c.maxHeaderBytes {
			b = b[:c.maxHeaderBytes]
		}
		n, err = c.Conn.Read(b)
		c.maxHeaderBytes -= n
		return
	}
	c.isNotFirstRead = true

	// b Default 576 Bytes

	// Read 1 Byte Header
	rb1n, err := c.Conn.Read(b[:1])
	if err != nil || rb1n == 0 {
		return
	}
	c.maxHeaderBytes -= 1

	switch b[0] {
	case 'G', 'H', 'P', 'O', 'D', 'C', 'T':
		// Looks like HTTP.
	default:
		// Not looks like HTTP.
		// TLS handshake: 0x16
		return rb1n, nil
	}

	// HTTP Read 4096 Bytes Cache for redirect
	c.byteBuf = b[0]
	c.isReadingHttpHeader = true
	req, err := http.ReadRequest(bufio.NewReader(c))
	c.isReadingHttpHeader = false
	if err != nil {
		return 0, fmt.Errorf("hlfhr: %s", err)
	}
	if req.Host == "" {
		return 0, ErrMissingHostHeader
	}

	err = ErrHttpOnHttpsPort

	// handler
	if c.httpOnHttpsPortErrorHandler != nil {
		c.httpOnHttpsPortErrorHandler(b, c.Conn)
		return
	}

	resp := NewResponse(c.Conn)
	resp.Request = req

	// 302 Found
	resp.Redirect(302, "https://"+req.Host+req.URL.RequestURI())
	return
}
