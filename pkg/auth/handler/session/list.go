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
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

func AttachListHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/session/list").
		Handler(pkg.MakeHandler(authDependency, newListHandler)).
		Methods("OPTIONS", "POST")
}

type sessionListManager interface {
	List(userID string) ([]auth.AuthSession, error)
}

type ListResponse struct {
	Sessions []model.Session `json:"sessions"`
}

// @JSONSchema
const ListResponseSchema = `
{
	"$id": "#SessionListResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"sessions": {
					"type": "array",
					"items": { "$ref": "#Session" }
				}
			}
		}
	}
}
`

/*
	@Operation POST /session/list - Get current user sessions
		List all sessions of current user, in ascending order of creation time.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@Response 200
			List of all sessions.
			@JSONSchema {SessionListResponse}
*/
type ListHandler struct {
	txContext      db.TxContext
	sessionManager sessionListManager
}

func (h *ListHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireValidUser
}

func (h *ListHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if err := handler.DecodeJSONBody(r, rw, &struct{}{}); err != nil {
		handler.WriteResponse(rw, handler.APIResponse{Error: err})
		return
	}

	var result ListResponse
	err := db.WithTx(h.txContext, func() error {
		userID := auth.GetSession(r.Context()).AuthnAttrs().UserID
		sessions, err := h.sessionManager.List(userID)
		if err != nil {
			return err
		}

		result.Sessions = make([]model.Session, len(sessions))
		for i, s := range sessions {
			result.Sessions[i] = *s.ToAPIModel()
		}
		return nil
	})

	if err == nil {
		handler.WriteResponse(rw, handler.APIResponse{Result: result})
	} else {
		handler.WriteResponse(rw, handler.APIResponse{Error: err})
	}
}
