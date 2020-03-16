package sso

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	coreauth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachAuthResultHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/sso/auth_result").
		Handler(server.FactoryToHandler(&AuthResultHandlerFactory{
			Dependency: authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type AuthResultHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f *AuthResultHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &AuthResultHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

type AuthResultAuthnProvider interface {
	OAuthExchangeCode(
		client config.OAuthClientConfiguration,
		session *session.Session,
		code *sso.SkygearAuthorizationCode,
	) (authn.Result, error)
}

type AuthResultHandler struct {
	TxContext     db.TxContext            `dependency:"TxContext"`
	RequireAuthz  handler.RequireAuthz    `dependency:"RequireAuthz"`
	AuthnProvider AuthResultAuthnProvider `dependency:"AuthnProvider"`
	Validator     *validation.Validator   `dependency:"Validator"`
	SSOProvider   sso.Provider            `dependency:"SSOProvider"`
}

func (h *AuthResultHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.RequireClient)
}

type AuthResultPayload struct {
	AuthorizationCode string `json:"authorization_code"`
	CodeVerifier      string `json:"code_verifier"`
}

// @JSONSchema
const AuthResultRequestSchema = `
{
	"$id": "#AuthResultRequest",
	"type": "object",
	"properties": {
		"authorization_code": { "type": "string", "minLength": 1 },
		"code_verifier": { "type": "string", "minLength": 1 }
	},
	"required": ["authorization_code", "code_verifier"]
}
`

func (h *AuthResultHandler) DecodeRequest(w http.ResponseWriter, r *http.Request) (payload *AuthResultPayload, err error) {
	err = handler.BindJSONBody(r, w, h.Validator, "#AuthResultRequest", &payload)
	return
}

func (h *AuthResultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := h.DecodeRequest(w, r)
	if err != nil {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
		return
	}

	var result authn.Result
	err = db.WithTx(h.TxContext, func() (err error) {
		result, err = h.Handle(r, payload)
		return
	})
	if err != nil {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
		return
	}

	// TODO(authn): write response
	fmt.Printf("%#v\n", result)
}

func (h *AuthResultHandler) Handle(r *http.Request, payload *AuthResultPayload) (authn.Result, error) {
	code, err := h.SSOProvider.DecodeSkygearAuthorizationCode(payload.AuthorizationCode)
	if err != nil {
		return nil, err
	}

	err = h.SSOProvider.VerifyPKCE(code, payload.CodeVerifier)
	if err != nil {
		return nil, err
	}

	result, err := h.AuthnProvider.OAuthExchangeCode(
		coreauth.GetAccessKey(r.Context()).Client,
		nil, // TODO(authn): pass session
		code,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}
