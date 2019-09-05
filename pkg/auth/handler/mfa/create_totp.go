package mfa

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func AttachCreateTOTPHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/mfa/totp/new", &CreateTOTPHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type CreateTOTPHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f CreateTOTPHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &CreateTOTPHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.RequireAuthz(handler.APIHandlerToHandler(h, h.TxContext), h.AuthContext, h)
}

func (h *CreateTOTPHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type CreateTOTPRequest struct {
	DisplayName string `json:"display_name"`
}

func (r CreateTOTPRequest) Validate() error {
	if r.DisplayName == "" {
		return skyerr.NewInvalidArgument("missing display name", []string{"display_name"})
	}
	return nil
}

type CreateTOTPResponse struct {
	AuthenticatorID   string `json:"authenticator_id"`
	AuthenticatorType string `json:"authenticator_type"`
	Secret            string `json:"secret"`
}

// @JSONSchema
const CreateTOTPRequestSchema = `
{
	"$id": "#CreateTOTPRequest",
	"type": "object",
	"properties": {
		"display_name": { "type": string },
	}
	"required": ["display_name"]
}
`

// @JSONSchema
const CreateTOTPResponseSchema = `
{
	"$id": "#CreateTOTPResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"authenticator_id": { "type": "string" },
				"authenticator_type": { "type": "string" },
				"secret": { "type": "string" }
			}
		}
	}
}
`

/*
	@Operation POST /mfa/totp/new - Create TOTP authenticator.
		Create TOTP authenticator. It must be activated.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody {CreateTOTPRequest}
		@Response 200
			Details of the authenticator
			@JSONSchema {CreateTOTPResponse}
*/
type CreateTOTPHandler struct {
	TxContext        db.TxContext            `dependency:"TxContext"`
	AuthContext      coreAuth.ContextGetter  `dependency:"AuthContextGetter"`
	MFAProvider      mfa.Provider            `dependency:"MFAProvider"`
	MFAConfiguration config.MFAConfiguration `dependency:"MFAConfiguration"`
}

func (h *CreateTOTPHandler) WithTx() bool {
	return true
}

func (h *CreateTOTPHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := CreateTOTPRequest{}
	err := handler.DecodeJSONBody(request, &payload)
	return payload, err
}

func (h *CreateTOTPHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(CreateTOTPRequest)
	userID := h.AuthContext.AuthInfo().ID
	a, err := h.MFAProvider.CreateTOTP(userID, payload.DisplayName)
	if err != nil {
		return
	}
	resp = CreateTOTPResponse{
		AuthenticatorID:   a.ID,
		AuthenticatorType: string(a.Type),
		Secret:            a.Secret,
	}
	return
}
