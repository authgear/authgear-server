package loginid

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	interactionflows "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachUpdateLoginIDHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/login_id/update").
		Handler(pkg.MakeHandler(authDependency, newUpdateLoginIDHandler)).
		Methods("OPTIONS", "POST")
}

type UpdateLoginIDRequestPayload struct {
	OldLoginID loginid.LoginID `json:"old_login_id"`
	NewLoginID loginid.LoginID `json:"new_login_id"`
}

// @JSONSchema
const UpdateLoginIDRequestSchema = `
{
	"$id": "#UpdateLoginIDRequest",
	"type": "object",
	"properties": {
		"old_login_id": {
			"type": "object",
			"properties": {
				"key": { "type": "string", "minLength": 1 },
				"value": { "type": "string", "minLength": 1 }
			},
			"required": ["key", "value"]
		},
		"new_login_id": {
			"type": "object",
			"properties": {
				"key": { "type": "string", "minLength": 1 },
				"value": { "type": "string", "minLength": 1 }
			},
			"required": ["key", "value"]
		}
	},
	"required": ["old_login_id", "new_login_id"]
}
`

type UpdateLoginIDInteractionFlow interface {
	UpdateLoginID(
		oldLoginID loginid.LoginID, newLoginID loginid.LoginID, session auth.AuthSession,
	) (*interactionflows.AuthResult, error)
}

/*
	@Operation POST /login_id/update - update login ID
		Update the specified login ID for current user.
		This operation is same as adding the new login ID and then deleting
		old login ID atomically.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			Describe the new login ID.
			@JSONSchema {UpdateLoginIDRequest}

		@Response 200
			Updated user and identity info.
			@JSONSchema {UserResponse}

		@Callback identity_create {UserSyncEvent}
		@Callback identity_delete {UserSyncEvent}
		@Callback user_sync {UserSyncEvent}
*/
type UpdateLoginIDHandler struct {
	Validator    *validation.Validator
	TxContext    db.TxContext
	Interactions UpdateLoginIDInteractionFlow
}

func (h UpdateLoginIDHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireValidUser
}

func (h UpdateLoginIDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := h.Handle(w, r)
	if err != nil {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
		return
	}

	result.WriteResponse(w)
}

func (h UpdateLoginIDHandler) Handle(w http.ResponseWriter, r *http.Request) (*interactionflows.AuthResult, error) {
	var payload UpdateLoginIDRequestPayload
	if err := handler.BindJSONBody(r, w, h.Validator, "#UpdateLoginIDRequest", &payload); err != nil {
		return nil, err
	}

	var result *interactionflows.AuthResult
	err := db.WithTx(h.TxContext, func() error {
		var err error
		session := auth.GetSession(r.Context())

		result, err = h.Interactions.UpdateLoginID(payload.OldLoginID, payload.NewLoginID, session)
		if err != nil {
			err = correctErrorCausePointer(err, func(relativePath string) string {
				return fmt.Sprintf("/new_login_id/%s", relativePath)
			})
			return err
		}

		// TODO(interaction): update verification state

		return nil
	})
	return result, err
}
