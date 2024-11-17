package hlfhr

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

type response struct {
	resp http.Response
}

func newResponse() *response {
	return &response{
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

func (r *response) Header() http.Header {
	if r.resp.StatusCode != 0 {
		// wrote header
		return http.Header{}
	}
	return r.resp.Header
}

func (r *response) Write(b []byte) (int, error) {
	if r.resp.Body == nil {
		r.resp.Body = &respBuf{bytes.NewBuffer(b)}
		return len(b), nil
	}
	return r.resp.Body.(*respBuf).Write(b)
}

func (r *response) WriteString(s string) (int, error) {
	if r.resp.Body == nil {
		r.resp.Body = &respBuf{bytes.NewBufferString(s)}
		return len(s), nil
	}
	return r.resp.Body.(*respBuf).WriteString(s)
}

func (r *response) WriteHeader(statusCode int) {
	if r.resp.StatusCode == 0 {
		// not wrote header
		r.resp.StatusCode = statusCode
	}
}

// flush flushes buffered data to the client.
func (r *response) flush(w io.Writer) error {
	r.WriteHeader(400)
	if r.resp.Body != nil {
		r.resp.ContentLength = int64(r.resp.Body.(*respBuf).Len())
	}
	return r.resp.Write(w)
}
