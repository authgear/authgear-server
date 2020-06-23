package redis

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/auth/config"
)

func accessEventStreamKey(appID config.AppID, sessionID string) string {
	return fmt.Sprintf("%s:access-events:%s", appID, sessionID)
}
