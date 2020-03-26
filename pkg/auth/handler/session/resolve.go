package session

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	coreauth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func AttachResolveHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/session/resolve").
		Handler(pkg.MakeHandler(authDependency, newResolveHandler))
}

type ResolveHandler struct {
	TimeProvider time.Provider
}

func (h *ResolveHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	valid := auth.IsValidAuthn(r.Context())
	user := auth.GetUser(r.Context())
	session := auth.GetSession(r.Context())
	accessKey := coreauth.GetAccessKey(r.Context())

	var info *authn.Info
	if valid && user != nil && session != nil {
		info = authn.NewAuthnInfo(session.AuthnAttrs(), user)
	} else if !valid {
		info = &authn.Info{IsValid: false}
	}
	info.PopulateHeaders(rw)

	accessKey.WriteTo(rw)
}
