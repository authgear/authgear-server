package sso

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/crypto"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
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
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/sso/{provider}/auth_redirect", &AuthRedirectHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "GET")
	return server
}

type AuthRedirectHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f AuthRedirectHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &AuthRedirectHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	vars := mux.Vars(request)
	h.ProviderID = vars["provider"]
	h.OAuthProvider = h.ProviderFactory.NewOAuthProvider(h.ProviderID)
	return h
}

type AuthRedirectHandler struct {
	ProviderFactory *sso.OAuthProviderFactory `dependency:"SSOOAuthProviderFactory"`
	SSOProvider     sso.Provider              `dependency:"SSOProvider"`
	OAuthProvider   sso.OAuthProvider
	ProviderID      string
}

func (h *AuthRedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := h.Handle(w, r)
	if err != nil {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
		return
	}
	http.Redirect(w, r, result.(string), http.StatusFound)
}

func (h *AuthRedirectHandler) Handle(w http.ResponseWriter, r *http.Request) (result interface{}, err error) {
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

	// Always generate a new nonce to ensure it is unpredictable.
	// The developer is expected to call auth_url just before they need to perform the flow.
	// If they call auth_url multiple times ahead of time,
	// only the last auth URL is valid because the nonce of the previous auth URLs are all overwritten.
	nonce := sso.GenerateOpenIDConnectNonce()
	cookie := &http.Cookie{
		Name:     coreHttp.CookieNameOpenIDConnectNonce,
		Value:    nonce,
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	state.Nonce = crypto.SHA256String(nonce)

	url, err := h.OAuthProvider.GetAuthURL(*state)
	if err != nil {
		return
	}

	result = url
	return
}
