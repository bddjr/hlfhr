//go:build go1.8 && !go1.19
// +build go1.8,!go1.19

package hlfhr_utils

import (
	"net/http"
	"reflect"
	"sync/atomic"
	"unsafe"
)

// Is [http.Server] shutting down?
func IsShuttingDown(s *http.Server) bool {
	return atomic.LoadInt32((*int32)(unsafe.Pointer(
		reflect.ValueOf(s).Elem().FieldByName("inShutdown").UnsafeAddr(),
	))) != 0
}
