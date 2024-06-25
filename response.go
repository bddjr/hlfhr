package hlfhr

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

func NewResponse() *http.Response {
	h := make(http.Header)
	h.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	h.Set("X-Powered-By", "github.com/bddjr/hlfhr")
	return &http.Response{
		ProtoMajor: 1,
		ProtoMinor: 1,
		StatusCode: 400,
		Header:     h,
		Close:      true,
	}
}

type ResponseWriter struct {
	resp    *http.Response
	writer  io.Writer
	bodyBuf *bytes.Buffer
}

func NewResponseWriter(w io.Writer, resp *http.Response) *ResponseWriter {
	if resp == nil {
		resp = NewResponse()
	}
	return &ResponseWriter{
		resp:    resp,
		writer:  w,
		bodyBuf: bytes.NewBuffer([]byte{}),
	}
}

func (rw *ResponseWriter) Header() http.Header {
	return rw.resp.Header
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	return rw.bodyBuf.Write(b)
}

func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.resp.StatusCode = statusCode
}

func (rw *ResponseWriter) WriteLock() error {
	rw.resp.ContentLength = int64(rw.bodyBuf.Len())
	rw.resp.Body = io.NopCloser(rw.bodyBuf)
	return rw.resp.Write(rw.writer)
}
