package hlfhr_utils

import (
	"net/http"
	"reflect"
	"sync/atomic"
	"unsafe"
)

// Is [http.Server] shutting down?
func IsShuttingDown(s *http.Server) bool {
	type atomic_bool struct {
		_ struct{}
		v uint32
	}
	inShutdown := reflect.ValueOf(s).Elem().FieldByName("inShutdown")
	p := unsafe.Pointer(inShutdown.UnsafeAddr())
	switch inShutdown.Kind() {
	case reflect.Struct:
		return atomic.LoadUint32(&(*atomic_bool)(p).v) != 0
	case reflect.Int32:
		return atomic.LoadInt32((*int32)(p)) != 0
	}
	panic("hlfhr: IsShuttingDown error: unknown type of http.Server.inShutdown")
}
