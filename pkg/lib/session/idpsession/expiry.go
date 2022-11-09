package idpsession

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func computeSessionStorageExpiry(session *IDPSession, cfg *config.SessionConfig) (expiry time.Time) {
	expiry = session.CreatedAt.Add(cfg.Lifetime.Duration())
	if *cfg.IdleTimeoutEnabled {
		sessionIdleExpiry := session.AccessInfo.LastAccess.Timestamp.Add(cfg.IdleTimeout.Duration())
		if sessionIdleExpiry.Before(expiry) {
			expiry = sessionIdleExpiry
		}
	}
	return
}
