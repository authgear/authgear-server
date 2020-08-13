package auth

import (
	"errors"
	"net/http"
	"sort"

	"github.com/authgear/authgear-server/pkg/util/httputil"

	"github.com/authgear/authgear-server/pkg/lib/api/event"
	"github.com/authgear/authgear-server/pkg/lib/api/model"
)

type HookProvider interface {
	DispatchEvent(payload event.Payload) error
}

var ErrSessionNotFound = errors.New("session not found")

type UserProvider interface {
	Get(id string) (*model.User, error)
}

type SessionManagementProvider interface {
	ClearCookie() *http.Cookie
	Get(id string) (AuthSession, error)
	Update(AuthSession) error
	Delete(AuthSession) error
	List(userID string) ([]AuthSession, error)
}

type IDPSessionManager SessionManagementProvider
type AccessTokenSessionManager SessionManagementProvider

type SessionManager struct {
	Users               UserProvider
	Hooks               HookProvider
	IDPSessions         IDPSessionManager
	AccessTokenSessions AccessTokenSessionManager
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

func (m *SessionManager) invalidate(session AuthSession, reason SessionDeleteReason) (SessionManagementProvider, error) {
	user, err := m.Users.Get(session.AuthnAttrs().UserID)
	if err != nil {
		return nil, err
	}
	s := session.ToAPIModel()

	err = m.Hooks.DispatchEvent(&event.SessionDeleteEvent{
		Reason:  string(reason),
		User:    *user,
		Session: *s,
	})
	if err != nil {
		return nil, err
	}

	provider := m.resolveManagementProvider(session)
	err = provider.Delete(session)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

func (m *SessionManager) Logout(session AuthSession, rw http.ResponseWriter) error {
	provider, err := m.invalidate(session, SessionDeleteReasonLogout)
	if err != nil {
		return err
	}

	if cookie := provider.ClearCookie(); cookie != nil {
		httputil.UpdateCookie(rw, cookie)
	}

	return nil
}

func (m *SessionManager) Revoke(session AuthSession) error {
	_, err := m.invalidate(session, SessionDeleteReasonRevoke)
	if err != nil {
		return err
	}

	return nil
}

func (m *SessionManager) Get(id string) (AuthSession, error) {
	session, err := m.IDPSessions.Get(id)
	if err != nil && !errors.Is(err, ErrSessionNotFound) {
		return nil, err
	} else if err == nil {
		return session, nil
	}

	session, err = m.AccessTokenSessions.Get(id)
	if err != nil && !errors.Is(err, ErrSessionNotFound) {
		return nil, err
	} else if err == nil {
		return session, nil
	}

	return nil, ErrSessionNotFound
}

func (m *SessionManager) Update(session AuthSession) error {
	provider := m.resolveManagementProvider(session)
	err := provider.Update(session)
	if err != nil {
		return err
	}

	return nil
}

func (m *SessionManager) List(userID string) ([]AuthSession, error) {
	idpSessions, err := m.IDPSessions.List(userID)
	if err != nil {
		return nil, err
	}
	accessGrantSessions, err := m.AccessTokenSessions.List(userID)
	if err != nil {
		return nil, err
	}

	sessions := make([]AuthSession, len(idpSessions)+len(accessGrantSessions))
	copy(sessions[0:], idpSessions)
	copy(sessions[len(idpSessions):], accessGrantSessions)

	// Sort by creation time in descending order.
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].GetCreatedAt().After(sessions[j].GetCreatedAt())
	})

	return sessions, nil
}
