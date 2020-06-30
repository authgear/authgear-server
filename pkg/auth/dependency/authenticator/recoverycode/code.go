package recoverycode

import (
	"crypto/subtle"

	"github.com/authgear/authgear-server/pkg/core/base32"
	"github.com/authgear/authgear-server/pkg/core/rand"
)

const (
	recoveryCodeLength = 10
)

func GenerateCode() string {
	code := rand.StringWithAlphabet(recoveryCodeLength, base32.Alphabet, rand.SecureRand)
	return code
}

func VerifyCode(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
