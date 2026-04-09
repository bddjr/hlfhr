package hlfhr_lib

// v1.2.3 not use [http.Response]

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

// Using for interface [http.ResponseWriter], [io.StringWriter] and [io.ByteWriter].
type Response struct {
	conn         net.Conn
	status       int
	header       http.Header
	lockedHeader http.Header
	body         []byte
	flushErr     error
	close        bool
	flushed      bool
}

func NewResponse(c net.Conn, status int, closeConnection bool) *Response {
	return &Response{
		conn:   c,
		status: status,
		header: http.Header{
			"Date": []string{time.Now().UTC().Format(http.TimeFormat)},
		},
		lockedHeader: nil,
		body:         []byte{},
		flushErr:     nil,
		close:        closeConnection,
		flushed:      false,
	}
}

func (r *Response) Header() http.Header {
	return r.header
}

// Set status code and lock header, if header does not locked.
func (r *Response) WriteHeader(statusCode int) {
	if r.lockedHeader == nil {
		r.status = statusCode
		r.lockedHeader = r.header.Clone()
	}
}

func (r *Response) lockHeader() {
	if r.lockedHeader == nil {
		r.lockedHeader = r.header.Clone()
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

func (r *Response) SetDeadline(t time.Time) error {
	return r.conn.SetDeadline(t)
}

func (r *Response) SetReadDeadline(t time.Time) error {
	return r.conn.SetReadDeadline(t)
}

func (r *Response) SetWriteDeadline(t time.Time) error {
	return r.conn.SetWriteDeadline(t)
}

// Flush flushes buffered data to the client.
func (r *Response) Flush() {
	r.FlushError()
}

func (r *Response) FlushError() error {
	if r.flushed {
		return r.flushErr
	}
	r.flushed = true
	r.lockHeader()

	// status
	_, r.flushErr = fmt.Fprint(r.conn, "HTTP/1.1 ", r.status, " ", http.StatusText(r.status), "\r\n")
	if r.flushErr != nil {
		return r.flushErr
	}

	// header
	if r.close {
		r.lockedHeader["Connection"] = []string{"close"}
	}
	r.lockedHeader["Content-Length"] = []string{strconv.Itoa(len(r.body))}

	r.flushErr = r.lockedHeader.Write(r.conn)
	if r.flushErr != nil {
		return r.flushErr
	}
	_, r.flushErr = io.WriteString(r.conn, "\r\n")
	if r.flushErr != nil {
		return r.flushErr
	}

	// body
	if len(r.body) == 0 {
		return nil
	}
	var n int
	n, r.flushErr = r.conn.Write(r.body)
	if r.flushErr == nil && n != len(r.body) {
		r.flushErr = io.ErrShortWrite
	}
	return r.flushErr
}
