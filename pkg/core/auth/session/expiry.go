package session

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func computeSessionExpiry(session *auth.Session, config config.APIClientConfiguration) (expiry time.Time) {
	expiry = session.AccessTokenCreatedAt.Add(time.Second * time.Duration(config.AccessTokenLifetime))
	if config.RefreshTokenDisabled {
		if config.SessionIdleTimeoutEnabled {
			sessionIdleExpiry := session.AccessedAt.Add(time.Second * time.Duration(config.SessionIdleTimeout))
			if sessionIdleExpiry.Before(expiry) {
				expiry = sessionIdleExpiry
			}
		}
	} else {
		// TODO(session): refresh token handling
		if config.SessionIdleTimeoutEnabled {
			sessionIdleExpiry := session.AccessedAt.Add(time.Second * time.Duration(config.SessionIdleTimeout))
			if sessionIdleExpiry.Before(expiry) {
				expiry = sessionIdleExpiry
			}
		}
	}
	return
}
