package hlfhr

import (
	"io"
	"net/http"
)

type connHttpHeaderReader struct {
	c *conn

	// max: >= 0
	// reading body: -1
	max int
}

func (r *connHttpHeaderReader) setMax() {
	if r.c.l.HttpServer != nil && r.c.l.HttpServer.MaxHeaderBytes > 0 {
		r.max = r.c.l.HttpServer.MaxHeaderBytes
		return
	}
	r.max = http.DefaultMaxHeaderBytes
}

func (r *connHttpHeaderReader) setReadingBody() {
	r.max = -1
}

func (r *connHttpHeaderReader) Read(b []byte) (int, error) {
	if r.max == -1 {
		return r.c.Conn.Read(b)
	}
	if r.max == 0 {
		return 0, io.EOF
	}
	if len(b) > r.max {
		b = b[:r.max]
	}
	n, err := r.c.Conn.Read(b)
	r.max -= n
	return n, err
}
