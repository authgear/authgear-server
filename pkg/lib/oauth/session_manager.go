package oauth

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/setutil"
)

type SessionManager struct {
	Store   OfflineGrantStore
	Config  *config.OAuthConfig
	Service OfflineGrantService
}

func (m *SessionManager) ClearCookie() []*http.Cookie {
	return []*http.Cookie{}
}

func (m *SessionManager) Get(id string) (session.ListableSession, error) {
	grant, err := m.Store.GetOfflineGrant(id)
	if errors.Is(err, ErrGrantNotFound) {
		return nil, session.ErrSessionNotFound
	} else if err != nil {
		return nil, err
	}
	return grant, nil
}

func (m *SessionManager) Delete(session session.ListableSession) error {
	err := m.Store.DeleteOfflineGrant(session.(*OfflineGrant))
	if err != nil {
		return err
	}
	return nil
}

func (m *SessionManager) List(userID string) ([]session.ListableSession, error) {
	grants, err := m.Store.ListOfflineGrants(userID)
	if err != nil {
		return nil, err
	}

	var sessions []session.ListableSession
	for _, session := range grants {
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (m *SessionManager) TerminateAllExcept(userID string, currentSession session.ListableSession) ([]session.ListableSession, error) {
	sessions, err := m.Store.ListOfflineGrants(userID)
	if err != nil {
		return nil, err
	}

	thirdPartyClientIDSet := make(setutil.Set[string])
	for _, c := range m.Config.Clients {
		if c.IsThirdParty() {
			thirdPartyClientIDSet[c.ClientID] = struct{}{}
		}
	}

	deletedSessions := []session.ListableSession{}
	for _, ss := range sessions {
		// skip third party client app refresh token
		// third party refresh token should be deleted through deleting authorization
		if _, ok := thirdPartyClientIDSet[ss.ClientID]; ok {
			continue
		}

		// skip the sessions that are in the same sso group
		if currentSession != nil && ss.IsSameSSOGroup(currentSession) {
			continue
		}

		if err := m.Delete(ss); err != nil {
			return nil, err
		}
		deletedSessions = append(deletedSessions, ss)
	}

	return deletedSessions, nil
}
