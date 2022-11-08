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
	ComputeSessionExpiry(session *idpsession.IDPSession) time.Time
}

type OfflineGrantService struct {
	OAuthConfig *config.OAuthConfig
	IDPSessions ServiceIDPSessionProvider
	Clock       clock.Clock
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

	expiry, err = s.ComputeOfflineGrantExpiryWithClient(session, clientConfig)
	return
}

func (s *OfflineGrantService) ComputeOfflineGrantExpiryWithClient(session *OfflineGrant, cfg *config.OAuthClientConfig) (expiry time.Time, err error) {
	expiry = session.CreatedAt.Add(cfg.RefreshTokenLifetime.Duration())
	if *cfg.RefreshTokenIdleTimeoutEnabled {
		idleExpiry := session.AccessInfo.LastAccess.Timestamp.Add(cfg.RefreshTokenIdleTimeout.Duration())
		if idleExpiry.Before(expiry) {
			expiry = idleExpiry
		}
	}

	if session.SSOEnabled {
		if session.IDPSessionID == "" {
			// expire sso enabled refresh token immediately if idp session is not found
			expiry = s.Clock.NowUTC()
			return
		}

		idp, e := s.IDPSessions.Get(session.IDPSessionID)
		if e != nil {
			if errors.Is(e, idpsession.ErrSessionNotFound) {
				// expire sso enabled refresh token immediately if idp session is not found
				expiry = s.Clock.NowUTC()
				return
			}
			err = e
			return
		}

		idpSessionExpiry := s.IDPSessions.ComputeSessionExpiry(idp)
		if idpSessionExpiry.Before(expiry) {
			expiry = idpSessionExpiry
		}
	}
	return
}
