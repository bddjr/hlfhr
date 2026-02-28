package hlfhr_utils

import (
	"net/http"
	"reflect"
	"sync/atomic"
	"unsafe"
)

var offset_inShutdown = func() uintptr {
	sf, ok := reflect.TypeOf(http.Server{}).FieldByName("inShutdown")
	if !ok {
		panic("hlfhr_utils: cannot get http.Server.inShutdown offset")
	}
	return sf.Offset
}()

// Is [http.Server] shutting down?
func IsShuttingDown(s *http.Server) bool {
	return (*atomic.Bool)(unsafe.Add(unsafe.Pointer(s), offset_inShutdown)).Load()
}
