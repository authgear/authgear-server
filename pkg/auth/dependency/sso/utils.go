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

// CustomCliams is the type for jwt encoded
type CustomCliams struct {
	State
	jwt.StandardClaims
}

// BaseURL returns base URL by provider name
func BaseURL(providerName string) (u string) {
	u = BaseURLs[providerName]
	return
}

// NewState constructs a new state
func NewState(params GetURLParams) State {
	return State{
		UXMode:      params.UXMode.String(),
		CallbackURL: params.CallbackURL,
		Action:      params.Action,
		UserID:      params.UserID,
	}
}

// EncodeState encodes state by JWT
func EncodeState(secret string, state State) (string, error) {
	claims := CustomCliams{
		state,
		jwt.StandardClaims{},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GetScope returns parameter scope or default scope
func GetScope(scope Scope, defaultScope Scope) Scope {
	if len(scope) != 0 {
		return scope
	}
	return defaultScope
}

// RedirectURI generates redirect uri from URLPrefix and provider name
func RedirectURI(URLPrefix string, providerName string) string {
	u, _ := url.Parse(URLPrefix)
	path := fmt.Sprintf("/sso/%s/auth_handler", providerName)
	u.Path = path
	return u.String()
}
