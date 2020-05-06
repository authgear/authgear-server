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

func AttachAddLoginIDHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/login_id/add").
		Handler(pkg.MakeHandler(authDependency, newAddLoginIDHandler)).
		Methods("OPTIONS", "POST")
}

type AddLoginIDRequestPayload struct {
	LoginIDs []loginid.LoginID `json:"login_ids"`
}

// @JSONSchema
const AddLoginIDRequestSchema = `
{
	"$id": "#AddLoginIDRequest",
	"type": "object",
	"properties": {
		"login_ids": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
					"key": { "type": "string", "minLength": 1 },
					"value": { "type": "string", "minLength": 1 }
				},
				"required": ["key", "value"]
			},
			"minItems": 1
		}
	},
	"required": ["login_ids"]
}
`

type AddLoginIDInteractionFlow interface {
	AddLoginID(
		loginIDKey string, loginID string, session auth.AuthSession,
	) error
}

/*
	@Operation POST /login_id/add - Add login ID
		Add new login ID for current user.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			Describe the new login ID.
			@JSONSchema {AddLoginIDRequest}

		@Response 200 {EmptyResponse}

		@Callback identity_create {UserSyncEvent}
		@Callback user_sync {UserSyncEvent}
*/
type AddLoginIDHandler struct {
	Validator    *validation.Validator
	TxContext    db.TxContext
	Interactions AddLoginIDInteractionFlow
}

func (h AddLoginIDHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireValidUser
}

func (h AddLoginIDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.Handle(w, r)
	if err == nil {
		handler.WriteResponse(w, handler.APIResponse{Result: struct{}{}})
	} else {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
	}
}

func (h AddLoginIDHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	var payload AddLoginIDRequestPayload
	if err := handler.BindJSONBody(r, w, h.Validator, "#AddLoginIDRequest", &payload); err != nil {
		return err
	}

	err := db.WithTx(h.TxContext, func() error {
		userSession := auth.GetSession(r.Context())

		for idx, loginID := range payload.LoginIDs {

			err := h.Interactions.AddLoginID(loginID.Key, loginID.Value, userSession)

			// TODO(interaction): hook identity create event

			if err != nil {
				err = correctErrorCausePointer(err, idx)
				return err
			}
		}

		// TODO(interaction): update verification state

		return nil
	})
	return err
}
