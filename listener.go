package hlfhr

import (
	"net"
	"net/http"
)

type listener struct {
	net.Listener

	httpServer                  *http.Server
	httpOnHttpsPortErrorHandler http.Handler
}

func IsMyListener(inner net.Listener) bool {
	_, ok := inner.(*listener)
	return ok
}

func NewListener(
	inner net.Listener,
	httpServer *http.Server,
	httpOnHttpsPortErrorHandler http.Handler,
) net.Listener {
	l, ok := inner.(*listener)
	if !ok {
		l = &listener{Listener: inner}
	}
	l.httpServer = httpServer
	l.httpOnHttpsPortErrorHandler = httpOnHttpsPortErrorHandler
	return l
}

func (l *listener) Accept() (c net.Conn, err error) {
	c, err = l.Listener.Accept()
	if err == nil {
		// Hijacking net.Conn
		c = NewConn(
			c,
			l.httpServer,
			l.httpOnHttpsPortErrorHandler,
		)
	}
	return
}
