package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/apiclientconfig"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

// AuthnMiddleware populate auth context information
type AuthnMiddleware struct {
	APIClientConfigurationProvider apiclientconfig.Provider `dependency:"APIClientConfigurationProvider"`
	AuthContextSetter              auth.ContextSetter       `dependency:"AuthContextSetter"`
	SessionProvider                session.Provider         `dependency:"SessionProvider"`
	SessionWriter                  session.Writer           `dependency:"SessionWriter"`
	AuthInfoStore                  authinfo.Store           `dependency:"AuthInfoStore"`
	TxContext                      db.TxContext             `dependency:"TxContext"`
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
		sess, authInfo, err := m.resolve(r)

		if errors.Is(err, model.ErrTokenConflict) {
			// Clear session if token conflicts
			m.SessionWriter.ClearSession(w)
			// Treat as session not found
			err = errors.WithSecondaryError(session.ErrSessionNotFound, err)

		} else if errors.Is(err, session.ErrSessionNotFound) {
			// Clear session if session is not found
			m.SessionWriter.ClearSession(w)

		} else if err != nil {
			panic(err)
		}

		m.AuthContextSetter.SetSessionAndAuthInfo(sess, authInfo, err)

		next.ServeHTTP(w, r)
	})
}

func (m *AuthnMiddleware) resolve(r *http.Request) (s *auth.Session, info *authinfo.AuthInfo, err error) {
	tenantConfig := config.GetTenantConfig(r)

	key := m.APIClientConfigurationProvider.GetAccessKeyByAPIKey(model.GetAPIKey(r))
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

	sess, err := m.SessionProvider.GetByToken(accessToken, auth.SessionTokenKindAccessToken)
	if err != nil {
		return
	}

	if tenantConfig.UserConfig.Clients[sess.ClientID].SessionTransport != transport {
		err = session.ErrSessionNotFound
		return
	}

	ai := authinfo.AuthInfo{}
	err = m.AuthInfoStore.GetAuth(sess.UserID, &ai)
	if err != nil {
		if errors.Is(err, authinfo.ErrNotFound) {
			err = session.ErrSessionNotFound
		}
		return
	}
	s = sess
	info = &ai

	// in case valid session is used, infer access key from session
	if !key.IsMasterKey() {
		key = model.NewAccessKey(sess.ClientID)
		m.AuthContextSetter.SetAccessKey(key)
	}

	// should not use new session data in context
	sessionCopy := *s
	err = m.SessionProvider.Access(&sessionCopy)
	return
}
