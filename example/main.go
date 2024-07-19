package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"time"

	"github.com/bddjr/hlfhr"
)

var srv *hlfhr.Server
var rootPath string

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
	http.ServeFile(w, r, filepath.Join(rootPath, "web", r.URL.Path))
}

func testPrint(srv *hlfhr.Server) {
	p := "\n  test:\n  "
	if runtime.GOOS == "windows" {
		p += "cmd /C "
	}
	p = fmt.Sprint(p, "curl -v -k -L http://localhost", srv.Addr, "/\n")
	fmt.Println(p)
}
