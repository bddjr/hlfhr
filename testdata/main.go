package main

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bddjr/hlfhr"
)

func main() {
	println("\n  http://127.0.0.1\n  http://127.0.0.1:443\n  http://localhost:443\n")

	srv := hlfhr.New(&http.Server{
		Addr: "127.0.0.1:443",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			enc := json.NewEncoder(w)
			enc.SetEscapeHTML(false)
			enc.SetIndent("", "  ")
			tlsVersion, _ := strings.CutPrefix(tls.VersionName(r.TLS.Version), "TLS ")
			err := enc.Encode(map[string]interface{}{
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
	err := srv.ListenAndServeTLS("invalid.crt", "invalid.key")
	if err != nil {
		panic(err)
	}
}
