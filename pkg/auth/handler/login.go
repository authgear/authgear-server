package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// AttachLoginHandler attach login handler to server
func AttachLoginHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/login", &LoginHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

// LoginHandlerFactory creates new handler
type LoginHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f LoginHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &LoginHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	h.AuditTrail = h.AuditTrail.WithRequest(request)
	return handler.RequireAuthz(h, h.AuthContext, h)
}

// LoginRequestPayload login handler request payload
type LoginRequestPayload struct {
	LoginIDKey string `json:"login_id_key"`
	LoginID    string `json:"login_id"`
	Realm      string `json:"realm"`
	Password   string `json:"password"`
}

// @JSONSchema
const LoginRequestSchema = `
{
	"$id": "#LoginRequest",
	"type": "object",
	"properties": {
		"login_id_key": { "type": "string" },
		"login_id": { "type": "string" },
		"realm": { "type": "string" },
		"password": { "type": "string" }
	}
}
`

// Validate request payload
func (p LoginRequestPayload) Validate() error {
	if p.LoginID == "" {
		return skyerr.NewInvalidArgument("empty login ID", []string{"login_id"})
	}

	if p.Password == "" {
		return skyerr.NewInvalidArgument("empty password", []string{"password"})
	}

	return nil
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
	AuthContext          coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	AuthInfoStore        authinfo.Store         `dependency:"AuthInfoStore"`
	PasswordAuthProvider password.Provider      `dependency:"PasswordAuthProvider"`
	AuditTrail           audit.Trail            `dependency:"AuditTrail"`
	HookProvider         hook.Provider          `dependency:"HookProvider"`
	AuthnSessionProvider authnsession.Provider  `dependency:"AuthnSessionProvider"`
	TxContext            db.TxContext           `dependency:"TxContext"`
}

// ProvideAuthzPolicy provides authorization policy
func (h LoginHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

// DecodeRequest decode request payload
func (h LoginHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (payload LoginRequestPayload, err error) {
	err = handler.DecodeJSONBody(request, resp, &payload)

	if err != nil {
		return
	}

	if payload.Realm == "" {
		payload.Realm = password.DefaultRealm
	}

	return
}

func (h LoginHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var err error
	var result interface{}
	defer func() {
		if err == nil {
			h.HookProvider.DidCommitTx()
		}
		h.AuthnSessionProvider.WriteResponse(resp, result, err)
	}()

	payload, err := h.DecodeRequest(req, resp)
	if err != nil {
		return
	}

	if err = payload.Validate(); err != nil {
		return
	}

	result, err = handler.Transactional(h.TxContext, func() (result interface{}, err error) {
		result, err = h.Handle(payload)
		if err == nil {
			err = h.HookProvider.WillCommitTx()
		}
		return
	})
}

// Handle api request
func (h LoginHandler) Handle(payload LoginRequestPayload) (resp interface{}, err error) {
	fetchedAuthInfo := authinfo.AuthInfo{}

	defer func() {
		if err != nil {
			h.AuditTrail.Log(audit.Entry{
				UserID: fetchedAuthInfo.ID,
				Event:  audit.EventLoginFailure,
			})
		} else {
			h.AuditTrail.Log(audit.Entry{
				UserID: fetchedAuthInfo.ID,
				Event:  audit.EventLoginSuccess,
			})
		}
	}()

	if payload.LoginIDKey != "" {
		loginIDs := []password.LoginID{
			password.LoginID{Key: payload.LoginIDKey, Value: payload.LoginID},
		}
		if err = h.PasswordAuthProvider.ValidateLoginIDs(loginIDs); err != nil {
			return
		}
	}

	if valid := h.PasswordAuthProvider.IsRealmValid(payload.Realm); !valid {
		err = skyerr.NewInvalidArgument("realm is not allowed", []string{"realm"})
		return
	}

	principal, err := h.getPrincipal(payload.Password, payload.LoginIDKey, payload.LoginID, payload.Realm)
	if err != nil {
		return
	}

	if err = h.AuthInfoStore.GetAuth(principal.UserID, &fetchedAuthInfo); err != nil {
		if err == skydb.ErrUserNotFound {
			err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
			return
		}
		// TODO: more error handling here if necessary
		err = skyerr.NewResourceFetchFailureErr("login_id", payload.LoginID)
		return
	}

	if err = checkUserIsNotDisabled(&fetchedAuthInfo); err != nil {
		return
	}

	sess, err := h.AuthnSessionProvider.NewFromScratch(fetchedAuthInfo.ID, principal, coreAuth.SessionCreateReasonLogin)
	if err != nil {
		return
	}
	resp, err = h.AuthnSessionProvider.GenerateResponseAndUpdateLastLoginAt(*sess)
	if err != nil {
		return
	}

	return
}

func (h LoginHandler) getPrincipal(pwd string, loginIDKey string, loginID string, realm string) (*password.Principal, error) {
	var principal password.Principal
	err := h.PasswordAuthProvider.GetPrincipalByLoginIDWithRealm(loginIDKey, loginID, realm, &principal)
	if err != nil {
		if err == skydb.ErrUserNotFound {
			return nil, skyerr.NewError(skyerr.ResourceNotFound, "user not found")
		}
		// TODO: more error handling here if necessary
		return nil, skyerr.NewResourceFetchFailureErr("login_id", loginID)
	}

	if !principal.IsSamePassword(pwd) {
		return nil, skyerr.NewError(skyerr.InvalidCredentials, "login_id or password incorrect")
	}

	return &principal, nil
}
