package hlfhr

import (
	"bufio"
	"io"
	"reflect"
	"unsafe"
)

const brMinBufSize = 16

const brFieldBufIndex = 0
const brFieldWIndex = 3

func NewBufioReaderWithBytes(buf []byte, contentLength int, rd io.Reader) *bufio.Reader {
	br := &bufio.Reader{}
	rv := reflect.ValueOf(br).Elem()

	// fill buffer
	if len(buf) < brMinBufSize {
		nb := make([]byte, brMinBufSize)
		copy(nb, buf)
		buf = nb
	}
	*(*[]byte)(unsafe.Pointer(rv.Field(brFieldBufIndex).UnsafeAddr())) = buf

	// fill reader
	br.Reset(rd)

	// fill length
	if contentLength > 0 {
		*(*int)(unsafe.Pointer(rv.Field(brFieldWIndex).UnsafeAddr())) =
			min(contentLength, len(buf))
	}

	return br
}
