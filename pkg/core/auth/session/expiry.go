package session

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func computeSessionStorageExpiry(session *auth.Session, config config.OAuthClientConfiguration) (expiry time.Time) {
	expiry = session.CreatedAt.Add(time.Second * time.Duration(config.RefreshTokenLifetime()))
	return
}

func checkSessionExpired(session *auth.Session, now time.Time, config config.OAuthClientConfiguration, kind auth.SessionTokenKind) (expired bool) {
	refreshTokenExpiry := session.CreatedAt.Add(time.Second * time.Duration(config.RefreshTokenLifetime()))
	if now.After(refreshTokenExpiry) {
		expired = true
		return
	}
	return false
}
