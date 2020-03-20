package redis

import "fmt"

func codeGrantKey(appID string, codeHash string) string {
	return fmt.Sprintf("%s:code-grant:%s", appID, codeHash)
}
