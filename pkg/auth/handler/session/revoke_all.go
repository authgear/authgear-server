package session

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	authSession "github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
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
	return h.RequireAuthz(handler.APIHandlerToHandler(h, h.TxContext), h)
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
	AuthContext      coreAuth.ContextGetter     `dependency:"AuthContextGetter"`
	RequireAuthz     handler.RequireAuthz       `dependency:"RequireAuthz"`
	TxContext        db.TxContext               `dependency:"TxContext"`
	SessionProvider  session.Provider           `dependency:"SessionProvider"`
	IdentityProvider principal.IdentityProvider `dependency:"IdentityProvider"`
	UserProfileStore userprofile.Store          `dependency:"UserProfileStore"`
	HookProvider     hook.Provider              `dependency:"HookProvider"`
}

func (h RevokeAllHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

func (h RevokeAllHandler) WithTx() bool {
	return true
}

func (h RevokeAllHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (handler.RequestPayload, error) {
	payload := handler.EmptyRequestPayload{}
	err := handler.DecodeJSONBody(request, resp, &payload)
	return payload, err
}

func (h RevokeAllHandler) Handle(req interface{}) (resp interface{}, err error) {
	authInfo, _ := h.AuthContext.AuthInfo()
	userID := authInfo.ID
	sess, _ := h.AuthContext.Session()
	sessionID := sess.ID

	profile, err := h.UserProfileStore.GetUserProfile(userID)
	if err != nil {
		return
	}
	user := model.NewUser(*authInfo, profile)

	sessions, err := h.SessionProvider.List(userID)
	if err != nil {
		return
	}

	n := 0
	for _, session := range sessions {
		if session.ID == sessionID {
			continue
		}
		sessions[n] = session
		n++

		var principal principal.Principal
		if principal, err = h.IdentityProvider.GetPrincipalByID(session.PrincipalID); err != nil {
			return
		}
		identity := model.NewIdentity(h.IdentityProvider, principal)
		sessionModel := authSession.Format(session)

		err = h.HookProvider.DispatchEvent(
			event.SessionDeleteEvent{
				Reason:   event.SessionDeleteReasonRevoke,
				User:     user,
				Identity: identity,
				Session:  sessionModel,
			},
			&user,
		)
		if err != nil {
			return
		}
	}
	sessions = sessions[:n]

	err = h.SessionProvider.InvalidateBatch(sessions)
	if err != nil {
		return
	}

	resp = map[string]string{}
	return
}
