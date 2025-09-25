package oauth

import (
	"context"
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

func (m *SessionManager) Get(ctx context.Context, id string) (session.ListableSession, error) {
	// It is intentionally not to use Service.GetOfflineGrant here.
	grant, err := m.Store.GetOfflineGrantWithoutExpireAt(ctx, id)
	if errors.Is(err, ErrGrantNotFound) {
		return nil, session.ErrSessionNotFound
	} else if err != nil {
		return nil, err
	}
	return grant, nil
}

func (m *SessionManager) Delete(ctx context.Context, session session.ListableSession) error {
	err := m.Store.DeleteOfflineGrant(ctx, session.(*OfflineGrant))
	if err != nil {
		return err
	}
	return nil
}

func (m *SessionManager) List(ctx context.Context, userID string) ([]session.ListableSession, error) {
	grants, err := m.Store.ListOfflineGrants(ctx, userID)
	if err != nil {
		return nil, err
	}

	var sessions []session.ListableSession
	for _, session := range grants {
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (m *SessionManager) TerminateAllExcept(ctx context.Context, userID string, currentSession session.ResolvedSession) ([]session.ListableSession, error) {
	sessions, err := m.Store.ListOfflineGrants(ctx, userID)
	if err != nil {
		return nil, err
	}

	thirdPartyClientIDs := []string{}
	for _, c := range m.Config.Clients {
		if c.IsThirdParty() {
			thirdPartyClientIDs = append(thirdPartyClientIDs, c.ClientID)
		}
	}

	deletedSessions := []session.ListableSession{}
	for _, ss := range sessions {
		// skip the sessions that are in the same sso group
		if currentSession != nil && ss.IsSameSSOGroup(currentSession) {
			continue
		}

		// skip third party client app refresh token
		// third party refresh token should be deleted through deleting authorization
		initialTokenHashes, shouldRemoveOfflineGrant := ss.GetAllRemovableInitialTokenHashesExcludeClientIDs(thirdPartyClientIDs)
		if shouldRemoveOfflineGrant {
			if err := m.Delete(ctx, ss); err != nil {
				return nil, err
			}
			deletedSessions = append(deletedSessions, ss)
			continue
		}
		if len(initialTokenHashes) > 0 {
			// ComputeOfflineGrantExpiry is needed because Store.ListOfflineGrants
			// does not populate ExpireAtForResolvedSession.
			expiry, err := m.Service.ComputeOfflineGrantExpiry(ss)
			if err != nil {
				return nil, err
			}
			_, err = m.Store.RemoveOfflineGrantRefreshTokens(ctx, ss.ID, initialTokenHashes, expiry)
			if err != nil {
				return nil, err
			}
		}

	}

	return deletedSessions, nil
}

func (m *SessionManager) CleanUpForDeletingUserID(ctx context.Context, userID string) error {
	return m.Store.CleanUpForDeletingUserID(ctx, userID)
}
