package sso

import (
	"fmt"
	"net/url"

	jwt "github.com/dgrijalva/jwt-go"
)

var (
	// BaseURLs is a map of provider base url
	BaseURLs = map[string]string{
		"google":    "https://accounts.google.com/o/oauth2/v2/auth",
		"facebook":  "https://www.facebook.com/dialog/oauth",
		"instagram": "https://api.instagram.com/oauth/authorize",
		"linkedin":  "https://www.linkedin.com/oauth/v2/authorization",
	}
)

func BaseURL(providerName string) (u string) {
	u = BaseURLs[providerName]
	return
}

func ToEncodedState(secret string, params GetURLParams) (string, error) {
	state := State{
		UXMode:      params.UXMode.String(),
		CallbackURL: params.CallbackURL,
		Action:      params.Action,
		UserID:      params.UserID,
	}
	return EncodeState(secret, state)
}

// EncodeState encodes state by JWT
func EncodeState(secret string, state State) (string, error) {
	type SSOCustomCliams struct {
		State
		jwt.StandardClaims
	}
	claims := SSOCustomCliams{
		state,
		jwt.StandardClaims{},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// RedirectURI generates redirect uri from URLPrefix and provider name
func RedirectURI(URLPrefix string, providerName string) string {
	u, _ := url.Parse(URLPrefix)
	path := fmt.Sprintf("/sso/%s/auth_handler", providerName)
	u.Path = path
	return u.String()
}
