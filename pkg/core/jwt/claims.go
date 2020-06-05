package jwt

import (
	"github.com/dgrijalva/jwt-go"
)

type StandardClaims jwt.StandardClaims

func (c StandardClaims) Valid() error {
	copy := jwt.StandardClaims(c)
	// iat should not be validated.
	copy.IssuedAt = 0
	return copy.Valid()
}

type MapClaims jwt.MapClaims

func (c MapClaims) Valid() error {
	copy := jwt.MapClaims{}
	for k, v := range c {
		// iat should not be validated.
		if k == "iat" {
			continue
		}
		copy[k] = v
	}
	return copy.Valid()
}

func (c MapClaims) VerifyAudience(cmp string, req bool) bool {
	return jwt.MapClaims(c).VerifyAudience(cmp, req)
}
