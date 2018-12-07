package customtoken

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
)

type SSOCustomTokenClaims struct {
	RawProfile map[string]interface{} `json:"skyprofile"`
	jwt.StandardClaims
}

func (c *SSOCustomTokenClaims) Validate() error {
	if c.Subject == "" {
		return errors.New("invalid token: subject (sub) not specified")
	}

	if c.ExpiresAt == 0 {
		return errors.New("invalid token: expires at (exp) not specified")
	}

	if c.IssuedAt == 0 {
		return errors.New("invalid token: issued at (iat) not specified")
	}

	if c.Valid() != nil {
		return errors.New("invalid token: token is not valid at this time")
	}

	return nil
}
