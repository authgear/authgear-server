package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/model"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
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

		// Remove untrusted headers first.
		r.Header.Del(coreHttp.HeaderAuthInfoID)
		r.Header.Del(coreHttp.HeaderAuthInfoVerified)
		r.Header.Del(coreHttp.HeaderAuthInfoDisabled)

		accessToken, transport, err := model.GetAccessToken(r)
		if err != nil {
			// invalid session token -> must not proceed
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

		if m.ClientConfigs[s.ClientID].SessionTransport != transport {
			// inconsistent session token transport -> must not proceed
			err = session.ErrSessionNotFound
			return
		}

		authInfo := authinfo.AuthInfo{}
		err = m.AuthInfoStore.GetAuth(s.UserID, &authInfo)
		if err != nil {
			// user not found -> treat as no access token is provided
			return
		}

		id := authInfo.ID
		disabled := authInfo.Disabled
		verified := authInfo.Verified

		r.Header.Set(coreHttp.HeaderAuthInfoID, id)
		r.Header.Set(coreHttp.HeaderAuthInfoVerified, strconv.FormatBool(verified))
		r.Header.Set(coreHttp.HeaderAuthInfoDisabled, strconv.FormatBool(disabled))

		// in case valid session is used, infer access key from session
		accessKey := model.NewAccessKey(s.ClientID)
		model.SetAccessKey(r, accessKey)

		err = m.SessionProvider.Access(s)
	})
}
