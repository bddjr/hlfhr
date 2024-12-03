package hlfhr

import (
	"crypto/tls"
	"net"
)

type TLSListener struct {
	net.Listener
	tlsConfig *tls.Config
	srv       *Server
}

func newTLSListener(
	l net.Listener,
	config *tls.Config,
	srv *Server,
) net.Listener {
	return &TLSListener{
		Listener:  l,
		tlsConfig: config,
		srv:       srv,
	}
}

func (l *TLSListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	myConn := &conn{
		Conn: c,
		l:    l,
	}

	tlsConn := tls.Server(myConn, l.tlsConfig)

	myConn.tlsConn = tlsConn

	return tlsConn, nil
}
