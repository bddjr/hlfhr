package hlfhr

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"time"
)

func NewResponse() *http.Response {
	h := make(http.Header)
	h.Set("Date", time.Now().UTC().Format(http.TimeFormat))
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
	Conn    net.Conn
	BodyBuf *bytes.Buffer

	HijackRBuf *bufio.Reader
	HijackRW   *bufio.ReadWriter
	Hijacked   bool
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

func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.Resp.StatusCode = statusCode
}

// http.ResponseController

// Flush flushes buffered data to the client.
func (rw *ResponseWriter) Flush() error {
	rw.Resp.ContentLength = int64(rw.BodyBuf.Len())
	rw.Resp.Body = io.NopCloser(rw.BodyBuf)
	return rw.Resp.Write(rw.Conn)
}

func (rw *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	rw.Hijacked = true
	if rw.HijackRW == nil {
		if rw.HijackRBuf == nil {
			rw.HijackRBuf = bufio.NewReader(rw.Conn)
		}
		rw.HijackRW = bufio.NewReadWriter(
			rw.HijackRBuf,
			bufio.NewWriter(rw.Conn),
		)
	}
	return rw.Conn, rw.HijackRW, nil
}

func (rw *ResponseWriter) SetReadDeadline(t time.Time) error {
	return rw.Conn.SetReadDeadline(t)
}

func (rw *ResponseWriter) SetWriteDeadline(t time.Time) error {
	return rw.Conn.SetWriteDeadline(t)
}

func (rw *ResponseWriter) EnableFullDuplex() error {
	return nil
}

// Redirect tools

func Redirect(w http.ResponseWriter, code int, url string) {
	w.Header().Set("Location", url)
	w.WriteHeader(code)
}

func RedirectToHttps(w http.ResponseWriter, r *http.Request, code int) {
	Redirect(w, code, "https://"+r.Host+r.URL.RequestURI())
}
