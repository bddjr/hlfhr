package hlfhr

import (
	"io"
	"net/http"
)

type connHttpHeaderReader struct {
	isReadingHttpHeader bool
	c                   *conn
	max                 int
}

func (r *connHttpHeaderReader) resetMaxHeaderBytes(Min int) {
	if srv := r.c.l.HttpServer; srv != nil {
		mhb := srv.MaxHeaderBytes
		if mhb != 0 {
			r.max = max(mhb, Min)
			return
		}
	}
	r.max = http.DefaultMaxHeaderBytes
}

func (r *connHttpHeaderReader) Read(b []byte) (int, error) {
	if !r.isReadingHttpHeader {
		return r.c.Conn.Read(b)
	}
	if r.max <= 0 {
		return 0, io.EOF
	}
	if len(b) > r.max {
		b = b[:r.max]
	}
	n, err := r.c.Conn.Read(b)
	r.max -= n
	return n, err
}
