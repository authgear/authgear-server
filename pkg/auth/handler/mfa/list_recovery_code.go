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

func AttachListRecoveryCodeHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/mfa/recovery_code/list", &ListRecoveryCodeHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type ListRecoveryCodeHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f ListRecoveryCodeHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ListRecoveryCodeHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.RequireAuthz(handler.APIHandlerToHandler(h, h.TxContext), h.AuthContext, h)
}

type ListRecoveryCodeResponse struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

// @JSONSchema
const ListRecoveryCodeResponseSchema = `
{
	"$id": "#ListRecoveryCodeResponse",
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
	@Operation POST /mfa/recovery_code/list - List recovery codes
		List recovery codes if allowed.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@Response 200
			List of recovery codes.
			@JSONSchema {ListRecoveryCodeResponse}
*/
type ListRecoveryCodeHandler struct {
	TxContext        db.TxContext            `dependency:"TxContext"`
	AuthContext      coreAuth.ContextGetter  `dependency:"AuthContextGetter"`
	MFAProvider      mfa.Provider            `dependency:"MFAProvider"`
	MFAConfiguration config.MFAConfiguration `dependency:"MFAConfiguration"`
}

func (h *ListRecoveryCodeHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

func (h *ListRecoveryCodeHandler) WithTx() bool {
	return true
}

func (h *ListRecoveryCodeHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := handler.EmptyRequestPayload{}
	err := handler.DecodeJSONBody(request, &payload)
	return payload, err
}

func (h *ListRecoveryCodeHandler) Handle(req interface{}) (resp interface{}, err error) {
	if !h.MFAConfiguration.RecoveryCode.ListEnabled {
		return nil, skyerr.NewError(skyerr.UndefinedOperation, "listing recovery code is disabled")
	}
	authInfo, _ := h.AuthContext.AuthInfo()
	userID := authInfo.ID
	codes, err := h.MFAProvider.GetRecoveryCode(userID)
	if err != nil {
		return nil, err
	}
	return ListRecoveryCodeResponse{
		RecoveryCodes: codes,
	}, nil
}
