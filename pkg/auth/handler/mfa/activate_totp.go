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

func AttachActivateTOTPHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/mfa/totp/activate", &ActivateTOTPHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type ActivateTOTPHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f ActivateTOTPHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ActivateTOTPHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.RequireAuthz(handler.APIHandlerToHandler(h, h.TxContext), h.AuthContext, h)
}

type ActivateTOTPRequest struct {
	AuthenticatorID string `json:"authenticator_id"`
	OTP             string `json:"otp"`
}

func (r ActivateTOTPRequest) Validate() error {
	if r.AuthenticatorID == "" {
		return skyerr.NewInvalidArgument("missing authenticator ID", []string{"authenticator_id"})
	}
	if r.OTP == "" {
		return skyerr.NewInvalidArgument("missing OTP", []string{"otp"})
	}
	return nil
}

type ActivateTOTPResponse struct {
	RecoveryCodes []string `json:"recovery_codes,omitempty"`
}

// @JSONSchema
const ActivateTOTPRequestSchema = `
{
	"$id": "#ActivateTOTPRequest",
	"type": "object",
	"properties": {
		"authenticator_id": { "type": "string" },
		"otp": { "type": "string" }
	}
	"required": ["authenticator_id", "otp"]
}
`

// @JSONSchema
const ActivateTOTPResponseSchema = `
{
	"$id": "#ActivateTOTPResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"recovery_codes": {
					"type": "array",
					"items": {
						"type": "string"
					}
				}
			}
		}
	}
}
`

/*
	@Operation POST /mfa/totp/activate - Create TOTP authenticator.
		Create TOTP authenticator. It must be activated.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody {ActivateTOTPRequest}
		@Response 200
			Details of the authenticator
			@JSONSchema {ActivateTOTPResponse}
*/
type ActivateTOTPHandler struct {
	TxContext        db.TxContext            `dependency:"TxContext"`
	AuthContext      coreAuth.ContextGetter  `dependency:"AuthContextGetter"`
	MFAProvider      mfa.Provider            `dependency:"MFAProvider"`
	MFAConfiguration config.MFAConfiguration `dependency:"MFAConfiguration"`
}

func (h *ActivateTOTPHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

func (h *ActivateTOTPHandler) WithTx() bool {
	return true
}

func (h *ActivateTOTPHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := ActivateTOTPRequest{}
	err := handler.DecodeJSONBody(request, &payload)
	return payload, err
}

func (h *ActivateTOTPHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(ActivateTOTPRequest)
	userID := h.AuthContext.AuthInfo().ID
	recoveryCodes, err := h.MFAProvider.ActivateTOTP(userID, payload.AuthenticatorID, payload.OTP)
	if err != nil {
		return
	}
	resp = ActivateTOTPResponse{
		RecoveryCodes: recoveryCodes,
	}
	return
}
