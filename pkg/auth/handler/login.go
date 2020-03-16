package handler

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

// AttachLoginHandler attach login handler to server
func AttachLoginHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/login").
		Handler(server.FactoryToHandler(&LoginHandlerFactory{
			authDependency,
		})).
		Methods("OPTIONS", "POST")
}

// LoginHandlerFactory creates new handler
type LoginHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f LoginHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &LoginHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
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
	RequireAuthz         handler.RequireAuthz  `dependency:"RequireAuthz"`
	Validator            *validation.Validator `dependency:"Validator"`
	HookProvider         hook.Provider         `dependency:"HookProvider"`
	AuthnSessionProvider authnsession.Provider `dependency:"AuthnSessionProvider"`
	TxContext            db.TxContext          `dependency:"TxContext"`
	AuthnLoginProvider   authn.LoginProvider   `dependency:"AuthnLoginProvider"`
}

// ProvideAuthzPolicy provides authorization policy
func (h LoginHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
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
		h.AuthnSessionProvider.WriteResponse(resp, nil, err)
		return
	}

	var result interface{}
	err = hook.WithTx(h.HookProvider, h.TxContext, func() (err error) {
		result, err = h.Handle(payload)
		return
	})
	h.AuthnSessionProvider.WriteResponse(resp, result, err)
}

// Handle api request
func (h LoginHandler) Handle(payload LoginRequestPayload) (resp interface{}, err error) {
	authInfo, principal, err := h.AuthnLoginProvider.AuthenticateWithLoginID(loginid.LoginID{
		Key:   payload.LoginIDKey,
		Value: payload.LoginID,
	}, payload.Password)
	if err != nil {
		return
	}

	sess, err := h.AuthnSessionProvider.NewFromScratch(authInfo.ID, principal, coreAuth.SessionCreateReasonLogin)
	if err != nil {
		return
	}
	resp, err = h.AuthnSessionProvider.GenerateResponseAndUpdateLastLoginAt(*sess)
	if err != nil {
		return
	}

	return
}
