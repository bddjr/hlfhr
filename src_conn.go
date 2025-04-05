package hlfhr

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"unsafe"
)

type conn struct {
	net.Conn
	tc  *tls.Conn // reading tls if nil
	srv *Server
}

func (c *conn) log(v ...any) {
	c.srv.log(v...)
}

func (c *conn) readRequest(b []byte, n int) (*http.Request, error) {
	rd := &io.LimitedReader{
		R: c.Conn,
		N: http.DefaultMaxHeaderBytes,
	}
	if c.srv.MaxHeaderBytes != 0 {
		rd.N = int64(c.srv.MaxHeaderBytes)
	}
	rd.N -= int64(n)

	br := NewBufioReaderWithBytes(b, n, rd)

	req, err := http.ReadRequest(br)
	if err == nil {
		BufioSetReader(br, c.Conn)
	}
	return req, err
}

func (c *conn) serve(b []byte, n int) {
	// Read request
	r, err := c.readRequest(b, n)
	if err != nil {
		c.log("hlfhr: Read request error from ", c.Conn.RemoteAddr(), ": ", err)
		return
	}

	// Response
	w := NewResponse(c.Conn, true)
	if r.Host == "" {
		// Error: missing HTTP/1.1 required "Host" header
		w.WriteHeader(400)
		w.WriteString("missing required Host header")
	} else if c.srv.HttpOnHttpsPortErrorHandler != nil {
		// Handler
		c.srv.HttpOnHttpsPortErrorHandler.ServeHTTP(w, r)
	} else {
		// Redirect
		RedirectToHttps(w, r, defaultRedirectStatus)
	}

	// Write
	err = w.FlushError()
	if err != nil {
		c.log("hlfhr: Write error for ", c.Conn.RemoteAddr(), ": ", err)
	}
}

func (c *conn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if c.tc == nil || err != nil || n <= 0 {
		return n, err
	}

	if !ConnFirstByteLooksLikeHttp(b[0]) {
		// Not looks like HTTP.
		(*struct{ conn net.Conn })(unsafe.Pointer(c.tc)).conn = c.Conn
		c.tc = nil
		return n, nil
	}

	// Looks like HTTP.
	// len(b) == 576
	c.serve(b, n)
	panic(http.ErrAbortHandler)
}
