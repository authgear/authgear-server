package idpsession

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type CookieManager interface {
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type Manager struct {
	Store     Store
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

func (m *Manager) Get(id string) (session.ListableSession, error) {
	s, err := m.Store.Get(id)
	if errors.Is(err, ErrSessionNotFound) {
		return nil, session.ErrSessionNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return s, nil
}

func (m *Manager) Delete(session session.ListableSession) error {
	err := m.Store.Delete(session.(*IDPSession))
	if err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}
	return nil
}

func (m *Manager) List(userID string) ([]session.ListableSession, error) {
	storedSessions, err := m.Store.List(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	var sessions []session.ListableSession
	for _, session := range storedSessions {
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (m *Manager) TerminateAllExcept(userID string, currentSession session.ResolvedSession) ([]session.ListableSession, error) {
	sessions, err := m.Store.List(userID)
	if err != nil {
		return nil, err
	}

	deletedSessions := []session.ListableSession{}
	for _, ss := range sessions {
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
