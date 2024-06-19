package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/bddjr/hlfhr"
)

var srv *hlfhr.Server
var rootPath string

func main() {
	getRootPath()

	srv = hlfhr.New(&http.Server{
		Addr:              ":5678",
		Handler:           http.HandlerFunc(httpResponseHandle),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       10 * time.Second,
	})
	// Then just use it like http.Server .

	testPrint(srv)

	err := srv.ListenAndServeTLS(
		filepath.Join(rootPath, "localhost.crt"),
		filepath.Join(rootPath, "localhost.key"),
	)
	fmt.Println(err)
}

func httpResponseHandle(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(rootPath, "web", r.URL.Path))
}

func getRootPath() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	rootPath = filepath.Dir(ex)
}

func testPrint(srv *hlfhr.Server) {
	p := "\n  test:\n  "
	if runtime.GOOS == "windows" {
		p += "cmd /C "
	}
	p = fmt.Sprint(p, "curl -v -k -L http://localhost", srv.Addr, "/\n")
	fmt.Println(p)
}
