package redis

import "fmt"

func sessionKey(appID string, sessionID string) string {
	return fmt.Sprintf("%s:session:%s", appID, sessionID)
}
