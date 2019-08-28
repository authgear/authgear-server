package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/config"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/db"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

// AuthInfoMiddleware injects auth info headers into the request
// if x-skygear-access-token is present in the request.
type AuthInfoMiddleware struct {
	SessionProvider session.Provider                         `dependency:"SessionProvider"`
	AuthInfoStore   authinfo.Store                           `dependency:"AuthInfoStore"`
	TxContext       db.TxContext                             `dependency:"TxContext"`
	ClientConfigs   map[string]config.APIClientConfiguration `dependency:"ClientConfigs"`
}

// AuthInfoMiddlewareFactory creates AuthInfoMiddleware per request.
type AuthInfoMiddlewareFactory struct{}

// NewInjectableMiddleware implements InjectableMiddlewareFactory.
func (f AuthInfoMiddlewareFactory) NewInjectableMiddleware() InjectableMiddleware {
	return &AuthInfoMiddleware{}
}

// Handle implements InjectableMiddleware.
func (m *AuthInfoMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer func() {
			if err == nil {
				next.ServeHTTP(w, r)
			}
		}()

		// Remove untrusted headers first.
		r.Header.Del(coreHttp.HeaderAuthInfoID)
		r.Header.Del(coreHttp.HeaderAuthInfoVerified)
		r.Header.Del(coreHttp.HeaderAuthInfoDisabled)

		accessToken, transport, err := model.GetAccessToken(r)
		if err == model.ErrTokenConflict {
			err = nil
			// clear session cookie if failed to get access token
			cookie := &http.Cookie{
				Name:    coreHttp.CookieNameSession,
				Path:    "/",
				Expires: time.Unix(0, 0),
			}
			http.SetCookie(w, cookie)
			accessToken = ""
		} else if err != nil {
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

		session, err := m.SessionProvider.GetByToken(accessToken, auth.SessionTokenKindAccessToken)
		if err != nil {
			http.Error(w, "invalid access token", http.StatusUnauthorized)
			return
		}

		if m.ClientConfigs[session.ClientID].SessionTransport != transport {
			http.Error(w, "invalid access token", http.StatusUnauthorized)
			return
		}

		authInfo := authinfo.AuthInfo{}
		err = m.AuthInfoStore.GetAuth(session.UserID, &authInfo)
		if err != nil {
			http.Error(w, "invalid access token", http.StatusUnauthorized)
			return
		}

		id := authInfo.ID
		disabled := authInfo.Disabled
		verified := authInfo.Verified

		r.Header.Set(coreHttp.HeaderAuthInfoID, id)
		r.Header.Set(coreHttp.HeaderAuthInfoVerified, strconv.FormatBool(verified))
		r.Header.Set(coreHttp.HeaderAuthInfoDisabled, strconv.FormatBool(disabled))

		// in case valid session is used, infer access key from session
		accessKey := model.NewAccessKey(session.ClientID)
		model.SetAccessKey(r, accessKey)

		err = m.SessionProvider.Access(session)
	})
}
