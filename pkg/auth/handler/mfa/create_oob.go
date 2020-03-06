package mfa

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachCreateOOBHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/mfa/oob/new").
		Handler(server.FactoryToHandler(&CreateOOBHandlerFactory{
			Dependency: authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type CreateOOBHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f CreateOOBHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &CreateOOBHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

func (h *CreateOOBHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.DenyInvalidSession),
	)
}

type CreateOOBRequest struct {
	Channel           coreAuth.AuthenticatorOOBChannel `json:"channel"`
	Phone             string                           `json:"phone"`
	Email             string                           `json:"email"`
	AuthnSessionToken string                           `json:"authn_session_token"`
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
	TxContext            db.TxContext            `dependency:"TxContext"`
	Validator            *validation.Validator   `dependency:"Validator"`
	AuthContext          coreAuth.ContextGetter  `dependency:"AuthContextGetter"`
	RequireAuthz         handler.RequireAuthz    `dependency:"RequireAuthz"`
	MFAProvider          mfa.Provider            `dependency:"MFAProvider"`
	MFAConfiguration     config.MFAConfiguration `dependency:"MFAConfiguration"`
	AuthnSessionProvider authnsession.Provider   `dependency:"AuthnSessionProvider"`
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
		userID, _, _, err := h.AuthnSessionProvider.Resolve(h.AuthContext, payload.AuthnSessionToken, authnsession.ResolveOptions{
			MFAOption: authnsession.ResolveMFAOptionOnlyWhenNoAuthenticators,
		})
		if err != nil {
			return err
		}
		a, err := h.MFAProvider.CreateOOB(userID, payload.Channel, payload.Phone, payload.Email)
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
