package redis

import "fmt"

func sessionKey(appID string, sessionID string) string {
	return fmt.Sprintf("%s:session:%s", appID, sessionID)
}

func sessionListKey(appID string, userID string) string {
	return fmt.Sprintf("%s:session-list:%s", appID, userID)
}
