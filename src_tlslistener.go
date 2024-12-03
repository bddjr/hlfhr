package hlfhr

import (
	"crypto/tls"
	"net"
)

type TLSListener struct {
	net.Listener
	TLSConf *tls.Config
	Server  *Server
}

func (l *TLSListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	mc := new(conn)
	tc := tls.Server(mc, l.TLSConf)

	*mc = conn{
		Conn: c,
		tc:   tc,
		srv:  l.Server,
	}
	return tc, nil
}
