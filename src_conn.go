package hlfhr

import (
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"time"
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

	w := new(Response)
	rd := &io.LimitedReader{R: c.Conn}
	if c.srv.MaxHeaderBytes != 0 {
		rd.N = int64(c.srv.MaxHeaderBytes - n)
	} else {
		rd.N = int64(http.DefaultMaxHeaderBytes - n)
	}

	br := NewBufioReaderWithBytes(b, n, rd)
	var req *http.Request
	isFirstReadHeader := true

	for {
		// Read request
		req, err = http.ReadRequest(br)
		if err != nil {
			if isFirstReadHeader {
				c.log("hlfhr: Read request error from ", c.Conn.RemoteAddr(), ": ", err)
			}
			break
		}
		isFirstReadHeader = false
		BufioSetReader(br, c.Conn)

		// Response
		w.Reset(c.Conn, br)
		w.header["Connection"] = []string{"close"}

		// set read timeout
		if c.srv.WriteTimeout > 0 {
			c.Conn.SetWriteDeadline(time.Now().Add(c.srv.WriteTimeout))
		} else {
			c.Conn.SetWriteDeadline(time.Time{})
		}

		if req.Host == "" {
			// missing "Host" header
			w.WriteHeader(400)
			w.WriteString("missing required Host header")
		} else if c.srv.HttpOnHttpsPortErrorHandler != nil {
			// Handler
			c.srv.HttpOnHttpsPortErrorHandler.ServeHTTP(w, req)
		} else {
			// Redirect
			RedirectToHttps(w, req, 302)
		}

		// Write
		err = w.FlushError()
		if err != nil {
			if err != http.ErrHijacked {
				c.log("hlfhr: Write error for ", c.Conn.RemoteAddr(), ": ", err)
			}
			break
		}

		// Keep Alive
		if !w.KeepAlive() {
			break
		}

		if c.srv.MaxHeaderBytes != 0 {
			rd.N = int64(c.srv.MaxHeaderBytes)
		} else {
			rd.N = int64(http.DefaultMaxHeaderBytes)
		}

		br.Reset(rd)

		// set read timeout
		if c.srv.IdleTimeout > 0 {
			c.Conn.SetReadDeadline(time.Now().Add(c.srv.IdleTimeout))
		} else if c.srv.ReadTimeout > 0 {
			c.Conn.SetReadDeadline(time.Now().Add(c.srv.ReadTimeout))
		} else {
			c.Conn.SetReadDeadline(time.Time{})
		}
	}
	return 0, ErrHttpOnHttpsPort
}
