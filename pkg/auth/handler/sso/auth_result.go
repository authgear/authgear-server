package sso

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	interactionflows "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachAuthResultHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/sso/auth_result").
		Handler(pkg.MakeHandler(authDependency, newAuthResultHandler)).
		Methods("OPTIONS", "POST")
}

type OAuthResultInteractionFlow interface {
	ExchangeCode(codeHash string, verifier string) (*interactionflows.AuthResult, error)
}

type AuthResultHandler struct {
	TxContext    db.TxContext
	Validator    *validation.Validator
	SSOProvider  sso.Provider
	Interactions OAuthResultInteractionFlow
}

func (h *AuthResultHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireClient
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

	var result *interactionflows.AuthResult
	err = db.WithTx(h.TxContext, func() (err error) {
		result, err = h.Interactions.ExchangeCode(
			payload.AuthorizationCode,
			payload.CodeVerifier,
		)
		return
	})
	if err != nil {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
		return
	}
	result.WriteResponse(w)
}
