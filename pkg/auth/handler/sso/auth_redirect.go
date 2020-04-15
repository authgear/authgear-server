package sso

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/crypto"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// The authorization flow invokes 2 steps.
// The first step is to call /_auth/sso/<provider>/login_auth_url
// The result is /_auth/sso/<provider>/auth_redirect?state=...
// The state query parameter includes all the information we need
// to generate the actual provider authorization URL without any
// headers nor cookies.
// This ensures the flow can be performed by a separate user agent
// that only support GET.

func AttachAuthRedirectHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/sso/{provider}/auth_redirect").
		Handler(auth.MakeHandler(authDependency, newAuthRedirectHandler)).
		Methods("OPTIONS", "GET")
}

type AuthRedirectHandler struct {
	SSOProvider   sso.Provider
	OAuthProvider sso.OAuthProvider
}

func (h *AuthRedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uxMode, result, err := h.Handle(w, r)
	if err != nil {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
		return
	}
	if uxMode == sso.UXModeManual {
		handler.WriteResponse(w, handler.APIResponse{Result: result})
	} else {
		http.Redirect(w, r, result.(string), http.StatusFound)
	}
}

func (h *AuthRedirectHandler) Handle(w http.ResponseWriter, r *http.Request) (uxMode sso.UXMode, result interface{}, err error) {
	if h.OAuthProvider == nil {
		err = skyerr.NewNotFound("unknown provider")
		return
	}

	err = r.ParseForm()
	if err != nil {
		return
	}

	encodedState := r.Form.Get("state")
	state, err := h.SSOProvider.DecodeState(encodedState)
	if err != nil {
		return
	}
	uxMode = state.UXMode

	// Always generate a new nonce to ensure it is unpredictable.
	nonce := sso.GenerateOpenIDConnectNonce()
	state.HashedNonce = crypto.SHA256String(nonce)
	encodedState, err = h.SSOProvider.EncodeState(*state)
	if err != nil {
		return
	}

	url, err := h.OAuthProvider.GetAuthURL(*state, encodedState)
	if err != nil {
		return
	}

	result = url
	return
}
