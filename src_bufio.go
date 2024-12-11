package hlfhr

import (
	"bufio"
	"io"
	"unsafe"
)

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
	// copy from bufio.minReadBufferSize
	const minReadBufferSize = 16

	if len(buf) < minReadBufferSize {
		nb := make([]byte, minReadBufferSize)
		copy(nb, buf)
		buf = nb
	}

	return (*bufio.Reader)(unsafe.Pointer(&bufioreader{
		buf:          buf,
		rd:           rd,
		w:            contentLength,
		lastByte:     -1,
		lastRuneSize: -1,
	}))
}

func BufioSetReader(br *bufio.Reader, rd io.Reader) {
	(*bufioreader)(unsafe.Pointer(br)).rd = rd
}
