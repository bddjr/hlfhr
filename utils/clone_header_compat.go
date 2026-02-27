//go:build !go1.13
// +build !go1.13

package hlfhr_utils

import (
	"net/http"
	_ "unsafe"
)

//go:linkname CloneHeader net/http.Header.clone
func CloneHeader(http.Header) http.Header
