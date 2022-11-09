package oauth

import (
	"errors"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type OfflineGrantService struct {
	OAuthConfig *config.OAuthConfig
	Clock       clock.Clock
}

func (s *OfflineGrantService) IsValid(session *OfflineGrant) (valid bool, expiry time.Time, err error) {
	now := s.Clock.NowUTC()
	expiry, err = s.ComputeOfflineGrantExpiry(session)
	if errors.Is(err, ErrGrantNotFound) {
		return false, now, nil
	} else if err != nil {
		return
	}

	return now.Before(expiry), expiry, nil
}

func (s *OfflineGrantService) ComputeOfflineGrantExpiry(session *OfflineGrant) (expiry time.Time, err error) {
	var clientConfig *config.OAuthClientConfig
	for _, c := range s.OAuthConfig.Clients {
		if c.ClientID == session.ClientID {
			cc := c
			clientConfig = &cc
		}
	}

	if clientConfig == nil {
		err = ErrGrantNotFound
		return
	}

	expiry = s.computeOfflineGrantExpiryWithClient(session, clientConfig)
	return
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
