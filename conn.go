package hlfhr

import (
	"errors"
	"fmt"
	"net"
	"regexp"
)

// hlfhr: Client sent an HTTP request to an HTTPS server
var ErrHttpOnHttpsPort = errors.New("hlfhr: Client sent an HTTP request to an HTTPS server")

var compiledRegexp_tlsRecordHeaderLooksLikeHTTP = regexp.MustCompile(`^(GET /|HEAD |POST |PUT /|OPTIO|DELET|CONNE|TRACE|PATCH)`)

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
	if len(b) <= 5 {
		// Never run this
		return c.Conn.Read(b)
	}

	// Read 5 Bytes Header
	rb5n := 0
	rb5 := b[:5]
	for rb5n < 5 {
		n, err := c.Conn.Read(rb5[rb5n:])
		if err != nil {
			return 0, err
		}
		rb5n += n
	}

	if !compiledRegexp_tlsRecordHeaderLooksLikeHTTP.Match(rb5) {
		// HTTPS
		n, err = c.Conn.Read(b[rb5n:])
		if err == nil {
			n += rb5n
		}
		return
	}

	// HTTP Read 4096 Bytes Cache for redirect
	if c.hlfhr_readFirstRequestBytesLen > len(b) {
		b = append(b, make([]byte, c.hlfhr_readFirstRequestBytesLen-len(b))...)
	}
	bn, err := c.Conn.Read(b[rb5n:])
	if err != nil {
		return
	}
	b = b[:bn]

	// Write and Close
	defer c.Close()
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
