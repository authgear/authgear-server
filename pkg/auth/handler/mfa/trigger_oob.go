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

func AttachTriggerOOBHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/mfa/oob/trigger", &TriggerOOBHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type TriggerOOBHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f TriggerOOBHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &TriggerOOBHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.RequireAuthz(handler.APIHandlerToHandler(h, h.TxContext), h.AuthContext, h)
}

type TriggerOOBRequest struct {
	AuthenticatorID string `json:"authenticator_id"`
}

func (r TriggerOOBRequest) Validate() error {
	return nil
}

// @JSONSchema
const TriggerOOBRequestSchema = `
{
	"$id": "#TriggerOOBRequest",
	"type": "object",
	"properties": {
		"authenticator_id": { "type": "string" }
	}
}
`

/*
	@Operation POST /mfa/oob/trigger - Trigger OOB authenticator.
		Trigger OOB authenticator.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody {TriggerOOBRequest}
		@Response 200 {EmptyResponse}
*/
type TriggerOOBHandler struct {
	TxContext        db.TxContext            `dependency:"TxContext"`
	AuthContext      coreAuth.ContextGetter  `dependency:"AuthContextGetter"`
	MFAProvider      mfa.Provider            `dependency:"MFAProvider"`
	MFAConfiguration config.MFAConfiguration `dependency:"MFAConfiguration"`
}

func (h *TriggerOOBHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

func (h *TriggerOOBHandler) WithTx() bool {
	return true
}

func (h *TriggerOOBHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := TriggerOOBRequest{}
	err := handler.DecodeJSONBody(request, &payload)
	return payload, err
}

func (h *TriggerOOBHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(TriggerOOBRequest)
	userID := h.AuthContext.AuthInfo().ID
	err = h.MFAProvider.TriggerOOB(userID, payload.AuthenticatorID)
	if err != nil {
		return
	}
	resp = map[string]interface{}{}
	return
}
