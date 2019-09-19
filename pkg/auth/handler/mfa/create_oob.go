package mfa

import (
	"net/http"

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
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/phone"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func AttachCreateOOBHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/mfa/oob/new", &CreateOOBHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type CreateOOBHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f CreateOOBHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &CreateOOBHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.RequireAuthz(handler.APIHandlerToHandler(h, h.TxContext), h.AuthContext, h)
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

func (r CreateOOBRequest) Validate() error {
	switch r.Channel {
	case coreAuth.AuthenticatorOOBChannelSMS:
		return phone.EnsureE164(r.Phone)
	case coreAuth.AuthenticatorOOBChannelEmail:
		return mail.EnsureAddressOnly(r.Email)
	default:
		return skyerr.NewInvalidArgument("invalid channel", []string{"channel"})
	}
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
			"properties": {
				"channel": { "const": "sms" },
				"phone": { "type": "string" },
				"authn_session_token": { "type": "string" }
			},
			"required": ["channel", "phone"]
		},
		{
			"properties": {
				"channel": { "const": "email" },
				"email": { "type": "string" },
				"authn_session_token": { "type": "string" }
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

		@RequestBody {CreateOOBRequest}
		@Response 200
			Details of the authenticator
			@JSONSchema {CreateOOBResponse}
*/
type CreateOOBHandler struct {
	TxContext            db.TxContext            `dependency:"TxContext"`
	AuthContext          coreAuth.ContextGetter  `dependency:"AuthContextGetter"`
	MFAProvider          mfa.Provider            `dependency:"MFAProvider"`
	MFAConfiguration     config.MFAConfiguration `dependency:"MFAConfiguration"`
	AuthnSessionProvider authnsession.Provider   `dependency:"AuthnSessionProvider"`
}

func (h *CreateOOBHandler) WithTx() bool {
	return true
}

func (h *CreateOOBHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := CreateOOBRequest{}
	err := handler.DecodeJSONBody(request, &payload)
	return payload, err
}

func (h *CreateOOBHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(CreateOOBRequest)
	userID, _, _, err := h.AuthnSessionProvider.Resolve(h.AuthContext, payload.AuthnSessionToken, authnsession.ResolveOptions{
		MFAOption: authnsession.ResolveMFAOptionOnlyWhenNoAuthenticators,
	})
	if err != nil {
		return nil, err
	}
	a, err := h.MFAProvider.CreateOOB(userID, payload.Channel, payload.Phone, payload.Email)
	if err != nil {
		return
	}
	resp = CreateOOBResponse{
		AuthenticatorID:   a.ID,
		AuthenticatorType: string(a.Type),
		Channel:           string(payload.Channel),
	}
	return
}
