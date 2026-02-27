//go:build go1.22
// +build go1.22

package hlfhr_utils

import (
	"net/http"
	"reflect"
	"sync/atomic"
	"unsafe"
)

// Can not use go:linkname, see go.dev/issue/67401

// Is [http.Server] shutting down?
func IsShuttingDown(s *http.Server) bool {
	return (*atomic.Bool)(unsafe.Pointer(
		reflect.ValueOf(s).Elem().FieldByName("inShutdown").UnsafeAddr(),
	)).Load()
}
