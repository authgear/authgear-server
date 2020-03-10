package session

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

func computeSessionStorageExpiry(session *Session, config config.APIClientConfiguration) (expiry time.Time) {
	// FIXME(session): use session lifetime here
	expiry = session.CreatedAt.Add(time.Second * time.Duration(config.RefreshTokenLifetime))
	if config.SessionIdleTimeoutEnabled {
		sessionIdleExpiry := session.AccessedAt.Add(time.Second * time.Duration(config.SessionIdleTimeout))
		if sessionIdleExpiry.Before(expiry) {
			expiry = sessionIdleExpiry
		}
	}
	return
}

func checkSessionExpired(session *Session, now time.Time, config config.APIClientConfiguration) (expired bool) {
	// FIXME(session): use session lifetime here
	sessionExpiry := session.CreatedAt.Add(time.Second * time.Duration(config.RefreshTokenLifetime))
	if now.After(sessionExpiry) {
		expired = true
		return
	}

	if config.SessionIdleTimeoutEnabled {
		sessionIdleExpiry := session.AccessedAt.Add(time.Second * time.Duration(config.SessionIdleTimeout))
		if now.After(sessionIdleExpiry) {
			expired = true
			return
		}
	}

	return false
}
