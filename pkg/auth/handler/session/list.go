package session

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachListHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/session/list").
		Handler(server.FactoryToHandler(&ListHandlerFactory{
			authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type ListHandlerFactory struct {
	Dependency pkg.DependencyMap
}

func (f ListHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ListHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
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
	RequireAuthz    handler.RequireAuthz `dependency:"RequireAuthz"`
	TxContext       db.TxContext         `dependency:"TxContext"`
	SessionProvider session.Provider     `dependency:"SessionProvider"`
}

func (h ListHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.RequireValidUser
}

func (h ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	var payload struct{}
	if err := handler.DecodeJSONBody(r, w, &payload); err != nil {
		response.Error = err
	} else {
		result, err := h.Handle(r)
		if err != nil {
			response.Error = err
		} else {
			response.Result = result
		}
	}
	handler.WriteResponse(w, response)
}

func (h ListHandler) Handle(r *http.Request) (resp interface{}, err error) {
	err = db.WithTx(h.TxContext, func() error {
		authInfo := auth.GetUser(r.Context())
		userID := authInfo.ID

		sessions, err := h.SessionProvider.List(userID)
		if err != nil {
			return err
		}

		sessionModels := make([]model.Session, len(sessions))
		for i, session := range sessions {
			// TODO(authn): use new session
			// sessionModels[i] = authSession.Format(session)
			_, _ = i, session
		}

		resp = ListResponse{Sessions: sessionModels}
		return nil
	})
	return
}
