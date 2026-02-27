//go:build go1.8 && !go1.22
// +build go1.8,!go1.22

package hlfhr_utils

import (
	"net/http"
	_ "unsafe"
)

// Is [http.Server] shutting down?
//
//go:linkname IsShuttingDown net/http.(*Server).shuttingDown
func IsShuttingDown(*http.Server) bool
