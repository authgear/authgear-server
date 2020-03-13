package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/model"
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
		sess, authInfo, err := m.resolve(r)

		if errors.Is(err, session.ErrSessionNotFound) {
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
	tenantConfig := config.GetTenantConfig(r.Context())

	accessToken := coreHttp.GetSessionIdentifier(r)
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

	_, ok := model.GetClientConfig(tenantConfig.AppConfig.Clients, sess.ClientID)
	if !ok {
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
	accessKey := auth.GetAccessKey(r.Context())
	if !accessKey.IsMasterKey {
		client, ok := model.GetClientConfig(tenantConfig.AppConfig.Clients, sess.ClientID)
		if ok {
			accessKey.Client = client
			auth.WithAccessKey(r.Context(), accessKey)
		}
	}

	// should not use new session data in context
	sessionCopy := *s
	err = m.SessionProvider.Access(&sessionCopy)
	return
}
