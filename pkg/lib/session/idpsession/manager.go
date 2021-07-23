package idpsession

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type CookieManager interface {
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type Manager struct {
	Store     Store
	Clock     clock.Clock
	Config    *config.SessionConfig
	Cookies   CookieManager
	CookieDef session.CookieDef
}

func (m *Manager) ClearCookie() []*http.Cookie {
	return []*http.Cookie{
		m.Cookies.ClearCookie(m.CookieDef.Def),
		m.Cookies.ClearCookie(m.CookieDef.SameSiteStrictDef),
	}
}

func (m *Manager) Get(id string) (session.Session, error) {
	s, err := m.Store.Get(id)
	if errors.Is(err, ErrSessionNotFound) {
		return nil, session.ErrSessionNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return s, nil
}

func (m *Manager) Delete(session session.Session) error {
	err := m.Store.Delete(session.(*IDPSession))
	if err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}
	return nil
}

func (m *Manager) List(userID string) ([]session.Session, error) {
	storedSessions, err := m.Store.List(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
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
