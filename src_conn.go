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

type Conn struct {
	net.Conn
	TLSConn *tls.Conn // If nil, it's reading TLS or serving port 80
	Server  *Server
}

func (c *Conn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if c.TLSConn == nil || err != nil || n <= 0 {
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
		c.HlfhrServe(b, n)
		panic(http.ErrAbortHandler)
	}

	// Cancel hijack
	(*struct{ conn net.Conn })(unsafe.Pointer(c.TLSConn)).conn = c.Conn
	c.TLSConn = nil
	return n, nil
}

func (c *Conn) HlfhrServe(b []byte, n int) {
	defer func() {
		if err := recover(); err != nil && err != http.ErrAbortHandler {
			buf := make([]byte, 64<<10)
			buf = buf[:runtime.Stack(buf, false)]
			c.Server.logf("hlfhr: panic serving %s: %v\n%s", c.RemoteAddr(), err, buf)
		}
	}()

	// Read request
	limitedReader := &io.LimitedReader{
		R: c.Conn,
		N: http.DefaultMaxHeaderBytes,
	}
	if c.Server.MaxHeaderBytes != 0 {
		limitedReader.N = int64(c.Server.MaxHeaderBytes)
	}
	limitedReader.N -= int64(n)

	br := hlfhr_utils.NewBufioReaderWithBytes(b, n, limitedReader)

	r, err := http.ReadRequest(br)
	if err != nil {
		c.Server.log("hlfhr: Read request error from ", c.Conn.RemoteAddr(), ": ", err)
		return
	}
	hlfhr_utils.BufioSetReader(br, c.Conn)

	// Response
	w := hlfhr_utils.NewResponse(c.Conn, 400, true)

	if r.Host == "" {
		// Error: missing HTTP/1.1 required "Host" header
		w.WriteString("missing required Host header")
	} else if c.Server.HlfhrHandler != nil {
		// Handler
		c.Server.HlfhrHandler.ServeHTTP(w, r)
	} else {
		// Redirect
		if c.TLSConn != nil {
			hlfhr_utils.RedirectToHttps_ForceSamePort(w, r, 307)
		} else {
			// Listen80RedirectTo443
			hlfhr_utils.RedirectToHttps(w, r, 307)
		}
	}

	// Write
	err = w.FlushError()
	if err != nil {
		c.Server.log("hlfhr: Write error for ", c.Conn.RemoteAddr(), ": ", err)
	}
}
