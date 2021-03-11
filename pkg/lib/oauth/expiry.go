package oauth

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func ComputeOfflineGrantExpiryWithClients(s *OfflineGrant, cfg *config.OAuthConfig) (expiry time.Time, err error) {
	var clientConfig *config.OAuthClientConfig
	for _, c := range cfg.Clients {
		if c.ClientID == s.ClientID {
			cc := c
			clientConfig = &cc
		}
	}

	if clientConfig == nil {
		err = ErrGrantNotFound
		return
	}

	expiry = ComputeOfflineGrantExpiryWithClient(s, clientConfig)
	return
}

func ComputeOfflineGrantExpiryWithClient(s *OfflineGrant, cfg *config.OAuthClientConfig) (expiry time.Time) {
	expiry = s.CreatedAt.Add(cfg.RefreshTokenLifetime.Duration())
	if *cfg.RefreshTokenIdleTimeoutEnabled {
		idleExpiry := s.AccessInfo.LastAccess.Timestamp.Add(cfg.RefreshTokenIdleTimeout.Duration())
		if idleExpiry.Before(expiry) {
			expiry = idleExpiry
		}
	}
	return
}
