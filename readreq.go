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

// "GET /index.html HTTP/1.1\r\nHost: localhost:5678\r\nUser-Agent: curl/8.7.1\r\nAccept: */*\r\n\r\n"
//
// ["GET /index.html HTTP/1.1\r\nHost: localhost:5678\r" "/index.html" "localhost:5678"]
var compiledRegexp_ReadReq = regexp.MustCompile(`^[A-Z]{3,7} (/\S*) HTTP/1.1\r\nHost: (\S+)\r`)

// Parse the request Host header and path from Hflhr_HttpOnHttpsPortErrorHandler.
// Suppose this request using HTTP/1.1
func ReadReqHostPath(b []byte) (host string, path string, ok bool) {
	fb := compiledRegexp_ReadReq.FindSubmatch(b)
	if fb == nil {
		return
	}
	path = string(fb[1])
	host = string(fb[2])
	ok = true
	return
}

// Parse the request from Hflhr_HttpOnHttpsPortErrorHandler
func ReadReq(b []byte) (*http.Request, error) {
	return http.ReadRequest(bufio.NewReader(bytes.NewBuffer(b)))
}
