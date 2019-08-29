package middleware

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/model"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// AuthnMiddleware populate auth context information
type AuthnMiddleware struct {
	AuthContextSetter auth.ContextSetter `dependency:"AuthContextSetter"`
	SessionProvider   session.Provider   `dependency:"SessionProvider"`
	AuthInfoStore     authinfo.Store     `dependency:"AuthInfoStore"`
	TxContext         db.TxContext       `dependency:"TxContext"`
}

// AuthnMiddlewareFactory creates AuthnMiddleware per request.
type AuthnMiddlewareFactory struct{}

// NewInjectableMiddleware implements InjectableMiddlewareFactory.
func (f AuthnMiddlewareFactory) NewInjectableMiddleware() InjectableMiddleware {
	return &AuthnMiddleware{}
}

// Handle implements InjectableMiddleware.
func (m *AuthnMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer func() {
			if err == nil {
				next.ServeHTTP(w, r)
			} else {
				// clear session cookie if error occurred
				cookie := &http.Cookie{
					Name:    coreHttp.CookieNameSession,
					Path:    "/",
					Expires: time.Unix(0, 0),
				}
				http.SetCookie(w, cookie)

				skyErr := skyerr.NewNotAuthenticatedErr()
				httpStatus := skyerr.ErrorDefaultStatusCode(skyErr)
				response := handler.APIResponse{Err: skyErr}
				encoder := json.NewEncoder(w)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(httpStatus)
				encoder.Encode(response)
			}
		}()

		tenantConfig := config.GetTenantConfig(r)

		key := model.CheckAccessKey(tenantConfig, model.GetAPIKey(r))
		m.AuthContextSetter.SetAccessKey(key)

		accessToken, transport, err := model.GetAccessToken(r)
		if err != nil {
			// invalid session token -> must not proceed

			if err == model.ErrTokenConflict {
				// clear session cookie if session token is conflicted
				cookie := &http.Cookie{
					Name:    coreHttp.CookieNameSession,
					Path:    "/",
					Expires: time.Unix(0, 0),
				}
				http.SetCookie(w, cookie)
			}
			return
		}

		// No access token found. Simply proceed.
		if accessToken == "" {
			return
		}

		if err = m.TxContext.BeginTx(); err != nil {
			panic(err)
		}
		defer m.TxContext.RollbackTx()

		s, err := m.SessionProvider.GetByToken(accessToken, auth.SessionTokenKindAccessToken)
		if err != nil {
			// session not found -> treat as no access token is provided
			err = nil
			return
		}

		if tenantConfig.UserConfig.Clients[s.ClientID].SessionTransport != transport {
			// inconsistent session token transport -> must not proceed
			err = session.ErrSessionNotFound
			return
		}

		authInfo := authinfo.AuthInfo{}
		err = m.AuthInfoStore.GetAuth(s.UserID, &authInfo)
		if err != nil {
			if err == skydb.ErrUserNotFound {
				// user not found -> treat as no access token is provided
				err = nil
			}
			return
		}

		// in case valid session is used, infer access key from session
		key = model.NewAccessKey(s.ClientID)
		m.AuthContextSetter.SetAccessKey(key)

		// should not use current session in context
		sessionCopy := *s
		err = m.SessionProvider.Access(&sessionCopy)
		if err != nil {
			// cannot access session -> must not proceed
			return
		}

		m.AuthContextSetter.SetSession(s)
		m.AuthContextSetter.SetAuthInfo(&authInfo)
	})
}
