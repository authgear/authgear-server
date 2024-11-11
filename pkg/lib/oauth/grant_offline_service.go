package oauth

import (
	"context"
	"errors"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type ServiceIDPSessionProvider interface {
	Get(ctx context.Context, id string) (*idpsession.IDPSession, error)
	CheckSessionExpired(session *idpsession.IDPSession) (expired bool)
}
type OfflineGrantServiceAccessEventProvider interface {
	RecordAccess(ctx context.Context, sessionID string, expiry time.Time, event *access.Event) error
}

type OfflineGrantServiceMeterService interface {
	TrackActiveUser(ctx context.Context, userID string) error
}

type OfflineGrantService struct {
	OAuthConfig    *config.OAuthConfig
	Clock          clock.Clock
	IDPSessions    ServiceIDPSessionProvider
	ClientResolver OAuthClientResolver
	AccessEvents   OfflineGrantServiceAccessEventProvider
	MeterService   OfflineGrantServiceMeterService

	OfflineGrants OfflineGrantStore
}

type CreateNewRefreshTokenResult struct {
	Token     string
	TokenHash string
}

// AccessOfflineGrant accesses oauth offline grant with 3 targeted side effects
// 1. set grant.AccessInfo.LastAccess to new accessEvent (inside UpdateOfflineGrantLastAccess)
// 2. call RecordAccess
// 3. call TrackActiveUser
func (s *OfflineGrantService) AccessOfflineGrant(ctx context.Context, grantID string, accessEvent *access.Event, expireAt time.Time) (*OfflineGrant, error) {
	grant, err := s.OfflineGrants.UpdateOfflineGrantLastAccess(ctx, grantID, *accessEvent, expireAt)
	if err != nil {
		return nil, err
	}

	err = s.AccessEvents.RecordAccess(ctx, grant.ID, grant.ExpireAtForResolvedSession, accessEvent)
	if err != nil {
		return nil, err
	}

	err = s.MeterService.TrackActiveUser(ctx, grant.Attrs.UserID)
	if err != nil {
		return nil, err
	}

	return grant, nil
}

func (s *OfflineGrantService) GetOfflineGrant(ctx context.Context, id string) (*OfflineGrant, error) {
	g, err := s.OfflineGrants.GetOfflineGrantWithoutExpireAt(ctx, id)
	if err != nil {
		return nil, err
	}

	expiry, err := s.ComputeOfflineGrantExpiry(g)
	if err != nil {
		return nil, err
	}
	g.ExpireAtForResolvedSession = expiry

	now := s.Clock.NowUTC()
	if now.After(g.ExpireAtForResolvedSession) {
		return nil, ErrGrantNotFound
	}

	// Check IDP session is valid.
	if g.SSOEnabled && g.IDPSessionID != "" {
		idp, err := s.IDPSessions.Get(ctx, g.IDPSessionID)
		if err != nil {
			if errors.Is(err, idpsession.ErrSessionNotFound) {
				return nil, ErrGrantNotFound
			}
			return nil, err
		}

		idpSessionExpired := s.IDPSessions.CheckSessionExpired(idp)
		if idpSessionExpired {
			return nil, ErrGrantNotFound
		}
	}

	return g, nil
}

func (s *OfflineGrantService) ComputeOfflineGrantExpiry(session *OfflineGrant) (expiry time.Time, err error) {
	clientConfig := s.ClientResolver.ResolveClient(session.InitialClientID)

	if clientConfig == nil {
		err = ErrGrantNotFound
		return
	}

	expiry = s.computeOfflineGrantExpiryWithClient(session, clientConfig)
	return
}

func (s *OfflineGrantService) CheckSessionExpired(session *OfflineGrant) (bool, time.Time, error) {
	now := s.Clock.NowUTC()
	expiry, err := s.ComputeOfflineGrantExpiry(session)
	if errors.Is(err, ErrGrantNotFound) {
		return true, now, nil
	} else if err != nil {
		return false, time.Time{}, err
	}

	offlineGrantExpired := now.After(expiry)
	return offlineGrantExpired, expiry, nil
}

func (s *OfflineGrantService) computeOfflineGrantExpiryWithClient(session *OfflineGrant, cfg *config.OAuthClientConfig) (expiry time.Time) {
	expiry = session.CreatedAt.Add(cfg.RefreshTokenLifetime.Duration())
	if *cfg.RefreshTokenIdleTimeoutEnabled {
		idleExpiry := session.AccessInfo.LastAccess.Timestamp.Add(cfg.RefreshTokenIdleTimeout.Duration())
		if idleExpiry.Before(expiry) {
			expiry = idleExpiry
		}
	}
	return
}

func (s *OfflineGrantService) CreateNewRefreshToken(
	ctx context.Context,
	grant *OfflineGrant,
	clientID string,
	scopes []string,
	authorizationID string,
	dpopJKT string,
) (*CreateNewRefreshTokenResult, *OfflineGrant, error) {
	expiry, err := s.ComputeOfflineGrantExpiry(grant)
	if err != nil {
		return nil, nil, err
	}
	newToken := GenerateToken()
	newTokenHash := HashToken(newToken)
	newGrant, err := s.OfflineGrants.AddOfflineGrantRefreshToken(
		ctx,
		grant.ID,
		expiry,
		newTokenHash,
		clientID,
		scopes,
		authorizationID,
		dpopJKT,
	)
	if err != nil {
		return nil, nil, err
	}
	result := &CreateNewRefreshTokenResult{
		Token:     newToken,
		TokenHash: newTokenHash,
	}
	return result, newGrant, nil
}

func (s *OfflineGrantService) AddSAMLServiceProviderParticipant(
	ctx context.Context,
	grant *OfflineGrant,
	serviceProviderID string,
) (*OfflineGrant, error) {
	expiry, err := s.ComputeOfflineGrantExpiry(grant)
	if err != nil {
		return nil, err
	}
	newGrant, err := s.OfflineGrants.AddOfflineGrantSAMLServiceProviderParticipant(
		ctx,
		grant.ID,
		serviceProviderID,
		expiry,
	)
	if err != nil {
		return nil, err
	}
	return newGrant, nil
}
