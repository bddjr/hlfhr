package hlfhr

import "net/http"

func Redirect(w http.ResponseWriter, code int, url string) {
	w.Header().Set("Location", url)
	w.WriteHeader(code)
}

func RedirectToHttps(w http.ResponseWriter, r *http.Request, code int) {
	url := "https://" + r.Host + r.URL.Path
	if r.URL.ForceQuery || r.URL.RawQuery != "" {
		url += "?" + r.URL.RawQuery
	}
	Redirect(w, code, url)
}
