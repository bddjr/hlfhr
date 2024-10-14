package hlfhr

import (
	"io"
	"net"
	"net/http"
)

type connHttpHeaderReader struct {
	isReadingHttpHeader bool
	c                   net.Conn
	max                 int
	httpServer          *http.Server
}

func (r *connHttpHeaderReader) resetMaxHeaderBytes(Min int) {
	if r.httpServer != nil && r.httpServer.MaxHeaderBytes != 0 {
		r.max = max(r.httpServer.MaxHeaderBytes, Min)
	} else {
		r.max = http.DefaultMaxHeaderBytes
	}
}

func (r *connHttpHeaderReader) Read(b []byte) (int, error) {
	if !r.isReadingHttpHeader {
		return r.c.Read(b)
	}
	if r.max <= 0 {
		return 0, io.EOF
	}
	if len(b) > r.max {
		b = b[:r.max]
	}
	n, err := r.c.Read(b)
	r.max -= n
	return n, err
}
