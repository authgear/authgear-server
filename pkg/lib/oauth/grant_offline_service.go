package oauth

import (
	"context"
	"errors"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
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
	RemoteIP        httputil.RemoteIP
	UserAgentString httputil.UserAgentString

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
func (s *OfflineGrantService) AccessOfflineGrant(ctx context.Context, grantID string, refreshTokenHash string, accessEvent *access.Event, expireAt time.Time) (*OfflineGrant, error) {
	grant, err := s.OfflineGrants.UpdateOfflineGrantLastAccess(ctx, grantID, refreshTokenHash, *accessEvent, expireAt)
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

	expiry = s.computeRefreshTokenExpiryWithClient(expirableRefreshToken{
		CreatedAt:    session.CreatedAt,
		LastAccessAt: session.AccessInfo.LastAccess.Timestamp,
	}, clientConfig)
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

type expirableRefreshToken struct {
	CreatedAt    time.Time
	LastAccessAt time.Time
}

func (s *OfflineGrantService) computeRefreshTokenExpiryWithClient(token expirableRefreshToken, cfg *config.OAuthClientConfig) (expiry time.Time) {
	expiry = token.CreatedAt.Add(cfg.RefreshTokenLifetime.Duration())
	if *cfg.RefreshTokenIdleTimeoutEnabled {
		idleExpiry := token.LastAccessAt.Add(cfg.RefreshTokenIdleTimeout.Duration())
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
	now := s.Clock.NowUTC()
	accessEvent := access.NewEvent(now, s.RemoteIP, s.UserAgentString)

	accessInfo := access.Info{
		InitialAccess: accessEvent,
		LastAccess:    accessEvent,
	}

	newToken := GenerateToken()
	newTokenHash := HashToken(newToken)
	newGrant, err := s.OfflineGrants.AddOfflineGrantRefreshToken(
		ctx,
		grant.ID,
		accessInfo,
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

	// Housekeep the OfflineGrant to ensure the size of the object will not increase indefinitely
	newGrant, err = s.housekeepOfflineGrant(ctx, newGrant)
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

func (s *OfflineGrantService) housekeepOfflineGrant(ctx context.Context, grant *OfflineGrant) (*OfflineGrant, error) {
	now := s.Clock.NowUTC()

	// Remove expired refresh tokens
	tokenHashesToRemove := []string{}
	for idx, token := range grant.RefreshTokens {
		if idx == 0 {
			// Never remove the root token
			continue
		}
		var lastAccess time.Time
		// For backward compatibility
		if token.AccessInfo == nil {
			lastAccess = grant.AccessInfo.InitialAccess.Timestamp
		} else {
			lastAccess = token.AccessInfo.LastAccess.Timestamp
		}

		clientConfig := s.ClientResolver.ResolveClient(token.ClientID)
		if clientConfig == nil {
			// The client was removed, remove the refresh token
			tokenHashesToRemove = append(tokenHashesToRemove, token.TokenHash)
			continue
		}

		expiry := s.computeRefreshTokenExpiryWithClient(expirableRefreshToken{
			CreatedAt:    token.CreatedAt,
			LastAccessAt: lastAccess,
		}, clientConfig)

		if now.After(expiry) {
			tokenHashesToRemove = append(tokenHashesToRemove, token.TokenHash)
			continue
		}
	}

	expiry, err := s.ComputeOfflineGrantExpiry(grant)
	if err != nil {
		return nil, err
	}

	newGrant, err := s.OfflineGrants.RemoveOfflineGrantRefreshTokens(ctx, grant.ID, tokenHashesToRemove, expiry)
	if err != nil {
		return nil, err
	}

	return newGrant, err
}
