package auth

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
)

type HookProvider interface {
	DispatchEvent(payload event.Payload, user *model.User) error
}

type SessionManagementProvider interface {
	CookieConfig() *corehttp.CookieConfiguration
	Delete(AuthSession) error
}

type IDPSessionManager SessionManagementProvider
type AccessTokenSessionManager SessionManagementProvider

type SessionManager struct {
	AuthInfoStore       authinfo.Store
	UserProfileStore    userprofile.Store
	IdentityProvider    principal.IdentityProvider
	Hooks               HookProvider
	IDPSessions         IDPSessionManager
	AccessTokenSessions AccessTokenSessionManager
}

func (m *SessionManager) loadModels(session AuthSession) (*model.User, *model.Identity, error) {
	authInfo := &authinfo.AuthInfo{}
	if err := m.AuthInfoStore.GetAuth(session.AuthnAttrs().UserID, authInfo); err != nil {
		return nil, nil, err
	}

	profile, err := m.UserProfileStore.GetUserProfile(session.AuthnAttrs().UserID)
	if err != nil {
		return nil, nil, err
	}

	principal, err := m.IdentityProvider.GetPrincipalByID(session.AuthnAttrs().PrincipalID)
	if err != nil {
		return nil, nil, err
	}

	user := model.NewUser(*authInfo, profile)
	identity := model.NewIdentity(nil, principal)
	return &user, &identity, nil
}

func (m *SessionManager) resolveManagementProvider(session AuthSession) SessionManagementProvider {
	switch session.SessionType() {
	case SessionTypeIdentityProvider:
		return m.IDPSessions
	case SessionTypeOfflineGrant:
		return m.AccessTokenSessions
	default:
		panic("auth: unexpected session type")
	}
}

func (m *SessionManager) Logout(session AuthSession, rw http.ResponseWriter) error {
	user, identity, err := m.loadModels(session)
	if err != nil {
		return err
	}
	s := session.ToAPIModel()

	err = m.Hooks.DispatchEvent(
		event.SessionDeleteEvent{
			Reason:   string(SessionDeleteReasonLogout),
			User:     *user,
			Identity: *identity,
			Session:  *s,
		},
		user,
	)
	if err != nil {
		return err
	}

	provider := m.resolveManagementProvider(session)
	err = provider.Delete(session)
	if err != nil {
		return err
	}

	if cookie := provider.CookieConfig(); cookie != nil {
		cookie.Clear(rw)
	}

	return nil
}
