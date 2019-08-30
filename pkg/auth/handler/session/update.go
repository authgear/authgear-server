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

func AttachUpdateHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/session/update", &UpdateHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type UpdateHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f UpdateHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &UpdateHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f UpdateHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type UpdateRequestPayload struct {
	SessionID string                 `json:"session_id"`
	Name      *string                `json:"name"`
	Data      map[string]interface{} `json:"data"`
}

// @JSONSchema
const UpdateRequestSchema = `
{
	"$id": "#SessionUpdateRequest",
	"type": "object",
	"properties": {
		"session_id": { "type": "string" },
		"name": { "type": "string" },
		"data": { "type": "object" }
	}
}
`

func (p UpdateRequestPayload) Validate() error {
	return nil
}

/*
	@Operation POST /session/update - Update name/custom data of session
		Update name/custom data of session. Custom data can be updated only if
		master key is used.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			Describe the session ID.
			@JSONSchema {SessionUpdateRequest}

		@Response 200 {EmptyResponse}
*/
type UpdateHandler struct {
	AuthContext     coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	TxContext       db.TxContext           `dependency:"TxContext"`
	SessionProvider session.Provider       `dependency:"SessionProvider"`
}

func (h UpdateHandler) WithTx() bool {
	return true
}

func (h UpdateHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := UpdateRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	return payload, err
}

func (h UpdateHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(UpdateRequestPayload)
	if !h.AuthContext.AccessKey().IsMasterKey() && payload.Data != nil {
		err = skyerr.NewError(skyerr.PermissionDenied, "Must update custom data using master key")
		return
	}

	userID := h.AuthContext.AuthInfo().ID
	sessionID := payload.SessionID

	s, err := h.SessionProvider.Get(sessionID)
	if err != nil {
		if err == session.ErrSessionNotFound {
			err = skyerr.NewError(skyerr.ResourceNotFound, "session not found")
		}
		return
	}
	if s.UserID != userID {
		err = skyerr.NewError(skyerr.ResourceNotFound, "session not found")
		return
	}

	err = h.SessionProvider.Update(s.ID, payload.Name, payload.Data)
	if err != nil {
		return
	}

	resp = map[string]string{}
	return
}
