package sso

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"

	"github.com/skygeario/skygear-server/pkg/auth"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachAuthURLHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/sso/{provider}/login_auth_url", &AuthURLHandlerFactory{
		Dependency: authDependency,
		Action:     "login",
	}).Methods("OPTIONS", "POST")
	server.Handle("/sso/{provider}/link_auth_url", &AuthURLHandlerFactory{
		Dependency: authDependency,
		Action:     "link",
	}).Methods("OPTIONS", "POST")
	return server
}

type AuthURLHandlerFactory struct {
	Dependency auth.DependencyMap
	Action     string
}

func (f AuthURLHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &AuthURLHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	vars := mux.Vars(request)
	h.ProviderName = vars["provider"]
	h.Provider = h.ProviderFactory.NewProvider(h.ProviderName)
	h.Action = f.Action
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f AuthURLHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

// AuthURLRequestPayload login handler request payload
type AuthURLRequestPayload struct {
	Scope       []string               `json:"scope"`
	Options     map[string]interface{} `json:"options"`
	CallbackURL string                 `json:"callback_url"`
	RawUXMode   string                 `json:"ux_mode"`
	UXMode      sso.UXMode
}

// Validate request payload
func (p AuthURLRequestPayload) Validate() error {
	if p.CallbackURL == "" {
		return skyerr.NewInvalidArgument("Callback url is required", []string{"callback_url"})
	}

	if p.UXMode == sso.Undefined {
		return skyerr.NewInvalidArgument("UX mode is required", []string{"ux_mode"})
	}

	return nil
}

// AuthURLHandler returns the SSO auth url by provider.
//
// curl \
//   -X POST \
//   -H "Content-Type: application/json" \
//   -H "X-Skygear-Api-Key: API_KEY" \
//   -d @- \
//   http://localhost:3000/sso/<provider>/login_auth_url \
// <<EOF
// {
//     "scope": ["openid", "profile"],
//     "options": {
//       "prompt": "select_account"
//     },
//     callback_url: <url>,
//     ux_mode: <ux_mode>
// }
// EOF
//
// {
//     "result": "<auth_url>"
// }
//
// curl \
//   -X POST \
//   -H "Content-Type: application/json" \
//   -H "X-Skygear-Api-Key: API_KEY" \
//   -d @- \
//   http://localhost:3000/sso/<provider>/link_auth_url \
// <<EOF
// {
//     "scope": ["openid", "profile"],
//     "options": {
//       "prompt": "select_account"
//     },
//     callback_url: <url>,
//     ux_mode: <ux_mode>
// }
// EOF
//
// {
//     "result": "<auth_url>"
// }
type AuthURLHandler struct {
	TxContext       db.TxContext           `dependency:"TxContext"`
	AuthContext     coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	ProviderFactory sso.ProviderFactory    `dependency:"SSOProviderFactory"`
	Provider        sso.Provider
	ProviderName    string
	Action          string
}

func (h AuthURLHandler) WithTx() bool {
	return true
}

func (h AuthURLHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := AuthURLRequestPayload{
		// avoid nil pointer
		Scope:   make([]string, 0),
		Options: make(sso.Options),
	}
	err := json.NewDecoder(request.Body).Decode(&payload)
	payload.UXMode = sso.UXModeFromString(payload.RawUXMode)

	return payload, err
}

func (h AuthURLHandler) Handle(req interface{}) (resp interface{}, err error) {
	if h.Provider == nil {
		err = skyerr.NewInvalidArgument("Provider is not supported", []string{h.ProviderName})
		return
	}
	payload := req.(AuthURLRequestPayload)
	params := sso.GetURLParams{
		Scope:       payload.Scope,
		Options:     payload.Options,
		CallbackURL: payload.CallbackURL,
		UXMode:      payload.UXMode,
		Action:      h.Action,
	}
	if h.AuthContext.AuthInfo() != nil {
		params.UserID = h.AuthContext.AuthInfo().ID
	}
	url, err := h.Provider.GetAuthURL(params)
	if err != nil {
		return
	}
	resp = url
	return
}
