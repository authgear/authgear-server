package middleware

import (
	"net/http"
	"strconv"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/config"
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
		tenantConfig := config.GetTenantConfig(r)
		accessKey := m.AuthContext.AccessKey()

		model.SetAccessKey(r, accessKey)

		// Remove untrusted headers first.
		r.Header.Del(coreHttp.HeaderUserID)
		r.Header.Del(coreHttp.HeaderUserVerified)
		r.Header.Del(coreHttp.HeaderUserDisabled)
		r.Header.Del(coreHttp.HeaderSessionIdentityType)
		r.Header.Del(coreHttp.HeaderSessionAuthenticatorType)

		// If refresh token is enabled and the session is invalid,
		// do not forward the request and write `x-skygear-try-refresh-token: true`
		authInfo, err := m.AuthContext.AuthInfo()
		if err == session.ErrSessionNotFound {
			if accessKey.ClientID != "" {
				clientConfig, ok := model.GetClientConfig(tenantConfig, accessKey.ClientID)
				if ok && !clientConfig.RefreshTokenDisabled {
					w.Header().Set(coreHttp.HeaderTryRefreshToken, "true")
					w.WriteHeader(401)
					return
				}
			}
		}

		if authInfo != nil {
			id := authInfo.ID
			disabled := authInfo.Disabled
			verified := authInfo.Verified

			r.Header.Set(coreHttp.HeaderUserID, id)
			r.Header.Set(coreHttp.HeaderUserVerified, strconv.FormatBool(verified))
			r.Header.Set(coreHttp.HeaderUserDisabled, strconv.FormatBool(disabled))
		}
		sess, _ := m.AuthContext.Session()
		if sess != nil {
			ptype := sess.PrincipalType
			if ptype != "" {
				r.Header.Set(coreHttp.HeaderSessionIdentityType, string(ptype))
			}
			atype := sess.AuthenticatorType
			if atype != "" {
				r.Header.Set(coreHttp.HeaderSessionAuthenticatorType, string(atype))
			}
		}

		next.ServeHTTP(w, r)
	})
}
