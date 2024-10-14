package main

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/bddjr/hlfhr"
)

var srv *hlfhr.Server

func main() {
	// Use hlfhr.New
	srv = hlfhr.New(&http.Server{
		Addr:              ":5678",
		Handler:           http.HandlerFunc(httpResponseHandle),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       10 * time.Second,
	})
	// Then just use it like http.Server .

	testPrint(srv)

	err := srv.ListenAndServeTLS("localhost.crt", "localhost.key")
	fmt.Println(err)
}

func httpResponseHandle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		io.WriteString(w, "Hello hlfhr!\n\n")
	} else {
		http.NotFound(w, r)
	}
}

func testPrint(srv *hlfhr.Server) {
	p := "\n  test:\n  "
	if runtime.GOOS == "windows" {
		p += "cmd /C "
	}
	p = fmt.Sprint(p, "curl -v -k -L http://localhost", srv.Addr, "/\n")
	fmt.Println(p)
}
