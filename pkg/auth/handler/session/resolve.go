package session

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func AttachResolveHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/session/resolve").
		Handler(auth.MakeHandler(authDependency, newResolveHandler))
}

type ResolveHandler struct {
	TimeProvider time.Provider
}

func (h *ResolveHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	valid := authn.IsValidAuthn(r.Context())
	user := authn.GetUser(r.Context())
	session := authn.GetSession(r.Context())

	var info *authn.Info
	if valid && user != nil && session != nil {
		info = authn.NewAuthnInfo(h.TimeProvider.NowUTC(), session.AuthnAttrs(), user)
	} else if !valid {
		info = &authn.Info{IsValid: false}
	}
	info.PopulateHeaders(rw)
}
