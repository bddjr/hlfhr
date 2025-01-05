package hlfhr

import (
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"unsafe"
)

var ErrHttpOnHttpsPort = errors.New("client sent an HTTP request to an HTTPS server")

type conn struct {
	net.Conn
	tc  *tls.Conn // reading tls if nil
	srv *Server
}

func (c *conn) log(v ...any) {
	if c.srv.ErrorLog != nil {
		c.srv.ErrorLog.Print(v...)
	} else {
		log.Print(v...)
	}
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

func (c *conn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if c.tc == nil || err != nil || n <= 0 {
		return n, err
	}

	if !ConnFirstByteLooksLikeHttp(b[0]) {
		// Not looks like HTTP.
		// TLS handshake: 0x16
		(*struct{ conn net.Conn })(unsafe.Pointer(c.tc)).conn = c.Conn
		c.tc = nil
		return n, nil
	}

	// Looks like HTTP.
	// len(b) == 576

	// Read request
	r, err := c.readRequest(b, n)
	if err != nil {
		c.log("hlfhr: Read request error from ", c.Conn.RemoteAddr(), ": ", err)
		return 0, ErrHttpOnHttpsPort
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
		RedirectToHttps(w, r, 302)
	}

	// Write
	err = w.FlushError()
	if err != nil {
		c.log("hlfhr: Write error for ", c.Conn.RemoteAddr(), ": ", err)
	}
	return 0, ErrHttpOnHttpsPort
}
