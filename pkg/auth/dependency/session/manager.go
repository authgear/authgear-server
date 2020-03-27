package session

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type Manager struct {
	Store  Store
	Time   time.Provider
	Config config.SessionConfiguration
	Cookie CookieConfiguration
}

func (m *Manager) CookieConfig() *corehttp.CookieConfiguration {
	return (*corehttp.CookieConfiguration)(&m.Cookie)
}

func (m *Manager) Get(id string) (auth.AuthSession, error) {
	s, err := m.Store.Get(id)
	if errors.Is(err, ErrSessionNotFound) {
		return nil, auth.ErrSessionNotFound
	} else if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to get session")
	}
	return s, nil
}

func (m *Manager) Delete(session auth.AuthSession) error {
	err := m.Store.Delete(session.(*IDPSession))
	if err != nil {
		return errors.HandledWithMessage(err, "failed to invalidate session")
	}
	return nil
}

func (m *Manager) List(userID string) ([]auth.AuthSession, error) {
	storedSessions, err := m.Store.List(userID)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to list sessions")
	}

	now := m.Time.NowUTC()
	var sessions []auth.AuthSession
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
