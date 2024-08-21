package idpsession

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func setSessionExpireAtForResolvedSession(session *IDPSession, cfg *config.SessionConfig) {
	session.ExpireAtForResolvedSession = session.CreatedAt.Add(cfg.Lifetime.Duration())
	if *cfg.IdleTimeoutEnabled {
		sessionIdleExpiry := session.AccessInfo.LastAccess.Timestamp.Add(cfg.IdleTimeout.Duration())
		if sessionIdleExpiry.Before(session.ExpireAtForResolvedSession) {
			session.ExpireAtForResolvedSession = sessionIdleExpiry
		}
	}
	return
}
