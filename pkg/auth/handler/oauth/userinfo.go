package oauth

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func AttachUserInfoHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	handler := pkg.MakeHandler(authDependency, newUserInfoHandler)
	handler = oauth.RequireScope(handler)
	router.NewRoute().
		Path("/oauth2/userinfo").
		Handler(handler).
		Methods("GET", "POST", "OPTIONS")
}

type oauthUserInfoProvider interface {
	LoadUserClaims(auth.AuthSession) (*oidc.UserClaims, error)
}

type UserInfoHandler struct {
	logger           *logrus.Entry
	txContext        db.TxContext
	userInfoProvider oauthUserInfoProvider
}

func (h *UserInfoHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r.Context())
	var claims *oidc.UserClaims
	err := db.WithTx(h.txContext, func() (err error) {
		claims, err = h.userInfoProvider.LoadUserClaims(session)
		return
	})

	if err != nil {
		h.logger.WithError(err).Error("oidc userinfo handler failed")
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
