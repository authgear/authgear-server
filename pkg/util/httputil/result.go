package httputil

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/util/otelutil"
)

type Result interface {
	WriteResponse(rw http.ResponseWriter, r *http.Request)
	IsInternalError() bool
}

// InternalRedirectResult is a redirect result that is only for internal use,
// for example, redirecting to the auth ui from authorization endpoint.
type InternalRedirectResult struct {
	Cookies []*http.Cookie
	URL     string
}

func (re *InternalRedirectResult) WriteResponse(rw http.ResponseWriter, r *http.Request) {
	for _, cookie := range re.Cookies {
		UpdateCookie(rw, cookie)
	}

	Redirect(r.Context(), rw, r, re.URL, http.StatusFound)
}

func (re *InternalRedirectResult) IsInternalError() bool {
	return false
}

func Redirect(ctx context.Context, w http.ResponseWriter, r *http.Request, redirectURI string, statusCode int) {
	http.Redirect(w, r, ConstructInternalRedirectURI(ctx, redirectURI), statusCode)
}

// ConstructInternalRedirectURI is to construct a redirect uri that is only for internal use,
// for example, redirecting to the auth ui from authorization endpoint.
func ConstructInternalRedirectURI(ctx context.Context, redirectURI string) string {
	u, err := url.Parse(redirectURI)
	if err != nil {
		panic(fmt.Errorf("unexpected: failed to parse redirect uri: %w", err))
	}

	u = otelutil.InjectTraceContextToURL(ctx, u)
	return u.String()
}
