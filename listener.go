package hlfhr

import (
	"net"
	"net/http"
)

type Listener struct {
	net.Listener

	HttpServer                  *http.Server
	HttpOnHttpsPortErrorHandler http.Handler
}

func IsMyListener(inner net.Listener) bool {
	_, ok := inner.(*Listener)
	return ok
}

func NewListener(
	inner net.Listener,
	httpServer *http.Server,
	httpOnHttpsPortErrorHandler http.Handler,
) net.Listener {
	l, ok := inner.(*Listener)
	if !ok {
		l = &Listener{Listener: inner}
	}
	l.HttpServer = httpServer
	l.HttpOnHttpsPortErrorHandler = httpOnHttpsPortErrorHandler
	return l
}

func (l *Listener) Accept() (c net.Conn, err error) {
	c, err = l.Listener.Accept()
	if err == nil {
		// Hijacking net.Conn
		c = NewConn(
			c,
			l.HttpServer,
			l.HttpOnHttpsPortErrorHandler,
		)
	}
	return
}
