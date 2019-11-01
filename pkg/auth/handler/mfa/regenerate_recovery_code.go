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
)

func AttachRegenerateRecoveryCodeHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/mfa/recovery_code/regenerate", &RegenerateRecoveryCodeHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type RegenerateRecoveryCodeHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f RegenerateRecoveryCodeHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &RegenerateRecoveryCodeHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(handler.APIHandlerToHandler(h, h.TxContext), h)
}

type RegenerateRecoveryCodeResponse struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

// @JSONSchema
const RegenerateRecoveryCodeResponseSchema = `
{
	"$id": "#RegenerateRecoveryCodeResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"recovery_codes": {
					"type": "array",
					"items": { "type": "string" }
				}
			}
		}
	}
}
`

/*
	@Operation POST /mfa/recovery_code/regenerate - Regenerate recovery codes
		Regenerate recovery codes. The old ones will no longer valid.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@Response 200
			List of newly generated recovery codes.
			@JSONSchema {RegenerateRecoveryCodeResponse}
*/
type RegenerateRecoveryCodeHandler struct {
	TxContext        db.TxContext            `dependency:"TxContext"`
	AuthContext      coreAuth.ContextGetter  `dependency:"AuthContextGetter"`
	RequireAuthz     handler.RequireAuthz    `dependency:"RequireAuthz"`
	MFAProvider      mfa.Provider            `dependency:"MFAProvider"`
	MFAConfiguration config.MFAConfiguration `dependency:"MFAConfiguration"`
}

func (h *RegenerateRecoveryCodeHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		policy.RequireValidUser,
	)
}

func (h *RegenerateRecoveryCodeHandler) WithTx() bool {
	return true
}

func (h *RegenerateRecoveryCodeHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (handler.RequestPayload, error) {
	payload := handler.EmptyRequestPayload{}
	err := handler.DecodeJSONBody(request, resp, &payload)
	return payload, err
}

func (h *RegenerateRecoveryCodeHandler) Handle(req interface{}) (resp interface{}, err error) {
	authInfo, _ := h.AuthContext.AuthInfo()
	userID := authInfo.ID
	codes, err := h.MFAProvider.GenerateRecoveryCode(userID)
	if err != nil {
		return nil, err
	}
	return RegenerateRecoveryCodeResponse{
		RecoveryCodes: codes,
	}, nil
}
