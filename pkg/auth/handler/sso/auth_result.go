package sso

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreauth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	coreauthn "github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachAuthResultHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/sso/auth_result").
		Handler(auth.MakeHandler(authDependency, newAuthResultHandler)).
		Methods("OPTIONS", "POST")
}

type AuthResultAuthnProvider interface {
	OAuthExchangeCode(
		client config.OAuthClientConfiguration,
		session model.SessionModeler,
		code *sso.SkygearAuthorizationCode,
	) (authn.Result, error)

	WriteResult(rw http.ResponseWriter, result authn.Result)
}

type AuthResultHandler struct {
	TxContext     db.TxContext
	AuthnProvider AuthResultAuthnProvider
	Validator     *validation.Validator
	SSOProvider   sso.Provider
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

	h.AuthnProvider.WriteResult(w, result)
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
		coreauthn.GetSession(r.Context()).(model.SessionModeler),
		code,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}
