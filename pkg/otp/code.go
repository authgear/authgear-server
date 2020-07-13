package otp

import (
	"crypto/subtle"

	"github.com/authgear/authgear-server/pkg/core/rand"
)

const (
	oobAlphabet = "0123456789"
	// TODO(interaction): configurable OOB code length
	OOBOTPLength = 4
)

func GenerateOOBOTP() string {
	code := rand.StringWithAlphabet(OOBOTPLength, oobAlphabet, rand.SecureRand)
	return code
}

func ValidateOOBOTP(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
