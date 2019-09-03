package middleware

import (
	"net/http"
	"strconv"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	coreMiddleware "github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

// AuthInfoMiddleware injects auth info headers into the request
// if x-skygear-access-token is present in the request.
type AuthInfoMiddleware struct {
	AuthContext auth.ContextGetter `dependency:"AuthContextGetter"`
}

// AuthInfoMiddlewareFactory creates AuthInfoMiddleware per request.
type AuthInfoMiddlewareFactory struct{}

// NewInjectableMiddleware implements InjectableMiddlewareFactory.
func (f AuthInfoMiddlewareFactory) NewInjectableMiddleware() coreMiddleware.InjectableMiddleware {
	return &AuthInfoMiddleware{}
}

// Handle implements InjectableMiddleware.
func (m *AuthInfoMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		model.SetAccessKey(r, m.AuthContext.AccessKey())

		// Remove untrusted headers first.
		r.Header.Del(coreHttp.HeaderAuthInfoID)
		r.Header.Del(coreHttp.HeaderAuthInfoVerified)
		r.Header.Del(coreHttp.HeaderAuthInfoDisabled)

		authInfo := m.AuthContext.AuthInfo()
		if authInfo != nil {
			id := authInfo.ID
			disabled := authInfo.Disabled
			verified := authInfo.Verified

			r.Header.Set(coreHttp.HeaderAuthInfoID, id)
			r.Header.Set(coreHttp.HeaderAuthInfoVerified, strconv.FormatBool(verified))
			r.Header.Set(coreHttp.HeaderAuthInfoDisabled, strconv.FormatBool(disabled))
		}

		next.ServeHTTP(w, r)
	})
}
