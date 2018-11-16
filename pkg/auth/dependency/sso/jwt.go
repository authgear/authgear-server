package sso

import (
	"github.com/dgrijalva/jwt-go"
)

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
