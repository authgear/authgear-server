package mfa

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	coreauthn "github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachCreateOOBHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/mfa/oob/new").
		Handler(pkg.MakeHandler(authDependency, newCreateOOBHandler)).
		Methods("OPTIONS", "POST")
}

type CreateOOBRequest struct {
	Channel           coreauthn.AuthenticatorOOBChannel `json:"channel"`
	Phone             string                            `json:"phone"`
	Email             string                            `json:"email"`
	AuthnSessionToken string                            `json:"authn_session_token"`
}

type CreateOOBResponse struct {
	AuthenticatorID   string `json:"authenticator_id"`
	AuthenticatorType string `json:"authenticator_type"`
	Channel           string `json:"channel"`
}

// @JSONSchema
const CreateOOBRequestSchema = `
{
	"$id": "#CreateOOBRequest",
	"oneOf": [
		{
			"additionalProperties": false,
			"properties": {
				"channel": { "enum": ["sms"] },
				"phone": { "type": "string", "format": "phone" },
				"authn_session_token": { "type": "string" }
			},
			"required": ["channel", "phone"]
		},
		{
			"additionalProperties": false,
			"properties": {
				"channel": { "enum": ["email"] },
				"email": { "type": "string", "format": "email" },
				"authn_session_token": { "type": "string", "minLength": 1 }
			},
			"required": ["channel", "email"]
		}
	]
}
`

// @JSONSchema
const CreateOOBResponseSchema = `
{
	"$id": "#CreateOOBResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"authenticator_id": { "type": "string" },
				"authenticator_type": { "type": "string" },
				"channel": { "type": "string" }
			}
		}
	}
}
`

/*
	@Operation POST /mfa/oob/new - Create OOB authenticator.
		Create inactive OOB authenticator. It must be activated later.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			@JSONSchema {CreateOOBRequest}
		@Response 200
			Details of the authenticator
			@JSONSchema {CreateOOBResponse}
*/
type CreateOOBHandler struct {
	TxContext     db.TxContext
	Validator     *validation.Validator
	MFAProvider   mfa.Provider
	authnResolver authnResolver
}

func (h *CreateOOBHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.RequireClient)
}

func (h *CreateOOBHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h *CreateOOBHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload CreateOOBRequest
	if err := handler.BindJSONBody(r, w, h.Validator, "#CreateOOBRequest", &payload); err != nil {
		return nil, err
	}

	err = db.WithTx(h.TxContext, func() error {
		var session coreauthn.Attributer = auth.GetSession(r.Context())
		if session == nil {
			session, err = h.authnResolver.Resolve(
				coreAuth.GetAccessKey(r.Context()).Client,
				payload.AuthnSessionToken,
				func(s authn.SessionStep) bool { return s == authn.SessionStepMFASetup },
			)
			if err != nil {
				return err
			}
		}

		a, err := h.MFAProvider.CreateOOB(session.AuthnAttrs().UserID, payload.Channel, payload.Phone, payload.Email)
		if err != nil {
			return err
		}

		resp = CreateOOBResponse{
			AuthenticatorID:   a.ID,
			AuthenticatorType: string(a.Type),
			Channel:           string(payload.Channel),
		}
		return nil
	})

	return
}
