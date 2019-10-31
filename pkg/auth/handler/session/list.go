package session

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	authSession "github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachListHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/session/list", &ListHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type ListHandlerFactory struct {
	Dependency auth.DependencyMap
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
	AuthContext     coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	RequireAuthz    handler.RequireAuthz   `dependency:"RequireAuthz"`
	TxContext       db.TxContext           `dependency:"TxContext"`
	SessionProvider session.Provider       `dependency:"SessionProvider"`
}

func (h ListHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		policy.RequireValidUser,
	)
}

func (h ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	var payload struct{}
	if err := handler.DecodeJSONBody(r, w, &payload); err != nil {
		response.Error = err
	} else {
		result, err := h.Handle()
		if err != nil {
			response.Error = err
		} else {
			response.Result = result
		}
	}
	handler.WriteResponse(w, response)
}

func (h ListHandler) Handle() (resp interface{}, err error) {
	err = db.WithTx(h.TxContext, func() error {
		authInfo, _ := h.AuthContext.AuthInfo()
		userID := authInfo.ID

		sessions, err := h.SessionProvider.List(userID)
		if err != nil {
			return err
		}

		sessionModels := make([]model.Session, len(sessions))
		for i, session := range sessions {
			sessionModels[i] = authSession.Format(session)
		}

		resp = ListResponse{Sessions: sessionModels}
		return nil
	})
	return
}
