package hlfhr_utils

import (
	"io"
	"sync"
)

type limitedReaderPoolType struct {
	p sync.Pool
}

func (t *limitedReaderPoolType) get() *io.LimitedReader {
	return t.p.Get().(*io.LimitedReader)
}

func (t *limitedReaderPoolType) Put(x *io.LimitedReader) {
	t.p.Put(x)
}

var LimitedReaderPool = limitedReaderPoolType{
	sync.Pool{
		New: func() any {
			return new(io.LimitedReader)
		},
	},
}

func NewLimitedReader(R io.Reader, N int64) *io.LimitedReader {
	r := LimitedReaderPool.get()
	*r = io.LimitedReader{
		R: R,
		N: N,
	}
	return r
}
