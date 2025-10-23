package hlfhr

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"runtime"
	"unsafe"

	hlfhr_utils "github.com/bddjr/hlfhr/utils"
)

type conn struct {
	net.Conn
	tc  *tls.Conn // If nil, it's reading TLS or serving port 80
	srv *Server
}

func (c *conn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if c.tc == nil || err != nil || n <= 0 {
		return n, err
	}

	// Does the first byte look like HTTP?
	switch b[0] {
	case 22, // recordTypeHandshake
		20, // recordTypeChangeCipherSpec
		21, // recordTypeAlert
		23: // recordTypeApplicationData
		// TLS

	case 'G', // GET
		'H', // HEAD
		'P', // POST PUT PATCH
		'D', // DELETE
		'C', // CONNECT
		'O', // OPTIONS
		'T': // TRACE
		// HTTP
		// len(b) == 576
		c.serve(b, n)
		panic(http.ErrAbortHandler)
	}

	// Cancel hijack
	(*struct{ conn net.Conn })(unsafe.Pointer(c.tc)).conn = c.Conn
	c.tc = nil
	return n, nil
}

func (c *conn) serve(b []byte, n int) {
	defer func() {
		if err := recover(); err != nil && err != http.ErrAbortHandler {
			buf := make([]byte, 64<<10)
			buf = buf[:runtime.Stack(buf, false)]
			c.srv.logf("hlfhr: panic serving %s: %v\n%s", c.RemoteAddr(), err, buf)
		}
	}()

	// Read request
	limitedReader := &io.LimitedReader{
		R: c.Conn,
		N: http.DefaultMaxHeaderBytes,
	}
	if c.srv.MaxHeaderBytes != 0 {
		limitedReader.N = int64(c.srv.MaxHeaderBytes)
	}
	limitedReader.N -= int64(n)

	br := hlfhr_utils.NewBufioReaderWithBytes(b, n, limitedReader)

	r, err := http.ReadRequest(br)
	if err != nil {
		c.srv.log("hlfhr: Read request error from ", c.Conn.RemoteAddr(), ": ", err)
		return
	}
	hlfhr_utils.BufioSetReader(br, c.Conn)

	// Response
	w := hlfhr_utils.NewResponse(c.Conn, 400, true)

	if r.Host == "" {
		// Error: missing HTTP/1.1 required "Host" header
		w.WriteString("missing required Host header")
	} else if c.srv.HlfhrHandler != nil {
		// Handler
		c.srv.HlfhrHandler.ServeHTTP(w, r)
	} else {
		// Redirect
		if c.tc != nil {
			hlfhr_utils.RedirectToHttps_ForceSamePort(w, r, 307)
		} else {
			// Listen80RedirectTo443
			hlfhr_utils.RedirectToHttps(w, r, 307)
		}
	}

	// Write
	err = w.FlushError()
	if err != nil {
		c.srv.log("hlfhr: Write error for ", c.Conn.RemoteAddr(), ": ", err)
	}
}
