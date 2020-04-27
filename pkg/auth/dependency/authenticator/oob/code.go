package oob

import (
	"crypto/subtle"

	"github.com/skygeario/skygear-server/pkg/core/rand"
)

const (
	oobAlphabet = "0123456789"
	// TODO(interaction): configurable OOB code length
	oobCodeLength = 4
)

func GenerateCode() string {
	code := rand.StringWithAlphabet(oobCodeLength, oobAlphabet, rand.SecureRand)
	return code
}

func VerifyCode(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
