package session

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

func computeSessionStorageExpiry(session *IDPSession, cfg config.SessionConfiguration) (expiry time.Time) {
	expiry = session.CreatedAt.Add(time.Second * time.Duration(cfg.Lifetime))
	if cfg.IdleTimeoutEnabled {
		sessionIdleExpiry := session.AccessedAt.Add(time.Second * time.Duration(cfg.IdleTimeout))
		if sessionIdleExpiry.Before(expiry) {
			expiry = sessionIdleExpiry
		}
	}
	return
}

func checkSessionExpired(session *IDPSession, now time.Time, cfg config.SessionConfiguration) (expired bool) {
	sessionExpiry := session.CreatedAt.Add(time.Second * time.Duration(cfg.Lifetime))
	if now.After(sessionExpiry) {
		expired = true
		return
	}

	if cfg.IdleTimeoutEnabled {
		sessionIdleExpiry := session.AccessedAt.Add(time.Second * time.Duration(cfg.IdleTimeout))
		if now.After(sessionIdleExpiry) {
			expired = true
			return
		}
	}

	return false
}
