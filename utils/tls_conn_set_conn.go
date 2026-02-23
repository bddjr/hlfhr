package hlfhr_utils

import (
	"crypto/tls"
	"net"
	"unsafe"
)

func TLSConnSetConn(tlsConn *tls.Conn, conn net.Conn) {
	(*struct{ conn net.Conn })(unsafe.Pointer(tlsConn)).conn = conn
}
