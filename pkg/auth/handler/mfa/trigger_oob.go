package mfa

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	coreauthn "github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachTriggerOOBHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/mfa/oob/trigger").
		Handler(pkg.MakeHandler(authDependency, newTriggerOOBHandler)).
		Methods("OPTIONS", "POST")
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
	TxContext     db.TxContext
	Validator     *validation.Validator
	MFAProvider   mfa.Provider
	authnResolver authnResolver
}

func (h *TriggerOOBHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireClient
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
		var session coreauthn.Attributer = auth.GetSession(r.Context())
		if session == nil {
			session, err = h.authnResolver.Resolve(
				coreAuth.GetAccessKey(r.Context()).Client,
				payload.AuthnSessionToken,
				authn.SessionStep.IsMFA,
			)
			if err != nil {
				return err
			}
		}

		err = h.MFAProvider.TriggerOOB(session.AuthnAttrs().UserID, payload.AuthenticatorID)
		if err != nil {
			return err
		}
		resp = struct{}{}
		return nil
	})
	return
}
