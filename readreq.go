// HTTPS Listener For HTTP Redirect
//
// Adapted from net/http
//
// BSD-3-clause license
package hlfhr

import (
	"bufio"
	"bytes"
	"net/http"
	"regexp"
)

var compiledRegexp_tlsRecordHeaderLooksLikeHTTP = regexp.MustCompile(`^(GET /|HEAD |POST |PUT /|OPTIO)`)

var compiledRegexp_ReqPath *regexp.Regexp
var compiledRegexp_ReqHost *regexp.Regexp

// Parse the request Host header and path from Hflhr_HttpOnHttpsPortErrorHandler
func ReadReqHostPath(b []byte) (host string, path string, ok bool) {
	if compiledRegexp_ReqPath == nil {
		compiledRegexp_ReqPath = regexp.MustCompile(`/\S*`)
	}
	pb := compiledRegexp_ReqPath.Find(b)
	if pb == nil {
		return
	}
	path = string(pb)

	if compiledRegexp_ReqHost == nil {
		compiledRegexp_ReqHost = regexp.MustCompile(`\r\nHost: \S+\r`)
	}
	hb := compiledRegexp_ReqHost.Find(b)
	if hb == nil {
		return
	}
	host = string(hb[len("\r\nHost: ") : len(hb)-len("\r")])

	ok = true
	return
}

// Parse the request from Hflhr_HttpOnHttpsPortErrorHandler
func ReadReq(b []byte) (*http.Request, error) {
	return http.ReadRequest(bufio.NewReader(bytes.NewBuffer(b)))
}
