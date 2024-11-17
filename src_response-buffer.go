package hlfhr

import "bytes"

type respBuf struct {
	*bytes.Buffer
}

func (b *respBuf) Close() error {
	return nil
}
