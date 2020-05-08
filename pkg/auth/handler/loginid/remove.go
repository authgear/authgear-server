package loginid

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachRemoveLoginIDHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/login_id/remove").
		Handler(pkg.MakeHandler(authDependency, newRemoveLoginIDHandler)).
		Methods("OPTIONS", "POST")
}

type RemoveLoginIDRequestPayload struct {
	loginid.LoginID
}

// @JSONSchema
const RemoveLoginIDRequestSchema = `
{
	"$id": "#RemoveLoginIDRequest",
	"type": "object",
	"properties": {
		"key": { "type": "string", "minLength": 1 },
		"value": { "type": "string", "minLength": 1 }
	},
	"required": ["key", "value"]
}
`

type RemoveLoginIDInteractionFlow interface {
	RemoveLoginID(
		loginIDKey string, loginID string, session auth.AuthSession,
	) error
}

/*
	@Operation POST /login_id/remove - Remove login ID
		Remove login ID from current user.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			Describe the login ID to remove.
			@JSONSchema {RemoveLoginIDRequest}

		@Response 200 {EmptyResponse}

		@Callback identity_delete {UserSyncEvent}
		@Callback user_sync {UserSyncEvent}
*/
type RemoveLoginIDHandler struct {
	Validator    *validation.Validator
	TxContext    db.TxContext
	Interactions RemoveLoginIDInteractionFlow
}

func (h RemoveLoginIDHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireValidUser
}

func (h RemoveLoginIDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.Handle(w, r)
	if err == nil {
		handler.WriteResponse(w, handler.APIResponse{Result: struct{}{}})
	} else {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
	}
}

func (h RemoveLoginIDHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	var payload RemoveLoginIDRequestPayload
	if err := handler.BindJSONBody(r, w, h.Validator, "#RemoveLoginIDRequest", &payload); err != nil {
		return err
	}

	err := db.WithTx(h.TxContext, func() error {
		session := auth.GetSession(r.Context())

		err := h.Interactions.RemoveLoginID(payload.Key, payload.Value, session)
		if err != nil {
			return err
		}

		// TODO(interaction): remove identity hook

		// TODO(interaction): update verification state

		return nil
	})
	return err
}
