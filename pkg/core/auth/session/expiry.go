package session

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func computeSessionStorageExpiry(session *auth.Session, config config.APIClientConfiguration) (expiry time.Time) {
	if config.RefreshTokenDisabled {
		expiry = session.AccessTokenCreatedAt.Add(time.Second * time.Duration(config.AccessTokenLifetime))
		if config.SessionIdleTimeoutEnabled {
			sessionIdleExpiry := session.AccessedAt.Add(time.Second * time.Duration(config.SessionIdleTimeout))
			if sessionIdleExpiry.Before(expiry) {
				expiry = sessionIdleExpiry
			}
		}
	} else {
		expiry = session.CreatedAt.Add(time.Second * time.Duration(config.RefreshTokenLifetime))
		if config.SessionIdleTimeoutEnabled {
			sessionIdleExpiry := session.AccessTokenCreatedAt.Add(time.Second * time.Duration(config.SessionIdleTimeout))
			if sessionIdleExpiry.Before(expiry) {
				expiry = sessionIdleExpiry
			}
		}
	}
	return
}

func checkSessionExpired(session *auth.Session, now time.Time, config config.APIClientConfiguration, kind auth.SessionTokenKind) (expired bool) {
	// treat refresh token as expired if disabled
	if kind == auth.SessionTokenKindRefreshToken && config.RefreshTokenDisabled {
		expired = true
		return
	}

	switch kind {
	case auth.SessionTokenKindAccessToken:
		accessTokenExpiry := session.AccessTokenCreatedAt.Add(time.Second * time.Duration(config.AccessTokenLifetime))
		if now.After(accessTokenExpiry) {
			expired = true
			return
		}

		if config.SessionIdleTimeoutEnabled && config.RefreshTokenDisabled {
			accessTokenIdleExpiry := session.AccessedAt.Add(time.Second * time.Duration(config.SessionIdleTimeout))
			if now.After(accessTokenIdleExpiry) {
				expired = true
				return
			}
		}

		if config.RefreshTokenDisabled {
			return
		}
		fallthrough // if refresh token is expired, treat access token as expired too

	case auth.SessionTokenKindRefreshToken:
		refreshTokenExpiry := session.CreatedAt.Add(time.Second * time.Duration(config.RefreshTokenLifetime))
		if now.After(refreshTokenExpiry) {
			expired = true
			return
		}

		if config.SessionIdleTimeoutEnabled {
			refreshTokenIdleExpiry := session.AccessTokenCreatedAt.Add(time.Second * time.Duration(config.SessionIdleTimeout))
			if now.After(refreshTokenIdleExpiry) {
				expired = true
				return
			}
		}
	}

	return false
}
