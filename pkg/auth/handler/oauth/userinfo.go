package oauth

import (
	"encoding/json"
	"net/http"

	"github.com/lestrrat-go/jwx/jwt"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/db"
	"github.com/skygeario/skygear-server/pkg/httproute"
	"github.com/skygeario/skygear-server/pkg/log"
)

func ConfigureUserInfoRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST", "OPTIONS").
		WithPathPattern("/oauth2/userinfo")
}

type ProtocolUserInfoProvider interface {
	LoadUserClaims(auth.AuthSession) (jwt.Token, error)
}

type UserInfoHandlerLogger struct{ *log.Logger }

func NewUserInfoHandlerLogger(lf *log.Factory) UserInfoHandlerLogger {
	return UserInfoHandlerLogger{lf.New("handler-user-info")}
}

type UserInfoHandler struct {
	Logger           UserInfoHandlerLogger
	DBContext        db.Context
	UserInfoProvider ProtocolUserInfoProvider
}

func (h *UserInfoHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r.Context())
	var claims jwt.Token
	err := db.WithTx(h.DBContext, func() (err error) {
		claims, err = h.UserInfoProvider.LoadUserClaims(session)
		return
	})

	if err != nil {
		h.Logger.WithError(err).Error("oidc userinfo handler failed")
		http.Error(rw, "Internal Server Error", 500)
		return
	}

	rw.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(rw)
	err = encoder.Encode(claims)
	if err != nil {
		http.Error(rw, err.Error(), 500)
	}
}
