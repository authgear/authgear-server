package authn

import (
	"errors"

	"github.com/dgrijalva/jwt-go"
)

type sessionToken struct {
	jwt.StandardClaims
	Session `json:"authn_session"`
}

func encodeSessionToken(secret string, claims sessionToken) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func decodeSessionToken(secret string, tokenString string) (*sessionToken, error) {
	t, err := jwt.ParseWithClaims(tokenString, &sessionToken{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected JWT alg")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, errInvalidToken
	}
	claims, ok := t.Claims.(*sessionToken)
	if !ok {
		return nil, errInvalidToken
	}
	if !t.Valid {
		return nil, errInvalidToken
	}
	return claims, nil
}
