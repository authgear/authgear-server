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
	"github.com/skygeario/skygear-server/pkg/core/validation"
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
	return h.RequireAuthz(h, h)
}

type ActivateOOBRequest struct {
	Code              string `json:"code"`
	AuthnSessionToken string `json:"authn_session_token"`
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
		"code": { "type": "string", "minLength": 1 },
		"authn_session_token": { "type": "string", "minLength": 1 }
	},
	"required": ["code"]
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

		@RequestBody
			@JSONSchema {ActivateOOBRequest}
		@Response 200
			Details of the authenticator
			@JSONSchema {ActivateOOBResponse}
*/
type ActivateOOBHandler struct {
	TxContext            db.TxContext            `dependency:"TxContext"`
	Validator            *validation.Validator   `dependency:"Validator"`
	AuthContext          coreAuth.ContextGetter  `dependency:"AuthContextGetter"`
	RequireAuthz         handler.RequireAuthz    `dependency:"RequireAuthz"`
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

func (h *ActivateOOBHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h *ActivateOOBHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload ActivateOOBRequest
	if err := handler.BindJSONBody(r, w, h.Validator, "#ActivateOOBRequest", &payload); err != nil {
		return nil, err
	}

	err = db.WithTx(h.TxContext, func() error {
		userID, _, _, err := h.AuthnSessionProvider.Resolve(h.AuthContext, payload.AuthnSessionToken, authnsession.ResolveOptions{
			MFAOption: authnsession.ResolveMFAOptionOnlyWhenNoAuthenticators,
		})
		if err != nil {
			return err
		}
		recoveryCodes, err := h.MFAProvider.ActivateOOB(userID, payload.Code)
		if err != nil {
			return err
		}

		resp = ActivateOOBResponse{
			RecoveryCodes: recoveryCodes,
		}
		return nil
	})
	return
}
