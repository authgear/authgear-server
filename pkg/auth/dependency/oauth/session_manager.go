package oauth

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/clock"
	"github.com/authgear/authgear-server/pkg/core/errors"
)

type SessionManager struct {
	Store OfflineGrantStore
	Clock clock.Clock
}

func (m *SessionManager) ClearCookie() *http.Cookie {
	return nil
}

func (m *SessionManager) Get(id string) (auth.AuthSession, error) {
	grant, err := m.Store.GetOfflineGrant(id)
	if errors.Is(err, ErrGrantNotFound) {
		return nil, auth.ErrSessionNotFound
	} else if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to get session")
	}
	return grant, nil
}

func (m *SessionManager) Update(session auth.AuthSession) error {
	err := m.Store.UpdateOfflineGrant(session.(*OfflineGrant))
	if err != nil {
		return errors.HandledWithMessage(err, "failed to update session")
	}
	return nil
}

func (m *SessionManager) Delete(session auth.AuthSession) error {
	err := m.Store.DeleteOfflineGrant(session.(*OfflineGrant))
	if err != nil {
		return errors.HandledWithMessage(err, "failed to invalidate session")
	}
	return nil
}

func (m *SessionManager) List(userID string) ([]auth.AuthSession, error) {
	grants, err := m.Store.ListOfflineGrants(userID)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to list sessions")
	}

	now := m.Clock.NowUTC()
	var sessions []auth.AuthSession
	for _, session := range grants {
		// ignore expired sessions
		if now.After(session.ExpireAt) {
			continue
		}

		sessions = append(sessions, session)
	}
	return sessions, nil
}
