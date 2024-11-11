package session

import (
	"context"
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

type revokeEventOption struct {
	IsAdminAPI    bool
	IsTermination bool
}

var ErrSessionNotFound = errors.New("session not found")

type UserQuery interface {
	Get(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error)
	GetRaw(ctx context.Context, id string) (*user.User, error)
}

type ManagementService interface {
	ClearCookie() []*http.Cookie
	Get(ctx context.Context, id string) (ListableSession, error)
	Delete(ctx context.Context, s ListableSession) error
	List(ctx context.Context, userID string) ([]ListableSession, error)
	TerminateAllExcept(ctx context.Context, userID string, currentSession ResolvedSession) ([]ListableSession, error)
}

type IDPSessionManager ManagementService
type AccessTokenSessionManager ManagementService

type EventService interface {
	DispatchEventOnCommit(ctx context.Context, payload event.Payload) error
}

type Manager struct {
	IDPSessions         IDPSessionManager
	AccessTokenSessions AccessTokenSessionManager
	Events              EventService
}

func (m *Manager) resolveManagementProvider(session ListableSession) ManagementService {
	switch session.SessionType() {
	case TypeIdentityProvider:
		return m.IDPSessions
	case TypeOfflineGrant:
		return m.AccessTokenSessions
	default:
		panic("auth: unexpected session type")
	}
}

func (m *Manager) invalidate(ctx context.Context, session SessionBase, option *revokeEventOption) (
	[]ListableSession,
	ManagementService,
	error,
) {
	sessions, err := m.List(ctx, session.GetAuthenticationInfo().UserID)
	if err != nil {
		return nil, nil, err
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

	invalidatedSessions := []ListableSession{}

	var provider ManagementService
	for _, s := range sessions {
		s := s
		// invalidate the sessions that are in the same sso group
		if s.IsSameSSOGroup(session) {
			invalidatedSessions = append(invalidatedSessions, s)
			p, err := m.invalidateSession(ctx, s)
			if err != nil {
				return nil, nil, err
			}
			if s.EqualSession(session) {
				provider = p
			}
		}
	}

	sessionModels := []model.Session{}
	for _, s := range invalidatedSessions {
		sessionModels = append(sessionModels, *s.ToAPIModel())
	}

	if option != nil && len(sessionModels) > 0 {
		if option.IsTermination {
			err = m.Events.DispatchEventOnCommit(ctx, &nonblocking.UserSessionTerminatedEventPayload{
				UserRef: model.UserRef{
					Meta: model.Meta{
						ID: session.GetAuthenticationInfo().UserID,
					},
				},
				Sessions:        sessionModels,
				AdminAPI:        option.IsAdminAPI,
				TerminationType: nonblocking.UserSessionTerminationTypeIndividual,
			})
		} else {
			err = m.Events.DispatchEventOnCommit(ctx, &nonblocking.UserSignedOutEventPayload{
				UserRef: model.UserRef{
					Meta: model.Meta{
						ID: session.GetAuthenticationInfo().UserID,
					},
				},
				Sessions: sessionModels,
				AdminAPI: option.IsAdminAPI,
			})
		}
		if err != nil {
			return nil, nil, err
		}
	}

	return invalidatedSessions, provider, nil

}

// invalidateSession should not be called directly
// invalidate should be called instead
func (m *Manager) invalidateSession(ctx context.Context, session ListableSession) (ManagementService, error) {
	provider := m.resolveManagementProvider(session)
	err := provider.Delete(ctx, session)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

func (m *Manager) Logout(ctx context.Context, session SessionBase, rw http.ResponseWriter) ([]ListableSession, error) {
	invalidatedSessions, provider, err := m.invalidate(ctx, session, &revokeEventOption{IsAdminAPI: false, IsTermination: false})
	if err != nil {
		return nil, err
	}

	for _, cookie := range provider.ClearCookie() {
		httputil.UpdateCookie(rw, cookie)
	}

	return invalidatedSessions, nil
}

func (m *Manager) RevokeWithEvent(ctx context.Context, session SessionBase, isTermination bool, isAdminAPI bool) error {
	_, _, err := m.invalidate(ctx, session, &revokeEventOption{
		IsAdminAPI:    isAdminAPI,
		IsTermination: isTermination,
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) RevokeWithoutEvent(ctx context.Context, session SessionBase) error {
	_, _, err := m.invalidate(ctx, session, nil)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) TerminateAllExcept(ctx context.Context, userID string, currentSession ResolvedSession, isAdminAPI bool) error {
	idpSessions, err := m.IDPSessions.TerminateAllExcept(ctx, userID, currentSession)
	if err != nil {
		return err
	}
	accessGrantSessions, err := m.AccessTokenSessions.TerminateAllExcept(ctx, userID, currentSession)
	if err != nil {
		return err
	}

	sessionModels := []model.Session{}
	for _, s := range idpSessions {
		sessionModel := s.ToAPIModel()
		sessionModels = append(sessionModels, *sessionModel)
	}
	for _, s := range accessGrantSessions {
		sessionModel := s.ToAPIModel()
		sessionModels = append(sessionModels, *sessionModel)
	}

	var sessionTerminationType nonblocking.UserSessionTerminationType
	if currentSession == nil {
		sessionTerminationType = nonblocking.UserSessionTerminationTypeAll
	} else {
		sessionTerminationType = nonblocking.UserSessionTerminationTypeAllExceptCurrent
	}

	if len(sessionModels) > 0 {
		err = m.Events.DispatchEventOnCommit(ctx, &nonblocking.UserSessionTerminatedEventPayload{
			UserRef: model.UserRef{
				Meta: model.Meta{
					ID: userID,
				},
			},
			Sessions:        sessionModels,
			AdminAPI:        isAdminAPI,
			TerminationType: sessionTerminationType,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) Get(ctx context.Context, id string) (ListableSession, error) {
	session, err := m.IDPSessions.Get(ctx, id)
	if err != nil && !errors.Is(err, ErrSessionNotFound) {
		return nil, err
	} else if err == nil {
		return session, nil
	}

	session, err = m.AccessTokenSessions.Get(ctx, id)
	if err != nil && !errors.Is(err, ErrSessionNotFound) {
		return nil, err
	} else if err == nil {
		return session, nil
	}

	return nil, ErrSessionNotFound
}

func (m *Manager) List(ctx context.Context, userID string) ([]ListableSession, error) {
	idpSessions, err := m.IDPSessions.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	accessGrantSessions, err := m.AccessTokenSessions.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	sessions := make([]ListableSession, len(idpSessions)+len(accessGrantSessions))
	copy(sessions[0:], idpSessions)
	copy(sessions[len(idpSessions):], accessGrantSessions)

	// Sort by creation time in ascending order.
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].GetCreatedAt().Before(sessions[j].GetCreatedAt())
	})

	return sessions, nil
}
