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
	srv = hlfhr.New(&http.Server{
		Addr:              ":5678",
		Handler:           http.HandlerFunc(httpResponseHandle),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       10 * time.Second,
	})
	// Then just use it like &http.Server .

	testPrint(srv)

	err := srv.ListenAndServeTLS("localhost.crt", "localhost.key")
	if err != nil && err != http.ErrServerClosed {
		fmt.Println(err)
	}
}

func httpResponseHandle(w http.ResponseWriter, r *http.Request) {
	header := w.Header()
	switch r.URL.Path {
	case "/":
		w.WriteHeader(200)
		header.Set("Content-Type", "text/html")
		io.WriteString(w, `
<html><head>
	<meta name="robots" content="noindex"/>
	<style>
		*{ color-scheme: light dark; }
	</style>
</head><body>
	<h1>Hello HTTPS!</h1>
</body></html>
`,
		)
		return
	}
	w.WriteHeader(404)
	header.Set("Content-Type", "text/plain")
	io.WriteString(w, "404 Not Found\n")
}

func testPrint(srv *hlfhr.Server) {
	p := "\n  test:\n  "
	if runtime.GOOS == "windows" {
		p += "cmd /C "
	}
	p = fmt.Sprint(p, "curl -v http://localhost", srv.Addr, "\n")
	fmt.Println(p)
}
