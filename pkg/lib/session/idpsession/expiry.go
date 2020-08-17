package idpsession

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func computeSessionStorageExpiry(session *IDPSession, cfg *config.SessionConfig) (expiry time.Time) {
	expiry = session.CreatedAt.Add(cfg.Lifetime.Duration())
	if cfg.IdleTimeoutEnabled {
		sessionIdleExpiry := session.AccessInfo.LastAccess.Timestamp.Add(time.Second * time.Duration(cfg.IdleTimeout))
		if sessionIdleExpiry.Before(expiry) {
			expiry = sessionIdleExpiry
		}
	}
	return
}

func checkSessionExpired(session *IDPSession, now time.Time, cfg *config.SessionConfig) (expired bool) {
	sessionExpiry := session.CreatedAt.Add(cfg.Lifetime.Duration())
	if now.After(sessionExpiry) {
		expired = true
		return
	}

	if cfg.IdleTimeoutEnabled {
		sessionIdleExpiry := session.AccessInfo.LastAccess.Timestamp.Add(cfg.IdleTimeout.Duration())
		if now.After(sessionIdleExpiry) {
			expired = true
			return
		}
	}

	return false
}
