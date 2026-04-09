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

	mc := &Conn{
		Conn:    c,
		TLSConn: nil,
		Server:  l.Server,
	}
	mc.TLSConn = tls.Server(mc, l.TLSConf)
	return mc.TLSConn, nil
}
