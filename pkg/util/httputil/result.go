package httputil

import "net/http"

type Result interface {
	WriteResponse(rw http.ResponseWriter, r *http.Request)
	IsInternalError() bool
}

type ResultRedirect struct {
	Cookies []*http.Cookie
	URL     string
}

func (re *ResultRedirect) WriteResponse(rw http.ResponseWriter, r *http.Request) {
	for _, cookie := range re.Cookies {
		UpdateCookie(rw, cookie)
	}
	http.Redirect(rw, r, re.URL, http.StatusFound)
}

func (re *ResultRedirect) IsInternalError() bool {
	return false
}
