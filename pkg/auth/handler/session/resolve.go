package session

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/anonymous"
	coreauth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/logging"
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

//go:generate mockgen -source=resolve.go -destination=resolve_mock_test.go -package session

type AnonymousIdentityProvider interface {
	List(userID string) ([]*anonymous.Identity, error)
}

type ResolveHandler struct {
	TimeProvider  time.Provider
	Anonymous     AnonymousIdentityProvider
	LoggerFactory logging.Factory
}

func (h *ResolveHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	info, err := h.resolve(r)
	if err != nil {
		logger := h.LoggerFactory.NewLogger("resolve-handler")
		logger.WithError(err).Error("failed to resolve user")

		http.Error(rw, "internal error", http.StatusInternalServerError)
		return
	}
	info.PopulateHeaders(rw)

	accessKey := coreauth.GetAccessKey(r.Context())
	accessKey.WriteTo(rw)
}

func (h *ResolveHandler) resolve(r *http.Request) (*authn.Info, error) {
	valid := auth.IsValidAuthn(r.Context())
	user := auth.GetUser(r.Context())
	session := auth.GetSession(r.Context())

	var info *authn.Info
	if valid && user != nil && session != nil {
		anonIdentities, err := h.Anonymous.List(user.ID)
		if err != nil {
			return nil, err
		}
		info = authn.NewAuthnInfo(session.AuthnAttrs(), user, len(anonIdentities) > 0)
	} else if !valid {
		info = &authn.Info{IsValid: false}
	}

	return info, nil
}
