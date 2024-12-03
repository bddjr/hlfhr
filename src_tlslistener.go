package hlfhr

import (
	"net"
	"net/http"
)

type Listener struct {
	net.Listener
	srv                         *http.Server
	HttpOnHttpsPortErrorHandler http.Handler
}

func NewListener(
	inner net.Listener,
	srv *http.Server,
	httpOnHttpsPortErrorHandler http.Handler,
) net.Listener {
	return &Listener{
		Listener:                    inner,
		srv:                         srv,
		HttpOnHttpsPortErrorHandler: httpOnHttpsPortErrorHandler,
	}
}

func (l *Listener) Accept() (c net.Conn, err error) {
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
