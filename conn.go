package hlfhr

import (
	"errors"
	"fmt"
	"net"
	"regexp"
)

var ErrHttpOnHttpsPort = errors.New("hlfhr: Client sent an HTTP request to an HTTPS server")

var compiledRegexp_tlsRecordHeaderLooksLikeHTTP = regexp.MustCompile(`^(GET /|HEAD |POST |PUT /|OPTIO|DELET|CONNE|TRACE|PATCH)`)

type conn struct {
	net.Conn

	// HttpOnHttpsPortErrorHandler handles HTTP requests sent to an HTTPS port.
	hlfhr_HttpOnHttpsPortErrorHandler func(b []byte, conn net.Conn)

	// Default 4096
	hlfhr_readFirstRequestBytesLen int

	hlfhr_isNotFirstRead bool
}

func newConn(inner net.Conn, l *listener) net.Conn {
	c := &conn{
		Conn:                              inner,
		hlfhr_HttpOnHttpsPortErrorHandler: l.hlfhr_httpOnHttpsPortErrorHandler,
		hlfhr_readFirstRequestBytesLen:    l.hlfhr_readFirstRequestBytesLen,
		hlfhr_isNotFirstRead:              false,
	}
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
	if c.hlfhr_HttpOnHttpsPortErrorHandler != nil {
		c.hlfhr_HttpOnHttpsPortErrorHandler(b, c.Conn)
		return
	}

	resp := NewResponse(c.Conn)
	// 302 Found
	if host, path, ok := ReadReqHostPath(b); ok {
		resp.Redirect(302, fmt.Sprint("https://", host, path))
		return
	}
	// script
	resp.StatusCode = 400
	resp.SetContentType("text/html")
	resp.Write(
		"<!-- ", ErrHttpOnHttpsPort, " -->\n",
		"<script> location.protocol = 'https:' </script>\n",
	)
	return
}
