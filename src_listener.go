package hlfhr

import (
	"net"
	"net/http"
)

type listener struct {
	net.Listener
	srv     *http.Server
	handler http.Handler
}

func NewListener(
	inner net.Listener,
	srv *http.Server,
	httpOnHttpsPortErrorHandler http.Handler,
) net.Listener {
	if _, ok := inner.(isMyListener); ok {
		return inner
	}
	return &listener{
		Listener: inner,
		srv:      srv,
		handler:  httpOnHttpsPortErrorHandler,
	}
}

func (l *listener) Accept() (c net.Conn, err error) {
	c, err = l.Listener.Accept()
	if err == nil {
		// Hijacking net.Conn
		c = &conn{
			Conn: c,
			l:    l,
		}
	}
	return
}

type isMyListener interface {
	IsHttpsListenerForHttpRedirect()
}

func (l *listener) IsHttpsListenerForHttpRedirect() {}
