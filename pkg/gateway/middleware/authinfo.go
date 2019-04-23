package middleware

import (
	"net/http"
	"strconv"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

const (
	httpHeaderAuthInfoID       = "x-skygear-auth-userid"
	httpHeaderAuthInfoVerified = "x-skygear-auth-verified"
	httpHeaderAuthInfoDisabled = "x-skygear-auth-disabled"
)

// AuthInfoMiddleware injects auth info headers into the request
// if x-skygear-access-token is present in the request.
type AuthInfoMiddleware struct {
	TokenStore    authtoken.Store `dependency:"TokenStore"`
	AuthInfoStore authinfo.Store  `dependency:"AuthInfoStore"`
	TxContext     db.TxContext    `dependency:"TxContext"`
}

// Handle implements InjectableMiddleware.
func (m *AuthInfoMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer next.ServeHTTP(w, r)

		// Remove untrusted headers first.
		r.Header.Del(httpHeaderAuthInfoID)
		r.Header.Del(httpHeaderAuthInfoVerified)
		r.Header.Del(httpHeaderAuthInfoDisabled)

		accessToken := model.GetAccessToken(r)
		// No access token found. Simply proceed.
		if accessToken == "" {
			return
		}

		if err := m.TxContext.BeginTx(); err != nil {
			panic(err)
		}
		defer m.TxContext.RollbackTx()

		token := authtoken.Token{}
		err := m.TokenStore.Get(accessToken, &token)
		if err != nil {
			http.Error(w, "invalid access token", http.StatusUnauthorized)
			return
		}

		authInfo := authinfo.AuthInfo{}
		err = m.AuthInfoStore.GetAuth(token.AuthInfoID, &authInfo)
		if err != nil {
			http.Error(w, "invalid access token", http.StatusUnauthorized)
			return
		}

		id := authInfo.ID
		disabled := authInfo.Disabled
		verified := authInfo.Verified

		r.Header.Set(httpHeaderAuthInfoID, id)
		r.Header.Set(httpHeaderAuthInfoVerified, strconv.FormatBool(verified))
		r.Header.Set(httpHeaderAuthInfoDisabled, strconv.FormatBool(disabled))
	})
}
