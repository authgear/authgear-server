package session

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"

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

func AttachRevokeHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/session/revoke", &RevokeHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type RevokeHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f RevokeHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &RevokeHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f RevokeHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type RevokeRequestPayload struct {
	SessionID string `json:"session_id"`
}

// @JSONSchema
const RevokeRequestSchema = `
{
	"$id": "#SessionRevokeRequest",
	"type": "object",
	"properties": {
		"session_id": { "type": "string" }
	}
}
`

func (p RevokeRequestPayload) Validate() error {
	return nil
}

/*
	@Operation POST /session/revoke - Revoke session
		Update specified session. Current session cannot be revoked.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			Describe the session ID.
			@JSONSchema {SessionRevokeRequest}

		@Response 200 {EmptyResponse}
*/
type RevokeHandler struct {
	AuthContext     coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	TxContext       db.TxContext           `dependency:"TxContext"`
	SessionProvider session.Provider       `dependency:"SessionProvider"`
}

func (h RevokeHandler) WithTx() bool {
	return true
}

func (h RevokeHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := RevokeRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	return payload, err
}

func (h RevokeHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(RevokeRequestPayload)

	userID := h.AuthContext.AuthInfo().ID
	sessionID := payload.SessionID

	if h.AuthContext.Session().ID == sessionID {
		err = skyerr.NewInvalidArgument("must not revoke current session", []string{"session_id"})
		return
	}

	// ignore session not found errors
	s, err := h.SessionProvider.Get(sessionID)
	if err != nil {
		if err == session.ErrSessionNotFound {
			err = nil
		}
		return
	}
	if s.UserID != userID {
		return
	}

	err = h.SessionProvider.Invalidate(s)
	if err != nil {
		return
	}

	resp = map[string]string{}
	return
}
