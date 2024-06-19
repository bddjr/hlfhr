package hlfhr

import (
	"errors"
	"fmt"
	"net"
)

type Listener struct {
	net.Listener

	// HttpOnHttpsPortErrorHandler handles HTTP requests sent to an HTTPS port.
	hlfhr_httpOnHttpsPortErrorHandler func(b []byte, conn net.Conn)

	// Default 4096 Bytes
	hlfhr_readFirstRequestBytesLen int
}

func NewListener(inner net.Listener, srv *Server) net.Listener {
	var l *Listener
	if innerThisListener, ok := inner.(*Listener); ok {
		l = innerThisListener
	} else {
		l = &Listener{
			Listener: inner,
		}
	}
	l.hlfhr_httpOnHttpsPortErrorHandler = srv.Hlfhr_HttpOnHttpsPortErrorHandler
	l.hlfhr_readFirstRequestBytesLen = srv.Hlfhr_ReadFirstRequestBytesLen
	if l.hlfhr_readFirstRequestBytesLen == 0 {
		l.hlfhr_readFirstRequestBytesLen = 4096
	}
	return l
}

func (l *Listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	// Hijacking net.Conn
	return newConn(c, l), nil
}

type conn struct {
	net.Conn

	// HttpOnHttpsPortErrorHandler handles HTTP requests sent to an HTTPS port.
	hlfhr_HttpOnHttpsPortErrorHandler func(b []byte, conn net.Conn)

	// Default 4096
	hlfhr_readFirstRequestBytesLen int

	hlfhr_isNotFirstRead bool
}

func newConn(inner net.Conn, l *Listener) net.Conn {
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
	if c.hlfhr_readFirstRequestBytesLen < len(b) {
		c.hlfhr_readFirstRequestBytesLen = len(b)
	}

	// Read 5 Bytes Header
	rb5n := 0
	rb5 := b[:5]
	for rb5n < 5 {
		frb5n, err := c.Conn.Read(rb5[rb5n:])
		if err != nil {
			return 0, err
		}
		rb5n += frb5n
	}

	if !compiledRegexp_tlsRecordHeaderLooksLikeHTTP.Match(rb5) {
		// HTTPS
		n, err = c.Conn.Read(b[rb5n:])
		n += rb5n
		return
	}

	// HTTP Read 4096 Bytes Cache for redirect
	rbAll := append(b, make([]byte, c.hlfhr_readFirstRequestBytesLen-len(b))...)
	rbAlln, err := c.Conn.Read(rbAll[rb5n:])
	if err != nil {
		return 0, err
	}
	rbAll = rbAll[:rbAlln]

	// Write and Close
	defer c.Close()
	err = errors.New("hlfhr: Client sent an HTTP request to an HTTPS server")
	// handler
	if c.hlfhr_HttpOnHttpsPortErrorHandler != nil {
		c.hlfhr_HttpOnHttpsPortErrorHandler(rbAll, c.Conn)
		return
	}
	// 302 Found
	if host, path, ok := ReadReqHostPath(rbAll); ok {
		fmt.Fprint(c.Conn,
			"HTTP/1.1 302 Found\r\n",
			"Location: https://", host, path, "\r\n",
			"Connection: close\r\n",
			"\r\n",
			"Redirect to HTTPS\n",
		)
		return
	}
	// script
	fmt.Fprint(c.Conn,
		"HTTP/1.1 400 Bad Request\r\n",
		"Content-Type: text/html\r\n",
		"Connection: close\r\n",
		"\r\n",
		"<script> location.protocol = 'https:' </script>\n",
	)
	return
}
