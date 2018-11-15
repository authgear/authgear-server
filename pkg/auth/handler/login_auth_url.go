package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachLoginAuthURLHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/sso/{provider}/login_auth_url", &LoginAuthURLHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type LoginAuthURLHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f LoginAuthURLHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &LoginAuthURLHandler{}
	inject.DefaultInject(h, f.Dependency, request)

	vars := mux.Vars(request)
	h.ProviderName = vars["provider"]
	providers := map[string]sso.Provider{
		"google":    h.GoogleProvider,
		"facebook":  h.FacebookProvider,
		"instagram": h.InstagramProvider,
		"linkedin":  h.LinkedInProvider,
	}
	if provider, ok := providers[h.ProviderName]; ok {
		h.Provider = provider
	}

	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f LoginAuthURLHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

// LoginAuthURLRequestPayload login handler request payload
type LoginAuthURLRequestPayload struct {
	Scope       []string               `json:"scope"`
	Options     map[string]interface{} `json:"options"`
	CallbackURL string                 `json:"callback_url"`
	RawUXMode   string                 `json:"ux_mode"`
	UXMode      sso.UXMode
}

// Validate request payload
func (p LoginAuthURLRequestPayload) Validate() error {
	if p.CallbackURL == "" {
		return skyerr.NewInvalidArgument("Callback url is required", []string{"callback_url"})
	}

	UXModeName := [...]string{"", "web_redirect", "web_popup", "ios", "android"}
	for k, v := range UXModeName {
		if p.RawUXMode == v {
			p.UXMode = sso.UXMode(k)
			break
		}
	}

	if p.UXMode == sso.Undefined {
		return skyerr.NewInvalidArgument("UX mode is required", []string{"ux_mode"})
	}

	return nil
}

// LoginAuthURLHandler returns roles of users specified by user IDs. Users can only
// get his own roles except that administrators can query roles of other users.
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
//
//     },
//     callback_url: <url>,
//     ux_mode: <ux_mode>
// }
// EOF
//
// {
//     "result": {
//         "user_id_1": [
//             "developer",
//         ],
//         "user_id_2": [
//         ],
//     }
// }
type LoginAuthURLHandler struct {
	TxContext         db.TxContext `dependency:"TxContext"`
	GoogleProvider    sso.Provider `dependency:"GoogleSSOProvider"`
	FacebookProvider  sso.Provider `dependency:"FacebookSSOProvider"`
	InstagramProvider sso.Provider `dependency:"InstagramSSOProvider"`
	LinkedInProvider  sso.Provider `dependency:"LinkedInSSOProvider"`
	ProviderName      string
	Provider          sso.Provider
}

func (h LoginAuthURLHandler) WithTx() bool {
	return true
}

func (h LoginAuthURLHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := LoginAuthURLRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	return payload, err
}

func (h LoginAuthURLHandler) Handle(req interface{}) (resp interface{}, err error) {
	/*
		payload := req.(LoginAuthURLRequestPayload)
		roleMap, err := h.AuthInfoStore.LoginAuthURLs(payload.UserIDs)
		if err != nil {
			err = skyerr.NewError(skyerr.UnexpectedError, "LoginAuthURLs failed")
			return
		}
		resp = roleMap
	*/
	params := sso.GetURLParams{}
	url, err := h.Provider.GetAuthURL(params)
	if err != nil {
		return
	}
	resp = url
	return
}
