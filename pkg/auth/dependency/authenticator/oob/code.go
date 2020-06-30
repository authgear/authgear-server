package oob

import (
	"crypto/subtle"
	"time"

	"github.com/authgear/authgear-server/pkg/core/rand"
)

const (
	oobAlphabet = "0123456789"
	// TODO(interaction): configurable OOB code length
	OOBCodeLength = 4
)

const (
	// OOBCodeValidDuration is 20 minutes according to the suggestion in
	// https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html#step-3-send-a-token-over-a-side-channel
	OOBCodeValidDuration time.Duration = 20 * time.Minute
	// OOBCodeSendCooldownSeconds is 60 seconds.
	OOBCodeSendCooldownSeconds = 60
)

func GenerateCode() string {
	code := rand.StringWithAlphabet(OOBCodeLength, oobAlphabet, rand.SecureRand)
	return code
}

func VerifyCode(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
