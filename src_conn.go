package hlfhr

import (
	"errors"
	"log"
	"net"
	"net/http"
)

var ErrHttpOnHttpsPort = errors.New("client sent an HTTP request to an HTTPS server")

type conn struct {
	net.Conn
	// Reading TLS if nil
	l *Listener
}

func (c *conn) log(v ...any) {
	if c.l.srv != nil && c.l.srv.ErrorLog != nil {
		c.l.srv.ErrorLog.Print(v...)
	} else {
		log.Print(v...)
	}
}

func (c *conn) readRequest(b []byte, n int) (req *http.Request, errStr string) {
	rd := &MaxHeaderBytesReader{Rd: c.Conn}
	rd.SetMax(c.l.srv)
	rd.Max -= n

	br := NewBufioReaderWithBytes(b, n, rd)

	req, err := http.ReadRequest(br)
	if err != nil {
		return nil, err.Error()
	}
	if req.Host == "" {
		return nil, "missing required Host header"
	}

	rd.SetReadingBody()
	return req, ""
}

func (c *conn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if c.l == nil || err != nil || n <= 0 {
		return n, err
	}

	if !ConnFirstByteLooksLikeHttp(b[0]) {
		// Not looks like HTTP.
		// TLS handshake: 0x16
		c.l = nil
		return n, nil
	}

	// Looks like HTTP.
	// len(b) == 576
	defer c.Conn.Close()

	// Read request
	r, errStr := c.readRequest(b, n)
	if r == nil {
		c.log("hlfhr: Read request error from ", c.Conn.RemoteAddr(), ": ", errStr)
		return 0, ErrHttpOnHttpsPort
	}

	// Response
	w := newResponse()
	if c.l.HttpOnHttpsPortErrorHandler != nil {
		// Handler
		c.l.HttpOnHttpsPortErrorHandler.ServeHTTP(w, r)
	} else {
		// Redirect
		RedirectToHttps(w, r, 302)
	}

	// Write
	err = w.flush(c.Conn)
	if err != nil {
		c.log("hlfhr: Write error for ", c.Conn.RemoteAddr(), ": ", err)
	}
	return 0, ErrHttpOnHttpsPort
}
