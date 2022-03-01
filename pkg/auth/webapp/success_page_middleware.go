package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/successpage"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type SuccessPageMiddleware struct {
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
		currentPath := r.URL.Path
		pathInCookie := m.Pop(r, w)

		if currentPath != pathInCookie {
			// Show invalid session error when the path cookie doesn't match
			// the current path
			apierror := apierrors.AsAPIError(ErrInvalidSession)
			errorCookie, err := m.ErrorCookie.SetError(r, apierror)
			if err != nil {
				panic(err)
			}
			httputil.UpdateCookie(w, errorCookie)
			http.Redirect(w, r, "/error", http.StatusFound)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
