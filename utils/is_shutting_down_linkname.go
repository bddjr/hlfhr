//go:build !go1.23
// +build !go1.23

package hlfhr_utils

import (
	"net/http"
	_ "unsafe"
)

// Is [http.Server] shutting down?
//
//go:linkname IsShuttingDown net/http.(*Server).shuttingDown
func IsShuttingDown(*http.Server) bool
