package hlfhr

import (
	"net/http"
	"strings"
)

const defaultRedirectStatus = 307

// Redirect without HTTP body.
func Redirect(w http.ResponseWriter, code int, url string) {
	w.Header()["Location"] = []string{url}
	w.WriteHeader(code)
}

// Redirect without HTTP body.
func RedirectToHttps(w http.ResponseWriter, r *http.Request, code int) {
	host, _ := strings.CutSuffix(r.Host, ":80")
	url := "https://" + host + r.URL.Path
	if r.URL.ForceQuery || r.URL.RawQuery != "" {
		url += "?" + r.URL.RawQuery
	}
	Redirect(w, code, url)
}
