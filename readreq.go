package hlfhr

import (
	"bufio"
	"bytes"
	"net/http"
	"regexp"
)

// "GET /index.html HTTP/1.1\r\nHost: localhost:5678\r\nUser-Agent: curl/8.7.1\r\nAccept: */*\r\n\r\n"
// ["GET /index.html HTTP/1.1\r\nHost: localhost:5678\r" "/index.html" "localhost:5678"]
var compiledRegexp_ReadReq = regexp.MustCompile(`^(GET|HEAD|POST|PUT|DELETE|CONNECT|OPTIONS|TRACE|PATCH) (/\S*) HTTP/1\.[01]\r\nHost: (\S+)\r`)

var compiledRegexp_ReadHostnamePort = regexp.MustCompile(`^(\S+?)(:(\d{1,5}))?$`)

// Parse the request Host header and path from Hlfhr_HttpOnHttpsPortErrorHandler.
// Suppose this request using HTTP/1.1
func ReadReqHostPath(b []byte) (host string, path string, ok bool) {
	_, host, path, ok = ReadReqMethodHostPath(b)
	return
}

func ReadReqMethodHostPath(b []byte) (method string, host string, path string, ok bool) {
	fb := compiledRegexp_ReadReq.FindSubmatch(b)
	if fb == nil {
		return
	}
	method = string(fb[1])
	path = string(fb[2])
	host = string(fb[3])
	ok = true
	return
}

// Parse the request from Hlfhr_HttpOnHttpsPortErrorHandler
func ReadReq(b []byte) (*http.Request, error) {
	return http.ReadRequest(bufio.NewReader(bytes.NewBuffer(b)))
}

func ReadHostnamePort(host string) (hostname string, port string) {
	s := compiledRegexp_ReadHostnamePort.FindStringSubmatch(host)
	if s == nil {
		return
	}
	return s[1], s[3]
}
