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
	RedirectToHttps_ModifyHost(w, r, code, strings.TrimSuffix(r.Host, ":80"))
}

func RedirectToHttps_ForceSamePort(w http.ResponseWriter, r *http.Request, code int) {
	host := r.Host
	if r.TLS == nil && !strings.HasSuffix(host, "]") && strings.LastIndexByte(host, ':') == -1 {
		host += ":80"
	}
	RedirectToHttps_ModifyHost(w, r, code, host)
}

func RedirectToHttps_NoCheckPort(w http.ResponseWriter, r *http.Request, code int) {
	RedirectToHttps_ModifyHost(w, r, code, r.Host)
}

func RedirectToHttps_ModifyHost(w http.ResponseWriter, r *http.Request, code int, host string) {
	url := "https://" + host + r.URL.Path
	if r.URL.ForceQuery || r.URL.RawQuery != "" {
		url += "?" + r.URL.RawQuery
	}
	Redirect(w, code, url)
}
