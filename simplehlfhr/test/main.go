package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bddjr/hlfhr/simplehlfhr"
)

func main() {
	srv := &http.Server{
		Addr:              ":5677",
		Handler:           http.HandlerFunc(httpResponseHandle),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       5 * time.Second,
	}

	fmt.Print("\n  test:\n",
		"  curl -v http://localhost", srv.Addr, "/\n",
		"  curl -v -k https://localhost", srv.Addr, "/\n\n",
	)

	err := simplehlfhr.ListenAndServeTLS(srv, "../../test/localhost.crt", "../../test/localhost.key")
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}

func httpResponseHandle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		io.WriteString(w, "Hello simple hlfhr!\n\n")
	} else {
		http.NotFound(w, r)
	}
}
