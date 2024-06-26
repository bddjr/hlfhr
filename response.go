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
	h.Set("Connection", "close")
	return &http.Response{
		ProtoMajor: 1,
		ProtoMinor: 1,
		Proto:      "HTTP/1.1",
		StatusCode: 400,
		Header:     h,
	}
}

type ResponseWriter struct {
	Resp    *http.Response
	Writer  io.Writer
	BodyBuf *bytes.Buffer
}

func NewResponseWriter(w io.Writer, resp *http.Response) *ResponseWriter {
	if resp == nil {
		resp = NewResponse()
	}
	return &ResponseWriter{
		Resp:    resp,
		Writer:  w,
		BodyBuf: bytes.NewBuffer([]byte{}),
	}
}

func (rw *ResponseWriter) Header() http.Header {
	return rw.Resp.Header
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	return rw.BodyBuf.Write(b)
}

func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.Resp.StatusCode = statusCode
}

func (rw *ResponseWriter) Finish() error {
	rw.Resp.ContentLength = int64(rw.BodyBuf.Len())
	rw.Resp.Body = io.NopCloser(rw.BodyBuf)
	return rw.Resp.Write(rw.Writer)
}

func Redirect(w http.ResponseWriter, code int, url string) {
	w.Header().Set("Location", url)
	w.WriteHeader(code)
}

func RedirectToHttps(w http.ResponseWriter, r *http.Request, code int) {
	Redirect(w, code, "https://"+r.Host+r.URL.RequestURI())
}
