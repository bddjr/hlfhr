package hlfhr

import (
	"crypto/tls"
	"net"
	"net/http"
	"runtime"
	"sync"
	"unsafe"

	hlfhr_utils "github.com/bddjr/hlfhr/utils"
)

type connPoolType struct {
	p sync.Pool
}

func (t *connPoolType) get() *conn {
	return t.p.Get().(*conn)
}

func (t *connPoolType) Put(x *conn) {
	t.p.Put(x)
}

var connPool = connPoolType{
	sync.Pool{
		New: func() any {
			return new(conn)
		},
	},
}

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
	limitedReader := hlfhr_utils.NewLimitedReader(c.Conn, http.DefaultMaxHeaderBytes)
	if c.srv.MaxHeaderBytes != 0 {
		limitedReader.N = int64(c.srv.MaxHeaderBytes)
	}
	limitedReader.N -= int64(n)

	br := hlfhr_utils.NewBufioReaderWithBytes(b, n, limitedReader)

	r, err := http.ReadRequest(br)
	hlfhr_utils.LimitedReaderPool.Put(limitedReader)
	if err != nil {
		c.srv.log("hlfhr: Read request error from ", c.Conn.RemoteAddr(), ": ", err)
		return
	}
	hlfhr_utils.BufioSetReader(br, c.Conn)

	// Response
	w := hlfhr_utils.NewResponse(c.Conn, 400, true)

	if r.Host == "" {
		// Error: missing HTTP/1.1 required "Host" header
		hlfhr_utils.BufioReaderPool.Put(br)
		w.WriteString("missing required Host header")
		defer c.putToPool(w)
	} else if c.srv.HlfhrHandler != nil {
		// Handler
		c.srv.HlfhrHandler.ServeHTTP(w, r)
	} else {
		// Redirect
		hlfhr_utils.BufioReaderPool.Put(br)
		if c.tc != nil {
			hlfhr_utils.RedirectToHttps_ForceSamePort(w, r, 307)
		} else {
			// Listen80RedirectTo443
			hlfhr_utils.RedirectToHttps(w, r, 307)
		}
		defer c.putToPool(w)
	}

	// Write
	err = w.FlushError()
	if err != nil {
		c.srv.log("hlfhr: Write error for ", c.Conn.RemoteAddr(), ": ", err)
	}
}

func (c *conn) putToPool(w *hlfhr_utils.Response) {
	hlfhr_utils.ResponsePool.Put(w)
	connPool.Put(c)
}
