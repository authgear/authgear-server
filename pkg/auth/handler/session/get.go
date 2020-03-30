package session

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachGetHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/session/get").
		Handler(pkg.MakeHandler(authDependency, newGetHandler)).
		Methods("OPTIONS", "POST")
}

type GetRequestPayload struct {
	SessionID string `json:"session_id"`
}

// @JSONSchema
const GetRequestSchema = `
{
	"$id": "#SessionGetRequest",
	"type": "object",
	"properties": {
		"session_id": { "type": "string", "minLength": 1 }
	},
	"required": ["session_id"]
}
`

type GetResponse struct {
	Session model.Session `json:"session"`
}

// @JSONSchema
const GetResponseSchema = `
{
	"$id": "#SessionGetResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"session": { "$ref": "#Session" }
			}
		}
	}
}
`

type sessionGetManager interface {
	Get(id string) (auth.AuthSession, error)
}

/*
	@Operation POST /session/get - Get current user sessions
		Get the sessions with specified ID of current user.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			Describe the session ID.
			@JSONSchema {SessionGetRequest}

		@Response 200
			The requested session.
			@JSONSchema {SessionGetResponse}
*/
type GetHandler struct {
	validator      *validation.Validator
	txContext      db.TxContext
	sessionManager sessionGetManager
}

func (h *GetHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireValidUser
}

func (h *GetHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var payload GetRequestPayload
	if err := handler.BindJSONBody(r, rw, h.validator, "#SessionGetRequest", &payload); err != nil {
		handler.WriteResponse(rw, handler.APIResponse{Error: err})
		return
	}

	var result GetResponse
	err := db.WithTx(h.txContext, func() error {
		userID := auth.GetSession(r.Context()).AuthnAttrs().UserID

		session, err := h.sessionManager.Get(payload.SessionID)
		if errors.Is(err, auth.ErrSessionNotFound) {
			return errSessionNotFound
		} else if err != nil {
			return err
		}

		if session.AuthnAttrs().UserID != userID {
			return errSessionNotFound
		}

		result.Session = *session.ToAPIModel()
		return nil
	})

	if err == nil {
		handler.WriteResponse(rw, handler.APIResponse{Result: result})
	} else {
		handler.WriteResponse(rw, handler.APIResponse{Error: err})
	}
}
