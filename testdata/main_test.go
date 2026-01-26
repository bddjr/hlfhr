package main_test

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/bddjr/hlfhr"
	hlfhr_utils "github.com/bddjr/hlfhr/utils"
	"golang.org/x/net/http2"
)

func request(serverAddr string) {
	println("request")
	// requestBody := make([]byte, 8192)
	// mathrand.Read(requestBody)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	err := http2.ConfigureTransport(transport)
	if err != nil {
		panic(err)
	}

	client := http.Client{
		Transport: transport,
		Timeout:   time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			println("Redirect")
			return nil
		},
	}
	defer client.CloseIdleConnections()

	// Intentionally occupy a connection to test whether
	// the server can handle requests in parallel.
	c, err := net.Dial("tcp", serverAddr)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	httpURL := "http://" + serverAddr + "/test?a=b&c=d"
	println(httpURL)
	println()

	// HTTP/1.1
	for _, method := range []string{
		"GET",
		"HEAD",
		"POST",
		"PUT",
		"DELETE",
		"CONNECT",
		"OPTIONS",
		"TRACE",
		"PATCH",
	} {
		println(method)

		// requestBodyReader := bytes.NewReader(requestBody)
		req, err := http.NewRequest(method, httpURL, nil)
		if err != nil {
			panic(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		if resp.ProtoMajor != 2 {
			panic("Response does not using h2 protocol!")
		}
		if method != "HEAD" {
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			print(string(respBody))
		}

		println()
	}

	// Missing "Host" header
	{
		println("Test missing \"Host\" header")
		_, err = io.WriteString(c, "GET / HTTP/1.0\r\n\r\n")
		if err != nil {
			panic(err)
		}
		resp, err := http.ReadResponse(bufio.NewReader(c), nil)
		if err != nil {
			panic(err)
		}
		if resp.StatusCode != 400 {
			panic(resp.StatusCode)
		}
		println()
	}
}

func requestTestHlfhrHandler(serverAddr string) {
	println("requestTestHlfhrHandler")
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	err := http2.ConfigureTransport(transport)
	if err != nil {
		panic(err)
	}

	client := http.Client{
		Transport: transport,
		Timeout:   time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			println("Redirect")
			return nil
		},
	}
	defer client.CloseIdleConnections()

	httpURL := "http://" + serverAddr + "/test?a=b&c=d"
	println(httpURL)
	println()
	// requestBodyReader := bytes.NewReader(requestBody)
	req, err := http.NewRequest("GET", httpURL, nil)
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	print(string(respBody))

	println()
}

func test1(serverAddr string) {
	println()

	srv := hlfhr.New(&http.Server{
		Addr: serverAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "HEAD" {
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			enc := json.NewEncoder(w)
			enc.SetEscapeHTML(false)
			tlsVersion, _ := strings.CutPrefix(tls.VersionName(r.TLS.Version), "TLS ")
			err := enc.Encode(map[string]any{
				"Method":         r.Method,
				"Proto":          r.Proto,
				"TLS_Version":    tlsVersion,
				"TLS_ServerName": r.TLS.ServerName,
				"Host":           r.Host,
				"URI":            r.RequestURI,
				// "RequestHeader":  r.Header,
			})
			if err != nil {
				panic(err)
			}
		}),
	})
	srv.Listen80RedirectTo443 = true

	println("Listen " + serverAddr)
	if hlfhr_utils.IsShuttingDown(srv.Server) {
		panic(true)
	}

	var err error
	go func() {
		err = srv.ListenAndServeTLS("invalid.crt", "invalid.key")
	}()
	time.Sleep(100 * time.Millisecond)
	if err != nil {
		panic(err)
	}
	println()

	request(serverAddr)
	if addr, b := strings.CutSuffix(serverAddr, ":443"); b {
		addr += ":80"
		request(addr)
	}

	srv.HlfhrHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		err := enc.Encode(map[string]any{
			"Method": r.Method,
			"Proto":  r.Proto,
			"IsTLS":  r.TLS != nil,
			"Host":   r.Host,
			"URI":    r.RequestURI,
			// "RequestHeader":  r.Header,
		})
		if err != nil {
			panic(err)
		}
	})
	requestTestHlfhrHandler(serverAddr)

	println("Shutdown")
	err = srv.Shutdown(context.Background())
	if err != nil {
		panic(err)
	}
	if !hlfhr_utils.IsShuttingDown(srv.Server) {
		panic(false)
	}
	println()
}

func Test(t *testing.T) {
	test1("127.0.0.1:45876")
	test1("[::1]:45876")
	test1("127.0.0.1:80")
	test1("[::1]:80")
	test1("127.0.0.1:443")
	test1("[::1]:443")

	println("OK\n")
}
