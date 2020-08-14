package oauth

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

type SessionManager struct {
	Store OfflineGrantStore
	Clock clock.Clock
}

func (m *SessionManager) ClearCookie() *http.Cookie {
	return nil
}

func (m *SessionManager) Get(id string) (session.Session, error) {
	grant, err := m.Store.GetOfflineGrant(id)
	if errorutil.Is(err, ErrGrantNotFound) {
		return nil, session.ErrSessionNotFound
	} else if err != nil {
		return nil, errorutil.HandledWithMessage(err, "failed to get session")
	}
	return grant, nil
}

func (m *SessionManager) Update(session session.Session) error {
	err := m.Store.UpdateOfflineGrant(session.(*OfflineGrant))
	if err != nil {
		return errorutil.HandledWithMessage(err, "failed to update session")
	}
	return nil
}

func (m *SessionManager) Delete(session session.Session) error {
	err := m.Store.DeleteOfflineGrant(session.(*OfflineGrant))
	if err != nil {
		return errorutil.HandledWithMessage(err, "failed to invalidate session")
	}
	return nil
}

func (m *SessionManager) List(userID string) ([]session.Session, error) {
	grants, err := m.Store.ListOfflineGrants(userID)
	if err != nil {
		return nil, errorutil.HandledWithMessage(err, "failed to list sessions")
	}

	now := m.Clock.NowUTC()
	var sessions []session.Session
	for _, session := range grants {
		// ignore expired sessions
		if now.After(session.ExpireAt) {
			continue
		}

		sessions = append(sessions, session)
	}
	return sessions, nil
}
