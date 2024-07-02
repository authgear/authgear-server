package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/successpage"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type SuccessPageMiddlewareEndpointsProvider interface {
	ErrorEndpointURL(uiImpl config.UIImplementation) *url.URL
}

type SuccessPageMiddleware struct {
	Endpoints   SuccessPageMiddlewareEndpointsProvider
	UIConfig    *config.UIConfig
	Cookies     CookieManager
	ErrorCookie *ErrorCookie
}

func (m *SuccessPageMiddleware) Pop(r *http.Request, rw http.ResponseWriter) string {
	cookie, err := m.Cookies.GetCookie(r, successpage.PathCookieDef)
	if err != nil {
		return ""
	}
	path := cookie.Value

	clearCookie := m.Cookies.ClearCookie(successpage.PathCookieDef)
	httputil.UpdateCookie(rw, clearCookie)
	return path
}

// SuccessPageMiddleware check the success path cookie to determine
// whether it is valid to visit the success page
// the cookie should be set right before redirecting to the success page
func (m *SuccessPageMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We want to allow POST in success page.
		// For example, POST in delete account success page to finish settings action.
		if r.Method == "GET" {
			currentPath := r.URL.Path
			pathInCookie := m.Pop(r, w)
			if currentPath != pathInCookie {
				// Show invalid session error when the path cookie doesn't match
				// the current path
				apierror := apierrors.AsAPIError(ErrInvalidSession)
				errorCookie, err := m.ErrorCookie.SetRecoverableError(r, apierror)
				if err != nil {
					panic(err)
				}
				httputil.UpdateCookie(w, errorCookie)
				http.Redirect(w, r, m.Endpoints.ErrorEndpointURL(m.UIConfig.Implementation).Path, http.StatusFound)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
