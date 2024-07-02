package redis

import "fmt"

func codeGrantKey(appID, codeHash string) string {
	return fmt.Sprintf("app:%s:code-grant:%s", appID, codeHash)
}

func settingsActionGrantKey(appID, codeHash string) string {
	return fmt.Sprintf("app:%s:settings-action-grant:%s", appID, codeHash)
}

func accessGrantKey(appID, tokenHash string) string {
	return fmt.Sprintf("app:%s:access-grant:%s", appID, tokenHash)
}

func offlineGrantKey(appID, id string) string {
	return fmt.Sprintf("app:%s:offline-grant:%s", appID, id)
}

func offlineGrantMutexName(appID, id string) string {
	return fmt.Sprintf("app:%s:offline-grant-mutex:%s", appID, id)
}

func offlineGrantListKey(appID, userID string) string {
	return fmt.Sprintf("app:%s:offline-grant-list:%s", appID, userID)
}

func appSessionTokenKey(appID string, tokenHash string) string {
	return fmt.Sprintf("app:%s:app-session-token:%s", appID, tokenHash)
}

func appSessionKey(appID string, tokenHash string) string {
	return fmt.Sprintf("app:%s:app-session:%s", appID, tokenHash)
}

func appInitiatedSSOToWebTokenKey(appID string, tokenHash string) string {
	return fmt.Sprintf("app:%s:app-initiated-sso-to-web-token:%s", appID, tokenHash)
}
