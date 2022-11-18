package oauth

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

type SessionManager struct {
	Store   OfflineGrantStore
	Config  *config.OAuthConfig
	Service OfflineGrantService
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

	var sessions []session.Session
	for _, session := range grants {
		sessions = append(sessions, session)
	}
	return sessions, nil
}
