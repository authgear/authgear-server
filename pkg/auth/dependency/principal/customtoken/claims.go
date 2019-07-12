package customtoken

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
)

type SSOCustomTokenClaims jwt.MapClaims

func (c SSOCustomTokenClaims) Subject() string {
	sub, ok := c["sub"].(string)
	if !ok {
		return ""
	}
	return sub
}

func (c SSOCustomTokenClaims) Email() string {
	sub, ok := c["email"].(string)
	if !ok {
		return ""
	}
	return sub
}

func (c SSOCustomTokenClaims) Valid() error {
	return jwt.MapClaims(c).Valid()
}

func (c SSOCustomTokenClaims) Validate(issuer string, audience string) error {
	mapClaims := jwt.MapClaims(c)

	if c.Subject() == "" {
		return errors.New("invalid token: subject (sub) not specified")
	}

	if _, ok := c["exp"]; !ok {
		return errors.New("invalid token: expires at (exp) not specified")
	}

	if _, ok := c["iat"]; !ok {
		return errors.New("invalid token: issued at (iat) not specified")
	}

	if !mapClaims.VerifyIssuer(issuer, true) {
		return errors.New("invalid token: issuer not matched")
	}

	if !mapClaims.VerifyAudience(audience, false) {
		return errors.New("invalid token: audience not matched")
	}

	if c.Valid() != nil {
		return errors.New("invalid token: token is not valid at this time")
	}

	return nil
}
