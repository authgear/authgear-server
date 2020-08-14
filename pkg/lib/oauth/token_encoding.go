package oauth

import (
	"errors"
	"fmt"
	"strings"
)

func EncodeAccessToken(token string) string {
	return token
}

func EncodeRefreshToken(token string, grantID string) string {
	return fmt.Sprintf("%s.%s", grantID, token)
}

func DecodeAccessToken(encodedToken string) (token string, err error) {
	return encodedToken, nil
}

func DecodeRefreshToken(encodedToken string) (token string, grantID string, err error) {
	parts := strings.SplitN(encodedToken, ".", 2)
	if len(parts) != 2 {
		return "", "", errors.New("invalid refresh token")
	}

	return parts[1], parts[0], nil
}
