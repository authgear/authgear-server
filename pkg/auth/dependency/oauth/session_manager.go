package oauth

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type SessionManager struct {
	Store OfflineGrantStore
	Time  time.Provider
}

func (m *SessionManager) CookieConfig() *corehttp.CookieConfiguration {
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

	now := m.Time.NowUTC()
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
