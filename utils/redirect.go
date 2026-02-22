package hlfhr_utils

import (
	"net/http"
	"strings"
)

// Redirect without HTTP body.
func Redirect(w http.ResponseWriter, code int, url string) {
	w.Header()["Location"] = []string{url}
	w.WriteHeader(code)
}

// Redirect without HTTP body.
func RedirectToHttps(w http.ResponseWriter, r *http.Request, code int) {
	host, _ := strings.CutSuffix(r.Host, ":80")
	redirectToHttps(w, r, code, host)
}

func RedirectToHttps_ForceSamePort(w http.ResponseWriter, r *http.Request, code int) {
	host := r.Host
	if r.TLS == nil {
		// http:
		notHasPortNumber := true
		if host[len(host)-1] != ']' {
			i := len(host) - 2
			j := i - 5
			if j < 0 {
				j = 0
			}
			for ; i > j; i-- {
				if host[i] == ':' {
					// It has port number.
					notHasPortNumber = false
					break
				}
			}
		}
		if notHasPortNumber {
			host += ":80"
		}

	}
	redirectToHttps(w, r, code, host)
}

func redirectToHttps(w http.ResponseWriter, r *http.Request, code int, host string) {
	url := "https://" + host + r.URL.Path
	if r.URL.ForceQuery || r.URL.RawQuery != "" {
		url += "?" + r.URL.RawQuery
	}
	Redirect(w, code, url)
}
