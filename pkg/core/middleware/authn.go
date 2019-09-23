package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/logging"

	"github.com/skygeario/skygear-server/pkg/core/apiclientconfig"
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
	APIClientConfigurationProvider apiclientconfig.Provider `dependency:"APIClientConfigurationProvider"`
	AuthContextSetter              auth.ContextSetter       `dependency:"AuthContextSetter"`
	LoggerFactory                  logging.Factory          `dependency:"LoggerFactory"`
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
	log := m.LoggerFactory.NewLogger("authn")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sess *auth.Session
		var authInfo *authinfo.AuthInfo
		var err error
		defer func() {
			m.AuthContextSetter.SetSessionAndAuthInfo(sess, authInfo, err)

			if err == model.ErrTokenConflict {
				// Clear session if token conflicts
				m.SessionWriter.ClearSession(w)
			} else if err == session.ErrSessionNotFound {
				// Clear session if session is not found
				m.SessionWriter.ClearSession(w)
			} else if err != nil {
				log.WithError(err).Error("failed to resolve session")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			next.ServeHTTP(w, r)
		}()

		tenantConfig := config.GetTenantConfig(r)

		key := m.APIClientConfigurationProvider.AccessKey(model.GetAPIKey(r))
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

		sess, err = m.SessionProvider.GetByToken(accessToken, auth.SessionTokenKindAccessToken)
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
			if err == skydb.ErrUserNotFound {
				err = session.ErrSessionNotFound
			}
			return
		}
		authInfo = &ai

		// in case valid session is used, infer access key from session
		if !key.IsMasterKey() {
			key = model.NewAccessKey(sess.ClientID)
			m.AuthContextSetter.SetAccessKey(key)
		}

		// should not use new session data in context
		sessionCopy := *sess
		err = m.SessionProvider.Access(&sessionCopy)
		if err != nil {
			return
		}
	})
}
