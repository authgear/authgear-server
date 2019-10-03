package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	authSession "github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	authModel "github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// AttachLogoutHandler attach logout handler to server
func AttachLogoutHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/logout", &LogoutHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

// LogoutHandlerFactory creates new handler
type LogoutHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new handler
func (f LogoutHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &LogoutHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	h.AuditTrail = h.AuditTrail.WithRequest(request)
	return h.RequireAuthz(h, h)
}

/*
	@Operation POST /logout - Logout current session
		Logout current session.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@Response 200 {EmptyResponse}

		@Callback session_delete {SessionDeleteEvent}
		@Callback user_sync {UserSyncEvent}
*/
type LogoutHandler struct {
	AuthContext      coreAuth.ContextGetter     `dependency:"AuthContextGetter"`
	RequireAuthz     handler.RequireAuthz       `dependency:"RequireAuthz"`
	UserProfileStore userprofile.Store          `dependency:"UserProfileStore"`
	IdentityProvider principal.IdentityProvider `dependency:"IdentityProvider"`
	SessionProvider  session.Provider           `dependency:"SessionProvider"`
	SessionWriter    session.Writer             `dependency:"SessionWriter"`
	AuditTrail       audit.Trail                `dependency:"AuditTrail"`
	HookProvider     hook.Provider              `dependency:"HookProvider"`
	TxContext        db.TxContext               `dependency:"TxContext"`
}

// ProvideAuthzPolicy provides authorization policy of handler
func (h LogoutHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

func (h LogoutHandler) WithTx() bool {
	return true
}

// DecodeRequest decode request payload
func (h LogoutHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (handler.RequestPayload, error) {
	payload := handler.EmptyRequestPayload{}
	err := handler.DecodeJSONBody(request, resp, &payload)
	return payload, err
}

func (h LogoutHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	result, err := handler.Transactional(h.TxContext, func() (resp interface{}, err error) {
		resp, err = h.Handle()
		if err == nil {
			err = h.HookProvider.WillCommitTx()
		}
		return
	})
	if err == nil {
		h.HookProvider.DidCommitTx()
		h.SessionWriter.ClearSession(resp)
		handler.WriteResponse(resp, handler.APIResponse{Result: result})
	} else {
		handler.WriteResponse(resp, handler.APIResponse{Err: skyerr.MakeError(err)})
	}
}

// Handle api request
func (h LogoutHandler) Handle() (resp interface{}, err error) {
	authInfo, _ := h.AuthContext.AuthInfo()
	sess, _ := h.AuthContext.Session()
	if err = h.SessionProvider.Invalidate(sess); err != nil {
		err = skyerr.MakeError(err)
		return
	}

	resp = map[string]string{}

	var profile userprofile.UserProfile
	if profile, err = h.UserProfileStore.GetUserProfile(authInfo.ID); err != nil {
		return
	}

	var principal principal.Principal
	if principal, err = h.IdentityProvider.GetPrincipalByID(sess.PrincipalID); err != nil {
		return
	}

	user := authModel.NewUser(*authInfo, profile)
	identity := authModel.NewIdentity(h.IdentityProvider, principal)
	session := authSession.Format(sess)

	err = h.HookProvider.DispatchEvent(
		event.SessionDeleteEvent{
			Reason:   event.SessionDeleteReasonLogout,
			User:     user,
			Identity: identity,
			Session:  session,
		},
		&user,
	)
	if err != nil {
		return
	}

	h.AuditTrail.Log(audit.Entry{
		UserID: authInfo.ID,
		Event:  audit.EventLogout,
	})

	return
}
