package hlfhr

import (
	"io"
	"unsafe"
)

type BytesAndStringWriter interface {
	io.Writer
	io.StringWriter
}

type FastStringWriter struct {
	io.Writer
}

func NewFastStringWriter(w io.Writer) BytesAndStringWriter {
	if sw, ok := w.(BytesAndStringWriter); ok {
		return sw
	}
	return &FastStringWriter{w}
}

func (sw *FastStringWriter) WriteString(s string) (int, error) {
	// no copy
	return sw.Write(*(*[]byte)(unsafe.Pointer(&s)))
}
