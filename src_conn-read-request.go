package hlfhr

import (
	"errors"
	"io"
	"net/http"
)

func (c *conn) readRequest(b []byte, n int) (*http.Request, error) {
	rd := &connReadRequestReader{rd: c.Conn}

	// set max
	if c.l.HttpServer != nil && c.l.HttpServer.MaxHeaderBytes != 0 {
		rd.max = c.l.HttpServer.MaxHeaderBytes
	} else {
		rd.max = http.DefaultMaxHeaderBytes
	}
	rd.max -= n
	if rd.max < 0 {
		rd.max = 0
	}

	// bufio
	br := NewBufioReaderWithBytes(b, n, rd)

	// read
	req, err := http.ReadRequest(br)
	if err != nil {
		return nil, err
	}
	if req.Host == "" {
		return nil, errors.New("missing required Host header")
	}

	rd.max = -1
	return req, nil
}

type connReadRequestReader struct {
	rd io.Reader

	// max: >= 0
	// reading body: -1
	max int
}

func (r *connReadRequestReader) Read(b []byte) (int, error) {
	if r.max == -1 {
		return r.rd.Read(b)
	}
	if r.max == 0 {
		return 0, io.EOF
	}
	if len(b) > r.max {
		b = b[:r.max]
	}
	n, err := r.rd.Read(b)
	r.max -= n
	return n, err
}
