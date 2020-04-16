package sso

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/skygeario/skygear-server/pkg/core/errors"
)

type StateClaims struct {
	State
	jwt.StandardClaims
}

func makeStandardClaims(audience string) jwt.StandardClaims {
	return jwt.StandardClaims{
		Audience:  audience,
		ExpiresAt: time.Now().UTC().Add(5 * time.Minute).Unix(),
	}
}

func isValidStandardClaims(audience string, claims jwt.StandardClaims) bool {
	err := claims.Valid()
	if err != nil {
		return false
	}
	ok := claims.VerifyAudience(audience, true)
	if !ok {
		return false
	}
	return true
}

// EncodeState encodes state by JWT
func EncodeState(secret string, audience string, state State) (string, error) {
	claims := StateClaims{
		state,
		makeStandardClaims(audience),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// DecodeState decodes state by JWT
func DecodeState(secret string, audience string, encoded string) (*State, error) {
	claims := StateClaims{}
	_, err := jwt.ParseWithClaims(encoded, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected JWT alg")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, NewSSOFailed(InvalidParams, "invalid sso state")
	}
	ok := isValidStandardClaims(audience, claims.StandardClaims)
	if !ok {
		return nil, NewSSOFailed(InvalidParams, "invalid sso state")
	}
	return &claims.State, nil
}
