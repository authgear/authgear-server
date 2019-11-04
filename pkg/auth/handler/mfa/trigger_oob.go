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
	return h.RequireAuthz(h, h)
}

type TriggerOOBRequest struct {
	AuthenticatorID   string `json:"authenticator_id"`
	AuthnSessionToken string `json:"authn_session_token"`
}

// @JSONSchema
const TriggerOOBRequestSchema = `
{
	"$id": "#TriggerOOBRequest",
	"type": "object",
	"properties": {
		"authenticator_id": { "type": "string", "minLength": 1 },
		"authn_session_token": { "type": "string", "minLength": 1 }
	}
}
`

/*
	@Operation POST /mfa/oob/trigger - Trigger OOB authenticator.
		Trigger OOB authenticator.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			@JSONSchema {TriggerOOBRequest}
		@Response 200 {EmptyResponse}
*/
type TriggerOOBHandler struct {
	TxContext            db.TxContext            `dependency:"TxContext"`
	Validator            *validation.Validator   `dependency:"Validator"`
	AuthContext          coreAuth.ContextGetter  `dependency:"AuthContextGetter"`
	RequireAuthz         handler.RequireAuthz    `dependency:"RequireAuthz"`
	MFAProvider          mfa.Provider            `dependency:"MFAProvider"`
	MFAConfiguration     config.MFAConfiguration `dependency:"MFAConfiguration"`
	AuthnSessionProvider authnsession.Provider   `dependency:"AuthnSessionProvider"`
}

func (h *TriggerOOBHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.DenyInvalidSession),
	)
}

func (h *TriggerOOBHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h *TriggerOOBHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload TriggerOOBRequest
	if err := handler.BindJSONBody(r, w, h.Validator, "#TriggerOOBRequest", &payload); err != nil {
		return nil, err
	}
	err = db.WithTx(h.TxContext, func() error {
		userID, _, _, err := h.AuthnSessionProvider.Resolve(h.AuthContext, payload.AuthnSessionToken, authnsession.ResolveOptions{
			MFAOption: authnsession.ResolveMFAOptionAlwaysAccept,
		})
		if err != nil {
			return err
		}
		err = h.MFAProvider.TriggerOOB(userID, payload.AuthenticatorID)
		if err != nil {
			return err
		}
		resp = struct{}{}
		return nil
	})
	return
}
