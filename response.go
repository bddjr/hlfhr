package hlfhr

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"time"
)

func NewResponse() *http.Response {
	return &http.Response{
		ProtoMajor: 1,
		ProtoMinor: 1,
		Proto:      "HTTP/1.1",
		StatusCode: 400,
		Close:      true,
		Header: http.Header{
			"Date": []string{time.Now().UTC().Format(http.TimeFormat)},
		},
	}
}

type ResponseWriter struct {
	Resp    *http.Response
	Conn    net.Conn
	BodyBuf *bytes.Buffer
}

func NewResponseWriter(conn net.Conn, resp *http.Response) *ResponseWriter {
	if resp == nil {
		resp = NewResponse()
	}
	return &ResponseWriter{
		Resp:    resp,
		Conn:    conn,
		BodyBuf: bytes.NewBuffer([]byte{}),
	}
}

func (rw *ResponseWriter) Header() http.Header {
	return rw.Resp.Header
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	return rw.BodyBuf.Write(b)
}

func (rw *ResponseWriter) WriteString(s string) (int, error) {
	return rw.BodyBuf.WriteString(s)
}

func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.Resp.StatusCode = statusCode
}

// Flush flushes buffered data to the client.
func (rw *ResponseWriter) Flush() error {
	rw.Resp.ContentLength = int64(rw.BodyBuf.Len())
	rw.Resp.Body = io.NopCloser(rw.BodyBuf)
	return rw.Resp.Write(rw.Conn)
}

// Redirect tools

func Redirect(w http.ResponseWriter, code int, url string) {
	w.Header().Set("Location", url)
	w.WriteHeader(code)
}

func RedirectToHttps(w http.ResponseWriter, r *http.Request, code int) {
	Redirect(w, code, "https://"+r.Host+r.URL.RequestURI())
}
