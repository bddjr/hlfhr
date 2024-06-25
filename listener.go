package hlfhr

import (
	"net"
	"net/http"
)

type listener struct {
	net.Listener

	// Default http.DefaultMaxHeaderBytes
	maxHeaderBytes int

	// HttpOnHttpsPortErrorHandler handles HTTP requests sent to an HTTPS port.
	httpOnHttpsPortErrorHandler HttpOnHttpsPortErrorHandler
}

func IsMyListener(inner net.Listener) bool {
	_, ok := inner.(*listener)
	return ok
}

func NewListener(inner net.Listener, maxHeaderBytes int, httpOnHttpsPortErrorHandler HttpOnHttpsPortErrorHandler) net.Listener {
	l, ok := inner.(*listener)
	if !ok {
		l = &listener{
			Listener: inner,
		}
	}
	l.maxHeaderBytes = maxHeaderBytes
	if l.maxHeaderBytes == 0 {
		l.maxHeaderBytes = http.DefaultMaxHeaderBytes
	}
	l.httpOnHttpsPortErrorHandler = httpOnHttpsPortErrorHandler
	return l
}

func (l *listener) Accept() (c net.Conn, err error) {
	c, err = l.Listener.Accept()
	if err == nil {
		// Hijacking net.Conn
		c = &conn{
			Conn:                        c,
			maxHeaderBytes:              l.maxHeaderBytes,
			httpOnHttpsPortErrorHandler: l.httpOnHttpsPortErrorHandler,
		}
	}
	return
}
