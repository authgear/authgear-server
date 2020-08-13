package sentry

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

var HeaderWhiteList = []string{
	"Referer",
	"User-Agent",
	"X-Forwarded-For",
	"X-Real-IP",
	"Forwarded",
}

func MakeMinimalRequest(r *http.Request, trustProxy bool) (req *http.Request) {
	u := *r.URL
	u.Scheme = httputil.GetProto(r, trustProxy)
	u.Host = httputil.GetHost(r, trustProxy)

	req, _ = http.NewRequest(r.Method, u.String(), nil)

	for _, name := range HeaderWhiteList {
		if header := r.Header.Get(name); header != "" {
			req.Header.Set(name, header)
		}
	}

	return
}
