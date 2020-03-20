package oauth

import "fmt"

func EncodeAccessToken(token string) string {
	return token
}

func EncodeRefreshToken(token string, grantID string) string {
	return fmt.Sprintf("%s.%s", grantID, token)
}
