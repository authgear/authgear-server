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
	OAuthConfig *config.OAuthConfig
	Clock       clock.Clock
	IDPSessions ServiceIDPSessionProvider
}

func (s *OfflineGrantService) IsValid(session *OfflineGrant) (bool, time.Time, error) {
	now := s.Clock.NowUTC()
	expiry, err := s.ComputeOfflineGrantExpiry(session)
	if errors.Is(err, ErrGrantNotFound) {
		return false, now, nil
	} else if err != nil {
		return false, time.Time{}, err
	}

	offlineGrantIsValid := now.Before(expiry)
	if !offlineGrantIsValid {
		return false, expiry, nil
	}

	if session.SSOEnabled {
		if session.IDPSessionID == "" {
			return false, now, nil
		}

		idp, err := s.IDPSessions.Get(session.IDPSessionID)
		if err != nil {
			if errors.Is(err, idpsession.ErrSessionNotFound) {
				return false, now, nil
			}
			return false, time.Time{}, err
		}

		idpSessionExpired := s.IDPSessions.CheckSessionExpired(idp)
		if idpSessionExpired {
			return false, now, nil
		}
	}

	return true, expiry, nil
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
