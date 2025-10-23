package hlfhr_utils

import (
	"bufio"
	"io"
	"sync"
	"unsafe"
)

type bufioReaderPoolType struct {
	p sync.Pool
}

func (t *bufioReaderPoolType) get() *bufio.Reader {
	return t.p.Get().(*bufio.Reader)
}

func (t *bufioReaderPoolType) Put(x *bufio.Reader) {
	t.p.Put(x)
}

var BufioReaderPool = bufioReaderPoolType{
	sync.Pool{
		New: func() any {
			return new(bufio.Reader)
		},
	},
}

// copy from bufio.Reader
type bufioreader struct {
	buf          []byte
	rd           io.Reader // reader provided by the client
	r, w         int       // buf read and write positions
	err          error
	lastByte     int // last byte read for UnreadByte; -1 means invalid
	lastRuneSize int // size of last rune read for UnreadRune; -1 means invalid
}

func NewBufioReaderWithBytes(buf []byte, contentLength int, rd io.Reader) *bufio.Reader {
	const defaultBufSize = 4096
	const minReadBufferSize = 16

	if len(buf) == 0 {
		buf = make([]byte, defaultBufSize)
	} else if len(buf) < minReadBufferSize {
		nb := make([]byte, minReadBufferSize)
		copy(nb, buf)
		buf = nb
	}

	br := BufioReaderPool.get()
	*(*bufioreader)(unsafe.Pointer(br)) = bufioreader{
		buf:          buf,
		rd:           rd,
		w:            contentLength,
		lastByte:     -1,
		lastRuneSize: -1,
	}
	return br
}

func BufioSetReader(br *bufio.Reader, rd io.Reader) {
	(*bufioreader)(unsafe.Pointer(br)).rd = rd
}
