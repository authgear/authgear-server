package customtoken

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
)

type SSOCustomTokenClaims struct {
	RawProfile map[string]interface{} `json:"skyprofile"`
	jwt.StandardClaims
}
