package idpsession

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

type Manager struct {
	Store         Store
	Clock         clock.Clock
	Config        *config.SessionConfig
	CookieFactory CookieFactory
	CookieDef     CookieDef
}

func (m *Manager) ClearCookie() *http.Cookie {
	return m.CookieFactory.ClearCookie(m.CookieDef.Def)
}

func (m *Manager) Get(id string) (session.Session, error) {
	s, err := m.Store.Get(id)
	if errorutil.Is(err, ErrSessionNotFound) {
		return nil, session.ErrSessionNotFound
	} else if err != nil {
		return nil, errorutil.HandledWithMessage(err, "failed to get session")
	}
	return s, nil
}

func (m *Manager) Update(session session.Session) error {
	s := session.(*IDPSession)
	expiry := computeSessionStorageExpiry(s, m.Config)
	err := m.Store.Update(s, expiry)
	if err != nil {
		return errorutil.HandledWithMessage(err, "failed to update session")
	}
	return nil
}

func (m *Manager) Delete(session session.Session) error {
	err := m.Store.Delete(session.(*IDPSession))
	if err != nil {
		return errorutil.HandledWithMessage(err, "failed to invalidate session")
	}
	return nil
}

func (m *Manager) List(userID string) ([]session.Session, error) {
	storedSessions, err := m.Store.List(userID)
	if err != nil {
		return nil, errorutil.HandledWithMessage(err, "failed to list sessions")
	}

	now := m.Clock.NowUTC()
	var sessions []session.Session
	for _, session := range storedSessions {
		maxExpiry := computeSessionStorageExpiry(session, m.Config)
		// ignore expired sessions
		if now.After(maxExpiry) {
			continue
		}

		sessions = append(sessions, session)
	}
	return sessions, nil
}
