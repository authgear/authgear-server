package mfa

import (
	"github.com/skygeario/skygear-server/pkg/core/base32"
	"github.com/skygeario/skygear-server/pkg/core/rand"
)

const (
	recoveryCodeLength = 10
)

func GenerateRandomRecoveryCode() string {
	code := rand.StringWithAlphabet(recoveryCodeLength, base32.Alphabet, rand.SecureRand)
	return code
}
