package hlfhr

import (
	"net"
)

type listener struct {
	net.Listener

	// Default 4096 Bytes
	hlfhr_readFirstRequestBytesLen int

	// HttpOnHttpsPortErrorHandler handles HTTP requests sent to an HTTPS port.
	hlfhr_httpOnHttpsPortErrorHandler HttpOnHttpsPortErrorHandler
}

func IsMyListener(inner net.Listener) bool {
	_, ok := inner.(*listener)
	return ok
}

func NewListener(inner net.Listener, readFirstRequestBytesLen int, httpOnHttpsPortErrorHandler HttpOnHttpsPortErrorHandler) net.Listener {
	l, ok := inner.(*listener)
	if !ok {
		l = &listener{
			Listener: inner,
		}
	}
	l.hlfhr_readFirstRequestBytesLen = readFirstRequestBytesLen
	if l.hlfhr_readFirstRequestBytesLen == 0 {
		l.hlfhr_readFirstRequestBytesLen = 4096
	}
	l.hlfhr_httpOnHttpsPortErrorHandler = httpOnHttpsPortErrorHandler
	return l
}

func (l *listener) Accept() (c net.Conn, err error) {
	c, err = l.Listener.Accept()
	if err == nil {
		// Hijacking net.Conn
		c = &conn{
			Conn:                              c,
			hlfhr_readFirstRequestBytesLen:    l.hlfhr_readFirstRequestBytesLen,
			hlfhr_httpOnHttpsPortErrorHandler: l.hlfhr_httpOnHttpsPortErrorHandler,
		}
	}
	return
}
