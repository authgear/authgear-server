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
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f ListHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
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
	TxContext       db.TxContext           `dependency:"TxContext"`
	SessionProvider session.Provider       `dependency:"SessionProvider"`
}

func (h ListHandler) WithTx() bool {
	return true
}

func (h ListHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := handler.EmptyRequestPayload{}
	err := handler.DecodeJSONBody(request, &payload)
	return payload, err
}

func (h ListHandler) Handle(req interface{}) (resp interface{}, err error) {
	userID := h.AuthContext.AuthInfo().ID

	sessions, err := h.SessionProvider.List(userID)
	if err != nil {
		return
	}

	sessionModels := make([]model.Session, len(sessions))
	for i, session := range sessions {
		sessionModels[i] = authSession.Format(session)
	}

	resp = ListResponse{Sessions: sessionModels}
	return
}
