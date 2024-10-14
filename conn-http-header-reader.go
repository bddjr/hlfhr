package hlfhr

import (
	"io"
	"net/http"
)

type connHttpHeaderReader struct {
	isReadingHttpHeader bool
	hasByteBuf          bool
	byteBuf             byte
	c                   *conn
	max                 int
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

func (r *connHttpHeaderReader) peekByte() (byte, bool, error) {
	b := make([]byte, 1)
	n, err := r.c.Conn.Read(b)
	if err != nil || n < 1 {
		return 0, false, err
	}
	r.byteBuf = b[0]
	r.hasByteBuf = true
	return b[0], true, nil
}

func (r *connHttpHeaderReader) Read(b []byte) (int, error) {
	if r.hasByteBuf {
		r.hasByteBuf = false
		b[0] = r.byteBuf
		return 1, nil
	}
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
