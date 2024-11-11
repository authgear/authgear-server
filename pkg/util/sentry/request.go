package sentry

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

var HeaderWhiteList = []string{
	"Origin",
	"Referer",
	"User-Agent",
	"X-Original-For",
	"X-Forwarded-For",
	"X-Real-IP",
	"Forwarded",
}

func MakeMinimalRequest(r *http.Request, trustProxy bool) (req *http.Request) {
	u := *r.URL
	u.Scheme = httputil.GetProto(r, trustProxy)
	u.Host = httputil.GetHost(r, trustProxy)

	ctx := r.Context()
	// Detach the deadline so that the context is not canceled along with the request.
	ctx = context.WithoutCancel(ctx)
	req, _ = http.NewRequestWithContext(ctx, r.Method, u.String(), nil)

	for _, name := range HeaderWhiteList {
		if header := r.Header.Get(name); header != "" {
			req.Header.Set(name, header)
		}
	}

	return
}
