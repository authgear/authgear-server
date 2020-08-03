package mfa

import (
	"crypto/subtle"

	"github.com/authgear/authgear-server/pkg/core/base32"
	"github.com/authgear/authgear-server/pkg/core/rand"
)

const (
	deviceTokenLength  = 64
	recoveryCodeLength = 10
)

func GenerateDeviceToken() string {
	code := rand.StringWithAlphabet(deviceTokenLength, base32.Alphabet, rand.SecureRand)
	return code
}

func GenerateRecoveryCode() string {
	code := rand.StringWithAlphabet(recoveryCodeLength, base32.Alphabet, rand.SecureRand)
	return code
}

func VerifyToken(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
