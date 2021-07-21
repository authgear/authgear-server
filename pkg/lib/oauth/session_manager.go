package oauth

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type SessionManager struct {
	Store  OfflineGrantStore
	Clock  clock.Clock
	Config *config.OAuthConfig
}

func (m *SessionManager) ClearCookie() []*http.Cookie {
	return []*http.Cookie{}
}

func (m *SessionManager) Get(id string) (session.Session, error) {
	grant, err := m.Store.GetOfflineGrant(id)
	if errors.Is(err, ErrGrantNotFound) {
		return nil, session.ErrSessionNotFound
	} else if err != nil {
		return nil, err
	}
	return grant, nil
}

func (m *SessionManager) Delete(session session.Session) error {
	err := m.Store.DeleteOfflineGrant(session.(*OfflineGrant))
	if err != nil {
		return err
	}
	return nil
}

func (m *SessionManager) List(userID string) ([]session.Session, error) {
	grants, err := m.Store.ListOfflineGrants(userID)
	if err != nil {
		return nil, err
	}

	now := m.Clock.NowUTC()
	var sessions []session.Session
	for _, session := range grants {
		maxExpiry, err := ComputeOfflineGrantExpiryWithClients(session, m.Config)

		// ignore sessions without client
		if errors.Is(err, ErrGrantNotFound) {
			continue
		} else if err != nil {
			return nil, err
		}

		// ignore expired sessions
		if now.After(maxExpiry) {
			continue
		}

		sessions = append(sessions, session)
	}
	return sessions, nil
}
