package hlfhr

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

type response struct {
	resp http.Response
	w    io.Writer
	body *bytes.Buffer
}

func newResponse(w io.Writer) *response {
	return &response{
		w: w,
		resp: http.Response{
			ProtoMajor: 1,
			ProtoMinor: 1,
			Close:      true,
			Header: http.Header{
				"Date": []string{time.Now().UTC().Format(http.TimeFormat)},
			},
		},
	}
}

func (rw *response) Header() http.Header {
	if rw.resp.StatusCode != 0 {
		// wrote header
		return http.Header{}
	}
	return rw.resp.Header
}

func (rw *response) Write(b []byte) (int, error) {
	if rw.body == nil {
		nb := make([]byte, len(b))
		copy(nb, b)
		rw.body = bytes.NewBuffer(nb)
		return len(b), nil
	}
	return rw.body.Write(b)
}

func (rw *response) WriteString(s string) (int, error) {
	if rw.body == nil {
		rw.body = bytes.NewBuffer([]byte(s))
		return len(s), nil
	}
	return rw.body.WriteString(s)
}

func (rw *response) WriteHeader(statusCode int) {
	if rw.resp.StatusCode == 0 {
		// not wrote header
		rw.resp.StatusCode = statusCode
	}
}

// flush flushes buffered data to the client.
func (rw *response) flush() error {
	rw.WriteHeader(400)
	if rw.body != nil {
		rw.resp.ContentLength = int64(rw.body.Len())
		rw.resp.Body = io.NopCloser(rw.body)
	}
	return rw.resp.Write(rw.w)
}
