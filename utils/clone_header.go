//go:build go1.13
// +build go1.13

package hlfhr_utils

import "net/http"

func CloneHeader(h http.Header) http.Header {
	return h.Clone()
}
