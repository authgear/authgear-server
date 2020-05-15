package handler

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	interactionflows "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
	coreauth "github.com/skygeario/skygear-server/pkg/core/auth"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

// AttachLoginHandler attach login handler to server
func AttachLoginHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/login").
		Handler(auth.MakeHandler(authDependency, newLoginHandler)).
		Methods("OPTIONS", "POST")
}

// LoginRequestPayload login handler request payload
type LoginRequestPayload struct {
	LoginIDKey string `json:"login_id_key"`
	LoginID    string `json:"login_id"`
	Password   string `json:"password"`
}

// @JSONSchema
const LoginRequestSchema = `
{
	"$id": "#LoginRequest",
	"type": "object",
	"properties": {
		"login_id_key": { "type": "string", "minLength": 1 },
		"login_id": { "type": "string", "minLength": 1 },
		"password": { "type": "string", "minLength": 1 }
	},
	"required": ["login_id", "password"]
}
`

type LoginInteractionFlow interface {
	LoginWithLoginIDPassword(
		clientID string, loginIDKey string, loginID string, password string,
	) (*interactionflows.AuthResult, error)
}

/*
	@Operation POST /login - Login using password
		Login user with login ID and password.

		@Tag User

		@RequestBody
			Describe login ID and password.
			@JSONSchema {LoginRequest}

		@Response 200
			Logged in user and access token.
			@JSONSchema {AuthResponse}

		@Callback session_create {SessionCreateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type LoginHandler struct {
	Validator    *validation.Validator
	Interactions LoginInteractionFlow
	TxContext    db.TxContext
}

// ProvideAuthzPolicy provides authorization policy
func (h LoginHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireClient
}

// DecodeRequest decode request payload
func (h LoginHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (payload LoginRequestPayload, err error) {
	err = handler.BindJSONBody(request, resp, h.Validator, "#LoginRequest", &payload)
	return
}

func (h LoginHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var err error

	payload, err := h.DecodeRequest(req, resp)
	if err != nil {
		handler.WriteResponse(resp, handler.APIResponse{Error: err})
		return
	}

	var result *interactionflows.AuthResult
	err = db.WithTx(h.TxContext, func() (err error) {
		result, err = h.Interactions.LoginWithLoginIDPassword(
			coreauth.GetAccessKey(req.Context()).Client.ClientID(),
			payload.LoginIDKey,
			payload.LoginID,
			payload.Password,
		)
		return
	})
	if err != nil {
		handler.WriteResponse(resp, handler.APIResponse{Error: err})
		return
	}

	result.WriteResponse(resp)
}
