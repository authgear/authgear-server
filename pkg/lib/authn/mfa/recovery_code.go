package mfa

import (
	"time"

	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

const (
	recoveryCodeLength = 10
)

type RecoveryCode struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Consumed  bool      `json:"consumed"`
}

func GenerateRecoveryCode() string {
	code := rand.StringWithAlphabet(recoveryCodeLength, base32.Alphabet, rand.SecureRand)
	return code
}
