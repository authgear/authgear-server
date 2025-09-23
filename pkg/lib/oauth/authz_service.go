package oauth

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type OfflineGrantSessionManager interface {
	List(ctx context.Context, userID string) ([]session.ListableSession, error)
	Delete(ctx context.Context, session session.ListableSession) error
}

type AuthorizationService struct {
	AppID               config.AppID
	Store               AuthorizationStore
	Clock               clock.Clock
	OAuthSessionManager OfflineGrantSessionManager
	OfflineGrantService *OfflineGrantService
	OfflineGrantStore   OfflineGrantStore
}

func (s *AuthorizationService) GetByID(ctx context.Context, id string) (*Authorization, error) {
	return s.Store.GetByID(ctx, id)
}

func (s *AuthorizationService) ListByUser(ctx context.Context, userID string, filters ...AuthorizationFilter) ([]*Authorization, error) {
	as, err := s.Store.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	filtered := []*Authorization{}
	for _, a := range as {
		keep := true
		for _, f := range filters {
			if !f.Keep(a) {
				keep = false
				break
			}
		}
		if keep {
			filtered = append(filtered, a)
		}
	}

	return filtered, nil
}

func (s *AuthorizationService) Delete(ctx context.Context, a *Authorization) error {
	sessions, err := s.OAuthSessionManager.List(ctx, a.UserID)
	if err != nil {
		return err
	}

	// delete the offline grants that belong to the authorization
	for _, sess := range sessions {
		if offlineGrant, ok := sess.(*OfflineGrant); ok {
			initialTokenHashes, shouldRemoveOfflineGrant := offlineGrant.GetRemovableInitialTokenHashesByAuthorizationID(a.ID)
			if shouldRemoveOfflineGrant {
				err := s.OAuthSessionManager.Delete(ctx, sess)
				if err != nil {
					return err
				}
			} else if len(initialTokenHashes) > 0 {
				// ComputeOfflineGrantExpiry is needed because SessionManager.List
				// does not populate ExpireAtForResolvedSession.
				expiry, err := s.OfflineGrantService.ComputeOfflineGrantExpiry(offlineGrant)
				if err != nil {
					return err
				}
				_, err = s.OfflineGrantStore.RemoveOfflineGrantRefreshTokens(ctx, offlineGrant.ID, initialTokenHashes, expiry)
				if err != nil {
					return err
				}
			}
		}
	}

	return s.Store.Delete(ctx, a)
}

func (s *AuthorizationService) CheckAndGrant(
	ctx context.Context,
	clientID string,
	userID string,
	scopes []string,
) (*Authorization, error) {
	timestamp := s.Clock.NowUTC()

	authz, err := s.Store.Get(ctx, userID, clientID)
	if err == nil && authz.IsAuthorized(scopes) {
		return authz, nil
	} else if err != nil && !errors.Is(err, ErrAuthorizationNotFound) {
		return nil, err
	}

	// Authorization of requested scopes not granted, requesting consent.
	// TODO(oauth): request consent, for now just always implicitly grant scopes.
	if authz == nil {
		authz = &Authorization{
			ID:        uuid.New(),
			AppID:     string(s.AppID),
			ClientID:  clientID,
			UserID:    userID,
			CreatedAt: timestamp,
			UpdatedAt: timestamp,
			Scopes:    scopes,
		}
		err = s.Store.Create(ctx, authz)
		if err != nil {
			return nil, err
		}
	} else {
		authz = authz.WithScopesAdded(scopes)
		authz.UpdatedAt = timestamp
		err = s.Store.UpdateScopes(ctx, authz)
		if err != nil {
			return nil, err
		}
	}

	return authz, nil
}

func (s *AuthorizationService) Check(
	ctx context.Context,
	clientID string,
	userID string,
	scopes []string,
) (*Authorization, error) {
	authz, err := s.Store.Get(ctx, userID, clientID)

	if err != nil {
		return nil, err
	}

	if !authz.IsAuthorized(scopes) {
		return nil, ErrAuthorizationScopesNotGranted
	}

	return authz, nil
}
