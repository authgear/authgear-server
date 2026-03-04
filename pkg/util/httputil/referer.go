package httputil

import "net/http"

type HTTPReferer string

func GetReferer(r *http.Request) HTTPReferer {
	if r == nil {
		return ""
	}
	return HTTPReferer(r.Header.Get("Referer"))
}
