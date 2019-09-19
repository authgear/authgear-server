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
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func AttachActivateOOBHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/mfa/oob/activate", &ActivateOOBHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type ActivateOOBHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f ActivateOOBHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ActivateOOBHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.RequireAuthz(handler.APIHandlerToHandler(h, h.TxContext), h.AuthContext, h)
}

type ActivateOOBRequest struct {
	AuthenticatorID   string `json:"authenticator_id"`
	Code              string `json:"code"`
	AuthnSessionToken string `json:"authn_session_token"`
}

func (r ActivateOOBRequest) Validate() error {
	if r.AuthenticatorID == "" {
		return skyerr.NewInvalidArgument("missing authenticator ID", []string{"authenticator_id"})
	}
	if r.Code == "" {
		return skyerr.NewInvalidArgument("missing code", []string{"code"})
	}
	return nil
}

type ActivateOOBResponse struct {
	RecoveryCodes []string `json:"recovery_codes,omitempty"`
}

// @JSONSchema
const ActivateOOBRequestSchema = `
{
	"$id": "#ActivateOOBRequest",
	"type": "object",
	"properties": {
		"authenticator_id": { "type": "string" },
		"code": { "type": "string" },
		"authn_session_token": { "type": "string" }
	},
	"required": ["authenticator_id", "code"]
}
`

// @JSONSchema
const ActivateOOBResponseSchema = `
{
	"$id": "#ActivateOOBResponse",
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
	@Operation POST /mfa/oob/activate - Activate OOB authenticator.
		Activate OOB authenticator.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody {ActivateOOBRequest}
		@Response 200
			Details of the authenticator
			@JSONSchema {ActivateOOBResponse}
*/
type ActivateOOBHandler struct {
	TxContext            db.TxContext            `dependency:"TxContext"`
	AuthContext          coreAuth.ContextGetter  `dependency:"AuthContextGetter"`
	MFAProvider          mfa.Provider            `dependency:"MFAProvider"`
	MFAConfiguration     config.MFAConfiguration `dependency:"MFAConfiguration"`
	AuthnSessionProvider authnsession.Provider   `dependency:"AuthnSessionProvider"`
}

func (h *ActivateOOBHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.DenyInvalidSession),
	)
}

func (h *ActivateOOBHandler) WithTx() bool {
	return true
}

func (h *ActivateOOBHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := ActivateOOBRequest{}
	err := handler.DecodeJSONBody(request, &payload)
	return payload, err
}

func (h *ActivateOOBHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(ActivateOOBRequest)
	userID, _, _, err := h.AuthnSessionProvider.Resolve(h.AuthContext, payload.AuthnSessionToken, authnsession.ResolveOptions{
		MFAOption: authnsession.ResolveMFAOptionOnlyWhenNoAuthenticators,
	})
	if err != nil {
		return nil, err
	}
	recoveryCodes, err := h.MFAProvider.ActivateOOB(userID, payload.AuthenticatorID, payload.Code)
	if err != nil {
		return
	}
	resp = ActivateOOBResponse{
		RecoveryCodes: recoveryCodes,
	}
	return
}
