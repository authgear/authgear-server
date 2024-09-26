package oauth

import (
	"errors"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type ServiceIDPSessionProvider interface {
	Get(id string) (*idpsession.IDPSession, error)
	CheckSessionExpired(session *idpsession.IDPSession) (expired bool)
}

type OfflineGrantService struct {
	OAuthConfig    *config.OAuthConfig
	Clock          clock.Clock
	IDPSessions    ServiceIDPSessionProvider
	ClientResolver OAuthClientResolver

	OfflineGrants OfflineGrantStore
}

type CreateNewRefreshTokenResult struct {
	Token     string
	TokenHash string
}

func (s *OfflineGrantService) GetOfflineGrant(id string) (*OfflineGrant, error) {
	g, err := s.OfflineGrants.GetOfflineGrantWithoutExpireAt(id)
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
		idp, err := s.IDPSessions.Get(g.IDPSessionID)
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
	grant *OfflineGrant,
	serviceProviderID string,
) (*OfflineGrant, error) {
	expiry, err := s.ComputeOfflineGrantExpiry(grant)
	if err != nil {
		return nil, err
	}
	newGrant, err := s.OfflineGrants.AddOfflineGrantSAMLServiceProviderParticipant(
		grant.ID,
		serviceProviderID,
		expiry,
	)
	if err != nil {
		return nil, err
	}
	return newGrant, nil
}
