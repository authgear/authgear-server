package session

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
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
}

func (h *ResolveHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := session.GetContext(r.Context())
	fmt.Printf("%#v\n", ctx)
}
