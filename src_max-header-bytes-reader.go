package hlfhr

import (
	"io"
	"net/http"
)

// 8388607 TiB
const MaxInt = int(^uint(0) >> 1)

type MaxHeaderBytesReader struct {
	Rd  io.Reader
	Max int
}

func (r *MaxHeaderBytesReader) SetMax(srv *http.Server) {
	if srv != nil && srv.MaxHeaderBytes != 0 {
		r.Max = srv.MaxHeaderBytes
	} else {
		r.Max = http.DefaultMaxHeaderBytes
	}
}

func (r *MaxHeaderBytesReader) SetReadingBody() {
	r.Max = MaxInt
}

func (r *MaxHeaderBytesReader) Read(b []byte) (int, error) {
	if r.Max <= 0 {
		return 0, io.EOF
	}
	if len(b) > r.Max {
		b = b[:r.Max]
	}
	n, err := r.Rd.Read(b)
	r.Max -= n
	return n, err
}
