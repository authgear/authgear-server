package session

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func AttachResolveHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/session/get").
		Handler(auth.MakeHandler(authDependency, newResolveHandler))
}

type ResolveHandler struct {
	TimeProvider time.Provider
}

func (h *ResolveHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := session.GetContext(r.Context())
	ctx.ToAuthnInfo(h.TimeProvider.NowUTC()).PopulateHeaders(rw)
}
