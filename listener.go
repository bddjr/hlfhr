// HTTPS Listener For HTTP Redirect
//
// Adapted from net/http
//
// BSD-3-clause license
package hlfhr

import (
	"bytes"
	"fmt"
	"net"
)

type Listener struct {
	net.Listener

	// HttpOnHttpsPortErrorHandler handles HTTP requests sent to an HTTPS port.
	hlfhr_httpOnHttpsPortErrorHandler func(b []byte, conn net.Conn)

	// Default 4096 Bytes
	hflhr_readFirstRequestBytesLen int
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
	l.hlfhr_httpOnHttpsPortErrorHandler = srv.Hflhr_HttpOnHttpsPortErrorHandler
	l.hflhr_readFirstRequestBytesLen = srv.Hflhr_ReadFirstRequestBytesLen
	if l.hflhr_readFirstRequestBytesLen == 0 {
		l.hflhr_readFirstRequestBytesLen = 4096
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
	hflhr_readFirstRequestBytesLen int

	isNotFirstRead            bool
	firstReadBytesForRedirect []byte
	firstReadBN               int
	isWritten                 bool
}

func newConn(inner net.Conn, l *Listener) net.Conn {
	c := &conn{
		Conn:                              inner,
		hlfhr_HttpOnHttpsPortErrorHandler: l.hlfhr_httpOnHttpsPortErrorHandler,
		hflhr_readFirstRequestBytesLen:    l.hflhr_readFirstRequestBytesLen,
		isNotFirstRead:                    false,
		firstReadBytesForRedirect:         nil,
	}
	return c
}

func (c *conn) Read(b []byte) (n int, err error) {
	if c.isNotFirstRead {
		if c.firstReadBytesForRedirect != nil {
			// In theory, this code should never be run.
			// If it runs, the standard library has been changed.
			if c.firstReadBN < len(c.firstReadBytesForRedirect) {
				n = copy(b, c.firstReadBytesForRedirect[c.firstReadBN:])
				c.firstReadBN += n
				if c.firstReadBN >= len(c.firstReadBytesForRedirect) {
					c.firstReadBytesForRedirect = nil
				}
				return
			}
			c.firstReadBytesForRedirect = nil
		}
		return c.Conn.Read(b)
	}
	c.isNotFirstRead = true

	// Default 576 Bytes
	if len(b) <= 5 {
		// In theory, this code should never be run.
		return c.Conn.Read(b)
	}
	if len(b) >= c.hflhr_readFirstRequestBytesLen {
		n, err = c.Conn.Read(b)
		if err == nil && compiledRegexp_tlsRecordHeaderLooksLikeHTTP.Match(b) {
			// HTTP Cache for redirect
			c.firstReadBytesForRedirect = b[:n]
			c.firstReadBN = n
		}
		return
	}

	// Read 5 Bytes Header
	rb5 := b[:5]
	rb5n, err := c.Conn.Read(rb5)
	if err != nil {
		return 0, err
	}
	rb5 = rb5[:rb5n]

	if !compiledRegexp_tlsRecordHeaderLooksLikeHTTP.Match(rb5) {
		// HTTPS
		n, err = c.Conn.Read(b[rb5n:])
		n += rb5n
		return
	}

	// HTTP Read 4096 Bytes Cache for redirect
	rbAll := append(b, make([]byte, c.hflhr_readFirstRequestBytesLen-len(b))...)
	rbAlln, err := c.Conn.Read(rbAll[rb5n:])
	if err != nil {
		return 0, err
	}
	rbAll = rbAll[:rbAlln]

	n = copy(b[rb5n:], rbAll[rb5n:])
	n += rb5n
	// Cache for redirect
	c.firstReadBytesForRedirect = rbAll
	c.firstReadBN = n
	return
}

// Hijacking the Write function to achieve redirection
func (c *conn) Write(b []byte) (n int, err error) {
	if !c.isWritten {
		c.isWritten = true
		if c.firstReadBytesForRedirect != nil && bytes.Equal(b, []byte("HTTP/1.0 400 Bad Request\r\n\r\nClient sent an HTTP request to an HTTPS server.\n")) {
			defer func() {
				c.firstReadBytesForRedirect = nil
			}()
			n = len(b)
			// handler
			if c.hlfhr_HttpOnHttpsPortErrorHandler != nil {
				c.hlfhr_HttpOnHttpsPortErrorHandler(c.firstReadBytesForRedirect, c.Conn)
				return
			}
			// 302 Found
			host, path, ok := ReadReqHostPath(c.firstReadBytesForRedirect)
			if ok {
				c.Conn.Write([]byte(fmt.Sprint("HTTP/1.1 302 Found\r\nLocation: https://", host, path, "\r\nConnection: close\r\n\r\nRedirect to HTTPS\n")))
				return
			}
			// script
			c.Conn.Write([]byte("HTTP/1.1 400 Bad Request\r\nContent-Type: text/html\r\nConnection: close\r\n\r\n<script> location.protocol = 'https:' </script>\n"))
			return
		}
	}
	return c.Conn.Write(b)
}
