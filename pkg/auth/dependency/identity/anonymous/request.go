package anonymous

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
)

type RequestAction string

const (
	RequestActionAuth    RequestAction = "auth"
	RequestActionPromote RequestAction = "promote"
)

type Request struct {
	jwt.StandardClaims
	Key       jwk.Key       `json:"-"`
	Challenge string        `json:"challenge"`
	Action    RequestAction `json:"action"`
}
