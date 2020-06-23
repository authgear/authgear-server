package oauth

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc"
	"github.com/skygeario/skygear-server/pkg/db"
	"github.com/skygeario/skygear-server/pkg/log"
)

func ConfigureUserInfoHandler(router *mux.Router, h http.Handler) {
	router.NewRoute().
		Path("/oauth2/userinfo").
		Methods("GET", "POST", "OPTIONS").
		Handler(oauth.RequireScope(h))
}

type oauthUserInfoProvider interface {
	LoadUserClaims(auth.AuthSession) (*oidc.UserClaims, error)
}

type UserInfoHandler struct {
	Logger           *log.Logger
	DBContext        db.Context
	UserInfoProvider oauthUserInfoProvider
}

func (h *UserInfoHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r.Context())
	var claims *oidc.UserClaims
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
