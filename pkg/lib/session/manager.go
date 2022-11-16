package session

import (
	"errors"
	"net/http"
	"sort"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type signedOutEventOption struct {
	IsAdminAPI bool
}

var ErrSessionNotFound = errors.New("session not found")

type UserQuery interface {
	Get(id string, role accesscontrol.Role) (*model.User, error)
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

func (m *Manager) invalidate(session Session, option *signedOutEventOption) (ManagementService, error) {
	sessions, err := m.List(session.GetAuthenticationInfo().UserID)
	if err != nil {
		return nil, err
	}

	sort.Slice(sessions, func(i, j int) bool {
		if sessions[i].SessionType() == sessions[j].SessionType() {
			// Sort by creation time in ascending order.
			return sessions[i].GetCreatedAt().Before(sessions[j].GetCreatedAt())
		}

		// delete offline grant first
		if sessions[i].SessionType() == TypeOfflineGrant {
			return true
		}
		return false
	})

	var provider ManagementService
	for _, s := range sessions {
		// invalidate the sessions that are in the same sso group
		if s.IsSameSSOGroup(session) {
			p, err := m.invalidateSession(s, option)
			if err != nil {
				return nil, err
			}
			if s.Equal(session) {
				provider = p
			}
		}
	}

	if provider == nil {
		// if the current session doesn't appear in the sso group (e.g. sso disabled offline grant)
		// delete it here
		provider, err = m.invalidateSession(session, option)
		if err != nil {
			return nil, err
		}
	}

	return provider, nil

}

// invalidateSession should not be called directly
// invalidate should be called instead
func (m *Manager) invalidateSession(session Session, option *signedOutEventOption) (ManagementService, error) {
	sessionModel := session.ToAPIModel()

	provider := m.resolveManagementProvider(session)
	err := provider.Delete(session)
	if err != nil {
		return nil, err
	}

	if option != nil {
		err = m.Events.DispatchEvent(&nonblocking.UserSignedOutEventPayload{
			UserRef: model.UserRef{
				Meta: model.Meta{
					ID: session.GetAuthenticationInfo().UserID,
				},
			},
			Session:  *sessionModel,
			AdminAPI: option.IsAdminAPI,
		})
		if err != nil {
			return nil, err
		}
	}

	return provider, nil
}

func (m *Manager) Logout(session Session, rw http.ResponseWriter) error {
	provider, err := m.invalidate(session, &signedOutEventOption{IsAdminAPI: false})
	if err != nil {
		return err
	}

	for _, cookie := range provider.ClearCookie() {
		httputil.UpdateCookie(rw, cookie)
	}

	return nil
}

func (m *Manager) Revoke(session Session, isAdminAPI bool) error {
	_, err := m.invalidate(session, &signedOutEventOption{IsAdminAPI: isAdminAPI})
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) Delete(session Session) error {
	_, err := m.invalidate(session, nil)
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
