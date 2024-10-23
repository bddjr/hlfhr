package hlfhr

import (
	"io"
	"net/http"
)

const chhrStatusReadingBody = -0x100

type connHttpHeaderReader struct {
	c *conn

	// max: >= 0
	// first byte: -0x01 ~ -0xff
	// reading body: -0x100
	status int
}

func (r *connHttpHeaderReader) setStatusMax() {
	srv := r.c.l.HttpServer
	if srv != nil {
		r.status = srv.MaxHeaderBytes
		if r.status > 0 {
			return
		}
	}
	r.status = http.DefaultMaxHeaderBytes
}

func (r *connHttpHeaderReader) setStatusFirstByte(b byte) {
	r.status = -int(b)
}

func (r *connHttpHeaderReader) setStatusReadingBody() {
	r.status = chhrStatusReadingBody
}

func (r *connHttpHeaderReader) Read(b []byte) (int, error) {
	if r.status < 0 {
		if r.status == chhrStatusReadingBody {
			// reading body
			return r.c.Conn.Read(b)
		}
		// first byte
		b[0] = byte(-r.status)
		r.setStatusMax()
		r.status--
		return 1, nil
	}

	// max
	if r.status == 0 {
		return 0, io.EOF
	}

	if len(b) > r.status {
		b = b[:r.status]
	}
	n, err := r.c.Conn.Read(b)
	r.status -= n
	return n, err
}
