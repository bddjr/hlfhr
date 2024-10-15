package hlfhr

import (
	"io"
	"net/http"
)

type connHttpHeaderReader struct {
	isReadingBody bool
	firstByte     byte
	c             *conn
	max           int
}

func (r *connHttpHeaderReader) resetMaxHeaderBytes() {
	srv := r.c.l.HttpServer
	if srv != nil {
		mhb := srv.MaxHeaderBytes
		if mhb != 0 {
			r.max = mhb
			return
		}
	}
	r.max = http.DefaultMaxHeaderBytes
}

func (r *connHttpHeaderReader) Read(b []byte) (int, error) {
	if r.isReadingBody || len(b) == 0 {
		return r.c.Conn.Read(b)
	}

	if r.max <= 0 {
		return 0, io.EOF
	}

	offset := 0

	if r.firstByte != 0 {
		b[0] = r.firstByte
		r.firstByte = 0
		r.max--
		if len(b) == 1 || r.max <= 0 {
			return 1, nil
		}
		b = b[1:]
		offset++
	}

	if len(b) > r.max {
		b = b[:r.max]
	}

	n, err := r.c.Conn.Read(b)
	r.max -= n
	return n + offset, err
}
