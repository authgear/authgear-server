package mfa

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachDeleteAuthenticatorHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/mfa/authenticator/delete").
		Handler(server.FactoryToHandler(&DeleteAuthenticatorHandlerFactory{
			Dependency: authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type DeleteAuthenticatorHandlerFactory struct {
	Dependency pkg.DependencyMap
}

func (f DeleteAuthenticatorHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &DeleteAuthenticatorHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

type DeleteAuthenticatorRequest struct {
	AuthenticatorID string `json:"authenticator_id"`
}

// @JSONSchema
const DeleteAuthenticatorRequestSchema = `
{
	"$id": "#DeleteAuthenticatorRequest",
	"type": "object",
	"properties": {
		"authenticator_id": { "type": "string", "minLength": 1 }
	},
	"required": ["authenticator_id"]
}
`

/*
	@Operation POST /mfa/authenticator/delete - Delete authenticator.
		Delete authenticator.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			@JSONSchema {DeleteAuthenticatorRequest}
		@Response 200 {EmptyResponse}
*/
type DeleteAuthenticatorHandler struct {
	TxContext    db.TxContext          `dependency:"TxContext"`
	Validator    *validation.Validator `dependency:"Validator"`
	RequireAuthz handler.RequireAuthz  `dependency:"RequireAuthz"`
	MFAProvider  mfa.Provider          `dependency:"MFAProvider"`
}

func (h *DeleteAuthenticatorHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.RequireValidUser
}

func (h *DeleteAuthenticatorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h *DeleteAuthenticatorHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload DeleteAuthenticatorRequest
	if err := handler.BindJSONBody(r, w, h.Validator, "#DeleteAuthenticatorRequest", &payload); err != nil {
		return nil, err
	}

	err = db.WithTx(h.TxContext, func() error {
		userID := auth.GetSession(r.Context()).AuthnAttrs().UserID
		return h.MFAProvider.DeleteAuthenticator(userID, payload.AuthenticatorID)
	})
	resp = struct{}{}
	return
}
