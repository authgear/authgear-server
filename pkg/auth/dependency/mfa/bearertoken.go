package mfa

import (
	"github.com/skygeario/skygear-server/pkg/core/base32"
	"github.com/skygeario/skygear-server/pkg/core/rand"
)

const (
	bearerTokenLength = 64
)

func GenerateRandomBearerToken() string {
	code := rand.StringWithAlphabet(bearerTokenLength, base32.Alphabet, rand.SecureRand)
	return code
}
