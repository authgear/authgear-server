package session

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachRevokeAllHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/session/revoke_all", &RevokeAllHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type RevokeAllHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f RevokeAllHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &RevokeAllHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f RevokeAllHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

/*
	@Operation POST /session/revoke_all - Revoke all sessions
		Revoke all sessions, excluding current session.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@Response 200 {EmptyResponse}
*/
type RevokeAllHandler struct {
	AuthContext     coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	TxContext       db.TxContext           `dependency:"TxContext"`
	SessionProvider session.Provider       `dependency:"SessionProvider"`
}

func (h RevokeAllHandler) WithTx() bool {
	return true
}

func (h RevokeAllHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	return handler.EmptyRequestPayload{}, nil
}

func (h RevokeAllHandler) Handle(req interface{}) (resp interface{}, err error) {
	userID := h.AuthContext.AuthInfo().ID
	sessionID := h.AuthContext.Session().ID

	err = h.SessionProvider.InvalidateAll(userID, sessionID)
	if err != nil {
		return
	}

	resp = map[string]string{}
	return
}
