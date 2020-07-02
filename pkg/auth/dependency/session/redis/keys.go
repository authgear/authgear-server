package redis

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/auth/config"
)

func sessionKey(appID config.AppID, sessionID string) string {
	return fmt.Sprintf("%s:session:%s", appID, sessionID)
}

func sessionListKey(appID config.AppID, userID string) string {
	return fmt.Sprintf("%s:session-list:%s", appID, userID)
}
