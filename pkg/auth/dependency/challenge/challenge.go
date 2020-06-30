package challenge

import (
	"time"

	"github.com/authgear/authgear-server/pkg/core/base32"
	"github.com/authgear/authgear-server/pkg/core/rand"
)

type Purpose string

const (
	PurposeAnonymousRequest Purpose = "anonymous_request"
)

func (p Purpose) IsValid() bool {
	switch p {
	case PurposeAnonymousRequest:
		return true
	}
	return false
}

func (p Purpose) ValidityPeriod() time.Duration {
	// TODO(challenge): allow customization?
	switch p {
	case PurposeAnonymousRequest:
		return time.Minute * 5
	default:
		panic("challenge: unknown purpose: " + p)
	}
}

type Challenge struct {
	Token     string    `json:"token"`
	Purpose   Purpose   `json:"purpose"`
	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
}

func GenerateChallengeToken() string {
	return rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)
}
