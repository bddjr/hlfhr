package hlfhr

import (
	"bufio"
	"errors"
	"log"
	"net"
	"net/http"
	"reflect"
)

var ErrHttpOnHttpsPort = errors.New("client sent an HTTP request to an HTTPS server")

type conn struct {
	isReadingTLS bool
	net.Conn
	l *Listener
}

func IsMyConn(inner net.Conn) bool {
	_, ok := inner.(*conn)
	return ok
}

func (c *conn) logf(format string, args ...any) {
	if c.l.HttpServer != nil && c.l.HttpServer.ErrorLog != nil {
		c.l.HttpServer.ErrorLog.Printf(format, args...)
		return
	}
	log.Printf(format, args...)
}

func (c *conn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if c.isReadingTLS || err != nil || n == 0 {
		return n, err
	}

	if !ConnFirstByteLooksLikeHttp(b[0]) {
		// Not looks like HTTP.
		// TLS handshake: 0x16
		c.isReadingTLS = true
		return n, nil
	}

	// Looks like HTTP.
	defer c.Conn.Close()

	chhr := &connHttpHeaderReader{c: c}
	chhr.setMax()

	// Buffer
	br := &bufio.Reader{}
	brElem := reflect.ValueOf(br).Elem()
	// fill buffer
	if len(b) < 16 {
		nb := make([]byte, 16)
		copy(nb, b)
		b = nb
	}
	*(*[]byte)(brElem.FieldByName("buf").Addr().UnsafePointer()) = b
	// fill reader
	br.Reset(chhr)
	// fill length
	*(*int)(brElem.FieldByName("w").Addr().UnsafePointer()) = len(b)

	// Read request
	r, err := http.ReadRequest(br)
	if err != nil {
		c.logf("hlfhr: Read request error from %s: %v", c.Conn.RemoteAddr(), err)
		return 0, ErrHttpOnHttpsPort
	}
	if r.Host == "" {
		c.logf("hlfhr: error form %s: missing required Host header", c.Conn.RemoteAddr())
		return 0, ErrHttpOnHttpsPort
	}

	// Response
	w := newResponse(c.Conn)
	chhr.setReadingBody()

	if c.l.HttpOnHttpsPortErrorHandler != nil {
		// Handler
		c.l.HttpOnHttpsPortErrorHandler.ServeHTTP(w, r)
	} else {
		// Redirect
		RedirectToHttps(w, r, 302)
	}

	// Write
	err = w.flush()
	if err != nil {
		c.logf("hlfhr: Write error for %s: %v", c.Conn.RemoteAddr(), err)
	}

	return 0, ErrHttpOnHttpsPort
}
