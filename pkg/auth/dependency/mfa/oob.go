package mfa

import (
	"github.com/skygeario/skygear-server/pkg/core/rand"
)

const (
	oobAlphabet   = "0123456789"
	oobCodeLength = 6
)

func GenerateRandomOOBCode() string {
	code := rand.StringWithAlphabet(oobCodeLength, oobAlphabet, rand.SecureRand)
	return code
}
