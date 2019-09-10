package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/model"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
)

// AuthnMiddleware populate auth context information
type AuthnMiddleware struct {
	AuthContextSetter auth.ContextSetter `dependency:"AuthContextSetter"`
	SessionProvider   session.Provider   `dependency:"SessionProvider"`
	SessionWriter     session.Writer     `dependency:"SessionWriter"`
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
			if err == session.ErrSessionNotFound {
				// clear session if session is not found
				m.SessionWriter.ClearSession(w)
			} else if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			next.ServeHTTP(w, r)
		}()

		tenantConfig := config.GetTenantConfig(r)

		key := model.CheckAccessKey(tenantConfig, model.GetAPIKey(r))
		m.AuthContextSetter.SetAccessKey(key)

		accessToken, transport, err := model.GetAccessToken(r)
		if err != nil {
			return
		}

		// No access token found. Simply proceed.
		if accessToken == "" {
			return
		}

		if err = m.TxContext.BeginTx(); err != nil {
			return
		}
		defer m.TxContext.RollbackTx()

		s, err := m.SessionProvider.GetByToken(accessToken, auth.SessionTokenKindAccessToken)
		if err != nil {
			return
		}

		if tenantConfig.UserConfig.Clients[s.ClientID].SessionTransport != transport {
			err = session.ErrSessionNotFound
			return
		}

		authInfo := authinfo.AuthInfo{}
		err = m.AuthInfoStore.GetAuth(s.UserID, &authInfo)
		if err != nil {
			if err == skydb.ErrUserNotFound {
				err = session.ErrSessionNotFound
			}
			return
		}

		// in case valid session is used, infer access key from session
		if !key.IsMasterKey() {
			key = model.NewAccessKey(s.ClientID)
			m.AuthContextSetter.SetAccessKey(key)
		}

		// should not use new session data in context
		sessionCopy := *s
		err = m.SessionProvider.Access(&sessionCopy)
		if err != nil {
			return
		}

		m.AuthContextSetter.SetSession(s)
		m.AuthContextSetter.SetAuthInfo(&authInfo)
	})
}
