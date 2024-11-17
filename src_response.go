package hlfhr

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

type response struct {
	r http.Response
}

func newResponse() *response {
	return &response{http.Response{
		ProtoMajor: 1,
		ProtoMinor: 1,
		Close:      true,
		Header: http.Header{
			"Date": []string{time.Now().UTC().Format(http.TimeFormat)},
		},
	}}
}

func (r *response) Header() http.Header {
	if r.r.StatusCode != 0 {
		// wrote header
		return http.Header{}
	}
	return r.r.Header
}

func (r *response) WriteHeader(statusCode int) {
	if r.r.StatusCode == 0 {
		// not wrote header
		r.r.StatusCode = statusCode
	}
}

type respBuf struct {
	*bytes.Buffer
	io.Closer
}

func (r *response) Write(b []byte) (int, error) {
	if r.r.Body == nil {
		r.r.Body = &respBuf{Buffer: bytes.NewBuffer(b)}
		return len(b), nil
	}
	return r.r.Body.(*respBuf).Write(b)
}

func (r *response) WriteString(s string) (int, error) {
	if r.r.Body == nil {
		r.r.Body = &respBuf{Buffer: bytes.NewBufferString(s)}
		return len(s), nil
	}
	return r.r.Body.(*respBuf).WriteString(s)
}

// flush flushes buffered data to the client.
func (r *response) flush(w io.WriteCloser) error {
	r.WriteHeader(400)
	if r.r.Body != nil {
		b := r.r.Body.(*respBuf)
		r.r.ContentLength = int64(b.Len())
		b.Closer = w
	}
	return r.r.Write(w)
}
