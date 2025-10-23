package hlfhr_utils

import (
	"net/http"
	"reflect"
	"sync/atomic"
	"unsafe"
)

// Is [http.Server] shutting down?
func IsShuttingDown(s *http.Server) bool {
	return (*atomic.Bool)(unsafe.Pointer(
		reflect.ValueOf(s).Elem().FieldByName("inShutdown").UnsafeAddr(),
	)).Load()
}
