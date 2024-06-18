package oauth

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type OfflineGrantSessionManager interface {
	List(userID string) ([]session.ListableSession, error)
	Delete(session session.ListableSession) error
}

type AuthorizationService struct {
	AppID               config.AppID
	Store               AuthorizationStore
	Clock               clock.Clock
	OAuthSessionManager OfflineGrantSessionManager
}

func (s *AuthorizationService) GetByID(id string) (*Authorization, error) {
	return s.Store.GetByID(id)
}

func (s *AuthorizationService) ListByUser(userID string, filters ...AuthorizationFilter) ([]*Authorization, error) {
	as, err := s.Store.ListByUserID(userID)
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

func (s *AuthorizationService) Delete(a *Authorization) error {
	sessions, err := s.OAuthSessionManager.List(a.UserID)
	if err != nil {
		return err
	}

	// delete the offline grants that belong to the authorization
	for _, sess := range sessions {
		if offlineGrant, ok := sess.(*OfflineGrant); ok {
			// TODO(DEV-1403): Check all authorization ids?
			if offlineGrant.AuthorizationID == a.ID {
				err := s.OAuthSessionManager.Delete(sess)
				if err != nil {
					return err
				}
			}
		}
	}

	return s.Store.Delete(a)
}

func (s *AuthorizationService) CheckAndGrant(
	clientID string,
	userID string,
	scopes []string,
) (*Authorization, error) {
	timestamp := s.Clock.NowUTC()

	authz, err := s.Store.Get(userID, clientID)
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
		err = s.Store.Create(authz)
		if err != nil {
			return nil, err
		}
	} else {
		authz = authz.WithScopesAdded(scopes)
		authz.UpdatedAt = timestamp
		err = s.Store.UpdateScopes(authz)
		if err != nil {
			return nil, err
		}
	}

	return authz, nil
}

func (s *AuthorizationService) Check(
	clientID string,
	userID string,
	scopes []string,
) (*Authorization, error) {
	authz, err := s.Store.Get(userID, clientID)

	if err != nil {
		return nil, err
	}

	if !authz.IsAuthorized(scopes) {
		return nil, ErrAuthorizationScopesNotGranted
	}

	return authz, nil
}
