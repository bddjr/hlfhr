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

	// b Default 576 Bytes

	// Read 1 Byte Header
	_, err = c.Conn.Read(b[:1])
	if err != nil {
		return
	}

	switch b[0] {
	case 'G', 'H', 'P', 'O', 'D', 'C', 'T':
		// Looks like HTTP.
	default:
		// Not looks like HTTP.
		// TLS handshake: 0x16
		if len(b) > 1 {
			n, err = c.Conn.Read(b[1:])
		}
		if err == nil {
			n += 1
		}
		return
	}

	// HTTP Read 4096 Bytes Cache for redirect
	if c.hlfhr_readFirstRequestBytesLen > len(b) {
		nb := make([]byte, c.hlfhr_readFirstRequestBytesLen)
		nb[0] = b[0]
		b = nb
	}

	if c.hlfhr_readFirstRequestBytesLen > 1 {
		bn, err := c.Conn.Read(b[1:c.hlfhr_readFirstRequestBytesLen])
		if err != nil {
			return 0, err
		}
		bn += 1
		b = b[:bn]
	} else {
		b = b[:max(c.hlfhr_readFirstRequestBytesLen, 1)]
	}

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
