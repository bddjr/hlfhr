package hlfhr

import (
	"crypto/tls"
	"net"
	"reflect"
)

type TLSListener struct {
	net.Listener
	// tlsConfig *tls.Config
	srv *Server
}

func newTLSListener(
	l net.Listener,
	config *tls.Config,
	srv *Server,
) net.Listener {
	return &TLSListener{
		Listener: tls.NewListener(l, config),
		srv:      srv,
	}
}

func (l *TLSListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	tlsConn := c.(*tls.Conn)
	innerConn := (*net.Conn)(reflect.ValueOf(tlsConn).Elem().FieldByName("conn").Addr().UnsafePointer())
	oldInnerConn := *innerConn

	myConn := &conn{
		Conn: oldInnerConn,
		l:    l,
		setReadingTLS: func() {
			*innerConn = oldInnerConn
		},
	}

	*innerConn = myConn

	return c, nil
}
