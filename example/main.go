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
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       5 * time.Second,
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
	fmt.Print("\n  test:\n  ")
	if runtime.GOOS == "windows" {
		fmt.Print("cmd /C ")
	}
	fmt.Print("curl -v -k -L http://localhost", srv.Addr, "/\n\n")
}
