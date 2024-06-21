package hlfhr

import (
	"errors"
	"fmt"
	"net"
)

// hlfhr: Client sent an HTTP request to an HTTPS server
var ErrHttpOnHttpsPort = errors.New("hlfhr: Client sent an HTTP request to an HTTPS server")

type HttpOnHttpsPortErrorHandler func(rb []byte, conn net.Conn)

type conn struct {
	net.Conn

	// Default 4096
	hlfhr_readFirstRequestBytesLen int

	// HttpOnHttpsPortErrorHandler handles HTTP requests sent to an HTTPS port.
	hlfhr_httpOnHttpsPortErrorHandler HttpOnHttpsPortErrorHandler

	hlfhr_isNotFirstRead bool
}

func IsMyConn(inner net.Conn) bool {
	_, ok := inner.(*conn)
	return ok
}

func NewConn(inner net.Conn, readFirstRequestBytesLen int, httpOnHttpsPortErrorHandler HttpOnHttpsPortErrorHandler) net.Conn {
	c, ok := inner.(*conn)
	if !ok {
		c = &conn{Conn: inner}
	}
	c.hlfhr_readFirstRequestBytesLen = readFirstRequestBytesLen
	if c.hlfhr_readFirstRequestBytesLen == 0 {
		c.hlfhr_readFirstRequestBytesLen = 4096
	}
	c.hlfhr_httpOnHttpsPortErrorHandler = httpOnHttpsPortErrorHandler
	return c
}

func (c *conn) Read(b []byte) (n int, err error) {
	if c.hlfhr_isNotFirstRead {
		return c.Conn.Read(b)
	}
	c.hlfhr_isNotFirstRead = true

	// Default 576 Bytes
	if len(b) <= 1 {
		// Never run this
		return c.Conn.Read(b)
	}

	// Read 1 Byte Header
	_, err = c.Conn.Read(b[:1])
	if err != nil {
		return
	}

	switch b[0] {
	case 0x16:
		// Looks like TLS handshake.
		n, err = c.Conn.Read(b[1:])
		if err == nil {
			n += 1
		}
		return
	case 'G', 'H', 'P', 'O', 'D', 'C', 'T':
		// Looks like HTTP.
	default:
		return 0, fmt.Errorf("hlfhr: First byte %#02x does not looks like TLS handshake or HTTP request", b[0])
	}

	// HTTP Read 4096 Bytes Cache for redirect
	if c.hlfhr_readFirstRequestBytesLen > len(b) {
		nb := make([]byte, c.hlfhr_readFirstRequestBytesLen)
		nb[0] = b[0]
		b = nb
	}
	bn, err := c.Conn.Read(b[1:])
	if err != nil {
		return
	}
	b = b[:bn]

	err = ErrHttpOnHttpsPort

	// handler
	if c.hlfhr_httpOnHttpsPortErrorHandler != nil {
		c.hlfhr_httpOnHttpsPortErrorHandler(b, c.Conn)
		return
	}

	resp := NewResponse(c.Conn)
	// 302 Found
	if host, path, ok := ReadReqHostPath(b); ok {
		resp.Redirect(302, fmt.Sprint("https://", host, path))
		return
	}
	// script
	resp.ScriptRedirect()
	return
}
