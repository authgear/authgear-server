package session

import (
	"errors"
	"net/http"
	"sort"

	"github.com/authgear/authgear-server/pkg/lib/api/event"
	"github.com/authgear/authgear-server/pkg/lib/api/model"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type HookProvider interface {
	DispatchEvent(payload event.Payload) error
}

var ErrSessionNotFound = errors.New("session not found")

type UserProvider interface {
	Get(id string) (*model.User, error)
}

type ManagementService interface {
	ClearCookie() *http.Cookie
	Get(id string) (Session, error)
	Update(Session) error
	Delete(Session) error
	List(userID string) ([]Session, error)
}

type IDPSessionManager ManagementService
type AccessTokenSessionManager ManagementService

type Manager struct {
	Users               UserProvider
	Hooks               HookProvider
	IDPSessions         IDPSessionManager
	AccessTokenSessions AccessTokenSessionManager
}

func (m *Manager) resolveManagementProvider(session Session) ManagementService {
	switch session.SessionType() {
	case TypeIdentityProvider:
		return m.IDPSessions
	case TypeOfflineGrant:
		return m.AccessTokenSessions
	default:
		panic("auth: unexpected session type")
	}
}

func (m *Manager) invalidate(session Session, reason DeleteReason) (ManagementService, error) {
	user, err := m.Users.Get(session.SessionAttrs().UserID)
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

func (m *Manager) Logout(session Session, rw http.ResponseWriter) error {
	provider, err := m.invalidate(session, DeleteReasonLogout)
	if err != nil {
		return err
	}

	if cookie := provider.ClearCookie(); cookie != nil {
		httputil.UpdateCookie(rw, cookie)
	}

	return nil
}

func (m *Manager) Revoke(session Session) error {
	_, err := m.invalidate(session, DeleteReasonRevoke)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) Get(id string) (Session, error) {
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

func (m *Manager) Update(session Session) error {
	provider := m.resolveManagementProvider(session)
	err := provider.Update(session)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) List(userID string) ([]Session, error) {
	idpSessions, err := m.IDPSessions.List(userID)
	if err != nil {
		return nil, err
	}
	accessGrantSessions, err := m.AccessTokenSessions.List(userID)
	if err != nil {
		return nil, err
	}

	sessions := make([]Session, len(idpSessions)+len(accessGrantSessions))
	copy(sessions[0:], idpSessions)
	copy(sessions[len(idpSessions):], accessGrantSessions)

	// Sort by creation time in descending order.
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].GetCreatedAt().After(sessions[j].GetCreatedAt())
	})

	return sessions, nil
}
