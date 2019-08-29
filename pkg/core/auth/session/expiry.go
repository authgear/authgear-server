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
