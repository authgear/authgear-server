package customtoken

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
)

type SSOCustomTokenClaims struct {
	jwt.StandardClaims
	Email string `json:"email,omitempty"`
}

func (c *SSOCustomTokenClaims) Validate(issuer string, audience string) error {
	if c.Subject == "" {
		return errors.New("invalid token: subject (sub) not specified")
	}

	if c.ExpiresAt == 0 {
		return errors.New("invalid token: expires at (exp) not specified")
	}

	if c.IssuedAt == 0 {
		return errors.New("invalid token: issued at (iat) not specified")
	}

	if c.Issuer != issuer {
		return errors.New("invalid token: issuer not matched")
	}

	if audience != "" && c.Audience != "" && c.Audience != audience {
		return errors.New("invalid token: audience not matched")
	}

	if c.Valid() != nil {
		return errors.New("invalid token: token is not valid at this time")
	}

	return nil
}
