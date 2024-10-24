package hlfhr

import (
	"bufio"
	"io"
	"reflect"
)

func NewBufioReaderWithBytes(buf []byte, contentLength int, rd io.Reader) *bufio.Reader {
	br := &bufio.Reader{}
	brElem := reflect.ValueOf(br).Elem()

	// fill buffer
	if len(buf) < 16 {
		nb := make([]byte, 16)
		copy(nb, buf)
		buf = nb
	}
	*(*[]byte)(brElem.FieldByName("buf").Addr().UnsafePointer()) =
		buf

	// fill reader
	br.Reset(rd)

	// fill length
	if contentLength > 0 {
		*(*int)(brElem.FieldByName("w").Addr().UnsafePointer()) =
			min(contentLength, len(buf))
	}

	return br
}
