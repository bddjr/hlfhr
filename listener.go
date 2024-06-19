package hlfhr

import (
	"net"
)

type listener struct {
	net.Listener

	// HttpOnHttpsPortErrorHandler handles HTTP requests sent to an HTTPS port.
	hlfhr_httpOnHttpsPortErrorHandler func(b []byte, conn net.Conn)

	// Default 4096 Bytes
	hlfhr_readFirstRequestBytesLen int
}

func IsMyListener(inner net.Listener) bool {
	_, ok := inner.(*listener)
	return ok
}

func NewListener(inner net.Listener, srv *Server) net.Listener {
	l, ok := inner.(*listener)
	if !ok {
		l = &listener{
			Listener: inner,
		}
	}
	l.hlfhr_httpOnHttpsPortErrorHandler = srv.Hlfhr_HttpOnHttpsPortErrorHandler
	l.hlfhr_readFirstRequestBytesLen = srv.Hlfhr_ReadFirstRequestBytesLen
	if l.hlfhr_readFirstRequestBytesLen == 0 {
		l.hlfhr_readFirstRequestBytesLen = 4096
	}
	return l
}

func (l *listener) Accept() (c net.Conn, err error) {
	c, err = l.Listener.Accept()
	if err == nil {
		// Hijacking net.Conn
		c = newConn(c, l)
	}
	return
}
