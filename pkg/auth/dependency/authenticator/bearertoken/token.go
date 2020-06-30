package bearertoken

import (
	"crypto/subtle"

	"github.com/authgear/authgear-server/pkg/core/base32"
	"github.com/authgear/authgear-server/pkg/core/rand"
)

const (
	bearerTokenLength = 64
)

func GenerateToken() string {
	code := rand.StringWithAlphabet(bearerTokenLength, base32.Alphabet, rand.SecureRand)
	return code
}

func VerifyToken(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
