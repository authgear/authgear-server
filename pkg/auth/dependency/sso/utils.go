package sso

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

// CustomClaims is the type for jwt encoded
type CustomClaims struct {
	State
	jwt.StandardClaims
}

// NewState constructs a new state
func NewState(params GetURLParams) State {
	return params.State
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
func DecodeState(secret string, encoded string) (*State, error) {
	claims := CustomClaims{}
	_, err := jwt.ParseWithClaims(encoded, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("fails to parse token")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	return &claims.State, nil
}

// RedirectURI generates redirect uri from URLPrefix and provider name
func RedirectURI(oauthConfig config.OAuthConfiguration, providerConfig config.OAuthProviderConfiguration) string {
	u, _ := url.Parse(oauthConfig.URLPrefix)
	orgPath := strings.TrimRight(u.Path, "/")
	path := fmt.Sprintf("%s/sso/%s/auth_handler", orgPath, providerConfig.ID)
	u.Path = path
	return u.String()
}
