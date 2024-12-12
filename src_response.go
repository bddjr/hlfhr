package hlfhr

// v1.2.3 not use [http.Response]

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

// Using for interface [http.ResponseWriter], [io.StringWriter], [io.ByteWriter].
//
// "Connection" header always set "close".
type Response struct {
	conn         net.Conn
	status       int // Default: 400
	header       http.Header
	lockedHeader http.Header
	body         []byte
	close        bool
}

func NewResponse(c net.Conn, ConnectionHeaderSetClose bool) *Response {
	return &Response{
		conn:   c,
		status: 400,
		header: http.Header{
			"Date": []string{time.Now().UTC().Format(http.TimeFormat)},
		},
		close: ConnectionHeaderSetClose,
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
	r.lockHeader()

	// status
	_, err := fmt.Fprint(r.conn, "HTTP/1.1 ", r.status, " ", http.StatusText(r.status), "\r\n")
	if err != nil {
		return err
	}

	// header
	if r.close {
		r.header["Connection"] = []string{"close"}
	}
	r.header["Content-Length"] = []string{strconv.Itoa(len(r.body))}

	err = r.header.Write(r.conn)
	if err == nil {
		_, err = io.WriteString(r.conn, "\r\n")
	}
	if err != nil || len(r.body) == 0 {
		return err
	}

	// body
	n, err := r.conn.Write(r.body)
	if err == nil && n != len(r.body) {
		return io.ErrShortWrite
	}
	return err
}
