package oauth

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type OfflineGrantService struct {
	OAuthConfig *config.OAuthConfig
}

func (s *OfflineGrantService) ComputeOfflineGrantExpiryWithClients(session *OfflineGrant) (expiry time.Time, err error) {
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

	expiry = s.ComputeOfflineGrantExpiryWithClient(session, clientConfig)
	return
}

func (s *OfflineGrantService) ComputeOfflineGrantExpiryWithClient(session *OfflineGrant, cfg *config.OAuthClientConfig) (expiry time.Time) {
	expiry = session.CreatedAt.Add(cfg.RefreshTokenLifetime.Duration())
	if *cfg.RefreshTokenIdleTimeoutEnabled {
		idleExpiry := session.AccessInfo.LastAccess.Timestamp.Add(cfg.RefreshTokenIdleTimeout.Duration())
		if idleExpiry.Before(expiry) {
			expiry = idleExpiry
		}
	}
	return
}
