package redis

import "fmt"

func codeGrantKey(appID, codeHash string) string {
	return fmt.Sprintf("%s:code-grant:%s", appID, codeHash)
}

func accessGrantKey(appID, tokenHash string) string {
	return fmt.Sprintf("%s:access-grant:%s", appID, tokenHash)
}

func offlineGrantKey(appID, id string) string {
	return fmt.Sprintf("%s:offline-grant:%s", appID, id)
}

func offlineGrantListKey(appID, userID string) string {
	return fmt.Sprintf("%s:offline-grant-list:%s", appID, userID)
}
