package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/bddjr/hlfhr"
)

var srv *hlfhr.Server

func main() {
	// Use hlfhr.New
	srv = hlfhr.New(&http.Server{
		Addr:              ":5678",
		Handler:           http.HandlerFunc(httpResponseHandle),
		ReadHeaderTimeout: time.Minute,
	})
	// Then just use it like http.Server .

	fmt.Println("IsShuttingDown:", srv.IsShuttingDown())
	testPrint(srv)
	fmt.Println("Press Ctrl+C close server")

	go func() {
		err := srv.ListenAndServeTLS("localhost.crt", "localhost.key")
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	<-c
	srv.Close()

	fmt.Println("IsShuttingDown:", srv.IsShuttingDown())
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
	fmt.Print("\n  test:\n  curl")
	if runtime.GOOS == "windows" {
		fmt.Print(".exe")
	}
	fmt.Print(" -v -k -L http://localhost", srv.Addr, "/\n\n")
}
