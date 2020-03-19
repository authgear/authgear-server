package oauth

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/handler"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
)

func AttachAuthorizeHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/oauth2/authorize").
		Handler(pkg.MakeHandler(authDependency, newAuthorizeHandler)).
		Methods("GET", "POST")
}

type oauthAuthorizeHandler interface {
	Handle(r protocol.AuthorizationRequest) handler.AuthorizationResult
}

type AuthorizeHandler struct {
	authzHandler oauthAuthorizeHandler
}

func (h *AuthorizeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}

	req := protocol.AuthorizationRequest{}
	for name, values := range r.Form {
		req[name] = values[0]
	}

	result := h.authzHandler.Handle(req)
	result.WriteResponse(rw, r)
}
