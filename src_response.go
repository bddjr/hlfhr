package hlfhr

// v1.2.3 not use [http.Response]

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Using for interface [http.ResponseWriter], [io.StringWriter], [io.ByteWriter].
//
// "Connection" header always set "close".
type Response struct {
	status int // Default: 500
	header http.Header
	body   []byte
}

func NewResponse() *Response {
	return &Response{
		header: http.Header{
			"Date": []string{time.Now().UTC().Format(http.TimeFormat)},
		},
	}
}

func (r *Response) Header() http.Header {
	if r.status == 0 {
		return r.header
	}
	return http.Header{}
}

// Set status code and lock header, if header does not locked.
// If input 0, set 500.
func (r *Response) WriteHeader(statusCode int) {
	if r.status == 0 {
		if statusCode == 0 {
			r.status = 500
		} else {
			r.status = statusCode
		}
	}
}

func (r *Response) Write(b []byte) (int, error) {
	r.WriteHeader(0)
	if len(b) != 0 {
		r.body = append(r.body, b...)
	}
	return len(b), nil
}

func (r *Response) WriteString(s string) (int, error) {
	r.WriteHeader(0)
	if len(s) != 0 {
		r.body = append(r.body, s...)
	}
	return len(s), nil
}

func (r *Response) WriteByte(c byte) error {
	r.WriteHeader(0)
	r.body = append(r.body, c)
	return nil
}

// Flush flushes buffered data to the client.
func (r *Response) Flush(w io.Writer) error {
	// status
	r.WriteHeader(0)
	_, err := fmt.Fprint(w, "HTTP/1.1 ", r.status, " ", http.StatusText(r.status), "\r\n")
	if err != nil {
		return err
	}

	// header
	r.header.Set("Connection", "close")
	r.header.Set("Content-Length", strconv.Itoa(len(r.body)))
	err = r.header.Write(w)
	if err == nil {
		_, err = io.WriteString(w, "\r\n")
	}

	if err != nil || len(r.body) == 0 {
		return err
	}

	// body
	n, err := w.Write(r.body)
	if err == nil && n != len(r.body) {
		return io.ErrShortWrite
	}
	return err
}
