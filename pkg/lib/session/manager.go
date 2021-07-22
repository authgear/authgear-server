package session

import (
	"errors"
	"net/http"
	"sort"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

var ErrSessionNotFound = errors.New("session not found")

type UserQuery interface {
	Get(id string) (*model.User, error)
	GetRaw(id string) (*user.User, error)
}

type ManagementService interface {
	ClearCookie() []*http.Cookie
	Get(id string) (Session, error)
	Delete(Session) error
	List(userID string) ([]Session, error)
}

type IDPSessionManager ManagementService
type AccessTokenSessionManager ManagementService

type EventService interface {
	DispatchEvent(payload event.Payload) error
}

type Manager struct {
	Users               UserQuery
	IDPSessions         IDPSessionManager
	AccessTokenSessions AccessTokenSessionManager
	Events              EventService
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

func (m *Manager) invalidate(session Session, reason DeleteReason, isAdminAPI bool) (ManagementService, error) {
	sessionModel := session.ToAPIModel()

	user, err := m.Users.Get(session.GetUserID())
	if err != nil {
		return nil, err
	}

	provider := m.resolveManagementProvider(session)
	err = provider.Delete(session)
	if err != nil {
		return nil, err
	}

	err = m.Events.DispatchEvent(&nonblocking.UserSignedOutEventPayload{
		User:     *user,
		Session:  *sessionModel,
		AdminAPI: isAdminAPI,
	})
	if err != nil {
		return nil, err
	}

	return provider, nil
}

func (m *Manager) Logout(session Session, rw http.ResponseWriter) error {
	provider, err := m.invalidate(session, DeleteReasonLogout, false)
	if err != nil {
		return err
	}

	for _, cookie := range provider.ClearCookie() {
		httputil.UpdateCookie(rw, cookie)
	}

	return nil
}

func (m *Manager) Revoke(session Session, isAdminAPI bool) error {
	_, err := m.invalidate(session, DeleteReasonRevoke, isAdminAPI)
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

	// Sort by creation time in ascending order.
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].GetCreatedAt().Before(sessions[j].GetCreatedAt())
	})

	return sessions, nil
}
