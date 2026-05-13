package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bddjr/hlfhr"
)

func tlsVersionName(version uint16) string {
	switch version {
	case 0x0301:
		return "1.0"
	case 0x0302:
		return "1.1"
	case 0x0303:
		return "1.2"
	case 0x0304:
		return "1.3"
	default:
		return fmt.Sprintf("Unknown 0x%04X", version)
	}
}

func main() {
	println("\n  http://127.0.0.1\n  http://127.0.0.1:443\n  http://localhost:443\n")

	srv := hlfhr.New(&http.Server{
		Addr: "127.0.0.1:443",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			enc := json.NewEncoder(w)
			enc.SetEscapeHTML(false)
			enc.SetIndent("", "  ")
			err := enc.Encode(map[string]interface{}{
				"Proto":          r.Proto,
				"TLS_Version":    tlsVersionName(r.TLS.Version),
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
