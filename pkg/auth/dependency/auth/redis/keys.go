package redis

import "fmt"

func accessEventStreamKey(appID string, sessionID string) string {
	return fmt.Sprintf("%s:access-events:%s", appID, sessionID)
}
