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
	srv = hlfhr.New(&http.Server{
		// Addr:              ":8443",
		Handler:           http.HandlerFunc(httpResponseHandle),
		ReadHeaderTimeout: time.Minute,
	})
	srv.Listen80RedirectTo443 = true

	fmt.Println("IsShuttingDown:", srv.IsShuttingDown())
	testPrint(srv)
	print("Press Ctrl+C close server\n\n")

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
	time.Sleep(500 * time.Millisecond)
}

func httpResponseHandle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		if r.TLS != nil {
			io.WriteString(w, "Hello hlfhr!\n\n")
		} else {
			io.WriteString(w, "Error: not using HTTPS\n\n")
		}
	} else {
		http.NotFound(w, r)
	}
}

func testPrint(srv *hlfhr.Server) {
	curl := "curl"
	if runtime.GOOS == "windows" {
		curl += ".exe"
	}
	pl := func(addr string) {
		println("  " + curl + " -vkL " + addr)
	}
	print("\n  test:\n")
	pl("http://127.0.0.1" + srv.Addr)
	pl("http://[::1]" + srv.Addr)
	if srv.Listen80RedirectTo443 {
		if srv.Addr == "" {
			pl("http://127.0.0.1:443")
			pl("http://[::1]:443")
		}
	}
	println()
}
