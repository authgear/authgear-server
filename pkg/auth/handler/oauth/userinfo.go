package oauth

import (
	"encoding/json"
	"net/http"

	"github.com/lestrrat-go/jwx/jwt"

	db "github.com/authgear/authgear-server/pkg/lib/infra/db/tenant"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func ConfigureUserInfoRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST", "OPTIONS").
		WithPathPattern("/oauth2/userinfo")
}

type ProtocolUserInfoProvider interface {
	LoadUserClaims(userID string) (jwt.Token, error)
}

type UserInfoHandlerLogger struct{ *log.Logger }

func NewUserInfoHandlerLogger(lf *log.Factory) UserInfoHandlerLogger {
	return UserInfoHandlerLogger{lf.New("handler-user-info")}
}

type UserInfoHandler struct {
	Logger           UserInfoHandlerLogger
	Database         *db.Handle
	UserInfoProvider ProtocolUserInfoProvider
}

func (h *UserInfoHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	s := session.GetSession(r.Context())
	var claims jwt.Token
	err := h.Database.WithTx(func() (err error) {
		claims, err = h.UserInfoProvider.LoadUserClaims(s.SessionAttrs().UserID)
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
