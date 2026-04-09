package hlfhr_utils

import (
	"bufio"
	"io"
	"reflect"
	"unsafe"
)

// Copy from [bufio.Reader]
type bufio_reader struct {
	buf          []byte
	rd           io.Reader // reader provided by the client
	r, w         int       // buf read and write positions
	err          error
	lastByte     int // last byte read for UnreadByte; -1 means invalid
	lastRuneSize int // size of last rune read for UnreadRune; -1 means invalid
}

// Automatic type checking
var _ = func() (_ struct{}) {
	const errmsg = "hlfhr_utils: failed to check type bufio_reader"
	a := reflect.TypeOf(bufio_reader{})
	b := reflect.TypeOf(bufio.Reader{})
	if a.Kind() != b.Kind() {
		panic(errmsg)
	}
	anf := a.NumField()
	for i := 0; i < anf; i++ {
		af := a.Field(i)
		bf := b.Field(i)
		if af.Offset != bf.Offset {
			panic(errmsg)
		}
		aft := af.Type
		aftk := aft.Kind()
		bft := bf.Type
		if aftk != bft.Kind() {
			panic(errmsg)
		}
		if aftk == reflect.Pointer {
			aft = af.Type.Elem()
			aftk = aft.Kind()
			bft = bf.Type.Elem()
			if aftk != bft.Kind() {
				panic(errmsg)
			}
		}
		if aft.Size() != bft.Size() {
			panic(errmsg)
		}
		if aft.PkgPath() != bft.PkgPath() {
			panic(errmsg)
		}
		if aft.Name() != bft.Name() {
			panic(errmsg)
		}
	}
	return
}()

func NewBufioReaderWithBytes(buf []byte, contentLength int, rd io.Reader) *bufio.Reader {
	if len(buf) == 0 {
		return bufio.NewReader(rd)
	}
	const minReadBufferSize = 16

	if len(buf) < minReadBufferSize {
		nb := make([]byte, minReadBufferSize)
		copy(nb, buf)
		buf = nb
	}

	br := new(bufio.Reader)
	*(*bufio_reader)(unsafe.Pointer(br)) = bufio_reader{
		buf:          buf,
		rd:           rd,
		w:            contentLength,
		lastByte:     -1,
		lastRuneSize: -1,
	}
	return br
}

func BufioSetReader(br *bufio.Reader, rd io.Reader) {
	(*bufio_reader)(unsafe.Pointer(br)).rd = rd
}
