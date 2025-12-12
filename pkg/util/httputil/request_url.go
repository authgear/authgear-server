package httputil

import (
	"net/http"
	"net/url"
)

type HTTPRequestURL string

func GetRequestURL(r *http.Request, proto HTTPProto, host HTTPHost) HTTPRequestURL {

	if r == nil {
		return ""
	}

	u := url.URL{
		Scheme:   string(proto),
		Host:     string(host),
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
		Fragment: r.URL.Fragment,
	}

	return HTTPRequestURL(u.String())
}
