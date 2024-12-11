package hlfhr

// v1.2.3 not use [http.Response]

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

// Using for interface
// [http.ResponseWriter],
// [http.Flusher],
// [http.Hijacker],
// [io.StringWriter],
// [io.ByteWriter].
//
// Using for struct [http.ResponseController].
type Response struct {
	conn   net.Conn
	status int // Default: 500
	header http.Header
	body   []byte
	br     *bufio.Reader
	brw    *bufio.ReadWriter

	headerLocked         bool
	flushed              bool
	hijacked             bool
	disableContentLength bool
}

func NewResponse(conn net.Conn, br *bufio.Reader) *Response {
	r := new(Response)
	r.Reset(conn, br)
	return r
}

func (r *Response) Reset(conn net.Conn, br *bufio.Reader) {
	*r = Response{
		status: 500,
		header: http.Header{
			"Date": []string{time.Now().UTC().Format(http.TimeFormat)},
		},
		conn: conn,
		br:   br,
	}
}

func (r *Response) Header() http.Header {
	if r.headerLocked {
		return r.header.Clone()
	}
	return r.header
}

func (r *Response) lockHeader() {
	if !r.headerLocked {
		r.header = r.header.Clone()
		r.headerLocked = true
	}
}

// Set status code and lock header, if header does not locked.
func (r *Response) WriteHeader(statusCode int) {
	if !r.headerLocked {
		r.status = statusCode
		r.lockHeader()
	}
}

func (r *Response) Write(b []byte) (int, error) {
	r.lockHeader()
	if len(b) != 0 {
		r.body = append(r.body, b...)
	}
	return len(b), nil
}

func (r *Response) WriteString(s string) (int, error) {
	r.lockHeader()
	if len(s) != 0 {
		r.body = append(r.body, s...)
	}
	return len(s), nil
}

func (r *Response) WriteByte(c byte) error {
	r.lockHeader()
	r.body = append(r.body, c)
	return nil
}

func (r *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r.hijacked = true
	if r.br == nil {
		r.br = bufio.NewReader(r.conn)
	}
	if r.brw == nil {
		r.brw = bufio.NewReadWriter(r.br, bufio.NewWriter(r.conn))
	}
	return r.conn, r.brw, nil
}

func (r *Response) SetReadDeadline(t time.Time) error {
	return r.conn.SetReadDeadline(t)
}

func (r *Response) SetWriteDeadline(t time.Time) error {
	return r.conn.SetWriteDeadline(t)
}

func (r *Response) ReadFrom(src io.Reader) (int64, error) {
	r.disableContentLength = true
	err := r.FlushError()
	if err != nil {
		return 0, err
	}
	return io.Copy(r.conn, src)
}

// func (r *Response) EnableFullDuplex() error {}
// func (r *Response) CloseNotify() <-chan bool {}

func (r *Response) KeepAlive() bool {
	if r.hijacked || r.disableContentLength {
		return false
	}
	if c, ok := r.header["Connection"]; ok && len(c) != 0 && c[0] == "close" {
		return false
	}
	return true
}

// Flush flushes buffered data to the client.
func (r *Response) Flush() {
	r.FlushError()
}

func (r *Response) FlushError() error {
	if r.hijacked {
		r.flushed = true
		return http.ErrHijacked
	}
	if r.flushed {
		return nil
	}
	r.flushed = true
	sw := NewFastStringWriter(r.conn)

	// status
	r.lockHeader()
	_, err := sw.WriteString("HTTP/1.1 ")
	if err == nil {
		_, err = sw.WriteString(strconv.Itoa(r.status) + " " + http.StatusText(r.status) + "\r\n")
	}
	if err != nil {
		return err
	}

	// header
	if r.disableContentLength {
		delete(r.header, "Content-Length")
		r.header["Connection"] = []string{"close"}
	} else {
		r.header["Content-Length"] = []string{strconv.Itoa(len(r.body))}
	}

	err = r.header.Write(sw)
	if err == nil {
		_, err = sw.WriteString("\r\n")
	}
	if err != nil || len(r.body) == 0 {
		return err
	}

	// body
	n, err := sw.Write(r.body)
	if err == nil && n != len(r.body) {
		return io.ErrShortWrite
	}
	return err
}
