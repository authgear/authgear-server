package sso

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

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
	// AccessTokenURLs is a map of request access token url
	AccessTokenURLs = map[string]string{
		"google":    "https://www.googleapis.com/oauth2/v4/token",
		"facebook":  "https://graph.facebook.com/v2.10/oauth/access_token",
		"instagram": "https://api.instagram.com/oauth/access_token",
		"linkedin":  "https://www.linkedin.com/oauth/v2/accessToken",
	}
	// UserProfileURLs is a map of request ursr profile with access token
	UserProfileURLs = map[string]string{
		"google":    "https://www.googleapis.com/oauth2/v1/userinfo",
		"facebook":  "https://graph.facebook.com/v2.10/me",
		"instagram": "https://api.instagram.com/v1/users/self",
		"linkedin":  "https://www.linkedin.com/v1/people/~?format=json",
	}
)

// CustomClaims is the type for jwt encoded
type CustomClaims struct {
	State
	jwt.StandardClaims
}

// BaseURL returns base URL by provider name
func BaseURL(providerName string) (u string) {
	u = BaseURLs[providerName]
	return
}

// AccessTokenURL returns access token URL by provider name
func AccessTokenURL(providerName string) (u string) {
	u = AccessTokenURLs[providerName]
	return
}

// UserProfileURL returns user profile URL by provider name
func UserProfileURL(providerName string) (u string) {
	u = UserProfileURLs[providerName]
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
	claims := CustomClaims{
		state,
		jwt.StandardClaims{},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// DecodeState decodes state by JWT
func DecodeState(secret string, encoded string) (State, error) {
	claims := CustomClaims{}
	_, err := jwt.ParseWithClaims(encoded, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("fails to parse token")
		}
		return []byte(secret), nil
	})
	return claims.State, err
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
	orgPath := strings.TrimRight(u.Path, "/")
	path := fmt.Sprintf("%s/sso/%s/auth_handler", orgPath, providerName)
	u.Path = path
	return u.String()
}
