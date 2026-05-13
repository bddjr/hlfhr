package hlfhr_utils

import (
	"crypto/tls"
	"net"
	"reflect"
	"unsafe"
)

type tls_conn_conn struct {
	conn net.Conn
}

// Automatic type checking
var _ = func() (_ struct{}) {
	const errmsg = "github.com/bddjr/hlfhr/hlfhr_utils: failed to check type tls_conn_conn"
	a := reflect.TypeOf(tls_conn_conn{})
	if a.NumField() != 1 {
		panic(errmsg)
	}
	if a.Field(0).Type != reflect.TypeOf(tls.Conn{}).Field(0).Type {
		panic(errmsg)
	}
	return
}()

func TLSConnSetConn(tlsConn *tls.Conn, conn net.Conn) {
	(*tls_conn_conn)(unsafe.Pointer(tlsConn)).conn = conn
}
