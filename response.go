package hlfhr

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Response struct {
	*http.Response

	writer io.Writer
}

func NewResponse(w io.Writer) *Response {
	resp := &Response{
		Response: &http.Response{
			ProtoMajor: 1,
			ProtoMinor: 1,
			Proto:      "HTTP/1.1",
			Header:     make(http.Header),
			Close:      true,
			StatusCode: 400,
		},
		writer: w,
	}
	resp.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	resp.Header.Set("X-Redirect-By", "hlfhr")
	return resp
}

// Example:
//
//	resp.SetContentType("text/html")
func (resp Response) SetContentType(value string) {
	resp.Header.Set("Content-Type", value)
}

// Example:
//
//	resp.Write(
//	  "Hello world!\n",
//	  "Hello hlfhr!\n",
//	)
func (resp Response) Write(a ...any) error {
	if len(a) > 0 {
		b := fmt.Append([]byte{}, a...)
		resp.Body = io.NopCloser(bytes.NewBuffer(b))
		resp.ContentLength = int64(len(b))
	} else {
		resp.Body = nil
		resp.ContentLength = 0
	}
	return resp.Response.Write(resp.writer)
}

// Example:
//
//	resp.Redirect(302, "https://example.com")
func (resp Response) Redirect(StatusCode int, Location string) error {
	resp.StatusCode = StatusCode
	resp.Header.Set("Location", Location)
	return resp.Write()
}

func (resp Response) ScriptRedirect() error {
	resp.StatusCode = 400
	resp.SetContentType("text/html")
	return resp.Write(
		"<noscript> ", ErrHttpOnHttpsPort, " </noscript>\n",
		"<script> location.protocol = 'https:' </script>\n",
	)
}
