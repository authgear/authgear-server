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
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	skyerr "github.com/skygeario/skygear-server/pkg/core/xskyerr"
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
	return h.RequireAuthz(handler.APIHandlerToHandler(hook.WrapHandler(h.HookProvider, h), h.TxContext), h)
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
	AuthContext      coreAuth.ContextGetter     `dependency:"AuthContextGetter"`
	RequireAuthz     handler.RequireAuthz       `dependency:"RequireAuthz"`
	TxContext        db.TxContext               `dependency:"TxContext"`
	SessionProvider  session.Provider           `dependency:"SessionProvider"`
	IdentityProvider principal.IdentityProvider `dependency:"IdentityProvider"`
	UserProfileStore userprofile.Store          `dependency:"UserProfileStore"`
	HookProvider     hook.Provider              `dependency:"HookProvider"`
}

func (h RevokeHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

func (h RevokeHandler) WithTx() bool {
	return true
}

func (h RevokeHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (handler.RequestPayload, error) {
	payload := RevokeRequestPayload{}
	err := handler.DecodeJSONBody(request, resp, &payload)
	return payload, err
}

func (h RevokeHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(RevokeRequestPayload)

	authInfo, _ := h.AuthContext.AuthInfo()
	userID := authInfo.ID
	sess, _ := h.AuthContext.Session()
	sessionID := payload.SessionID

	if sess.ID == sessionID {
		// TODO(error): JSON schema
		err = skyerr.NewInvalid("must not revoke current session")
		return
	}

	// ignore session not found errors
	s, err := h.SessionProvider.Get(sessionID)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			err = nil
			resp = map[string]string{}
		}
		return
	}
	if s.UserID != userID {
		resp = map[string]string{}
		return
	}

	var profile userprofile.UserProfile
	if profile, err = h.UserProfileStore.GetUserProfile(s.UserID); err != nil {
		return
	}

	var principal principal.Principal
	if principal, err = h.IdentityProvider.GetPrincipalByID(s.PrincipalID); err != nil {
		return
	}

	user := model.NewUser(*authInfo, profile)
	identity := model.NewIdentity(h.IdentityProvider, principal)
	session := authSession.Format(s)

	err = h.HookProvider.DispatchEvent(
		event.SessionDeleteEvent{
			Reason:   event.SessionDeleteReasonRevoke,
			User:     user,
			Identity: identity,
			Session:  session,
		},
		&user,
	)
	if err != nil {
		return
	}

	err = h.SessionProvider.Invalidate(s)
	if err != nil {
		return
	}

	resp = map[string]string{}
	return
}
