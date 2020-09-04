package mfa

import (
	"time"

	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

const (
	deviceTokenLength = 64
)

type DeviceToken struct {
	UserID    string    `json:"-"`
	Token     string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
}

func GenerateDeviceToken() string {
	code := rand.StringWithAlphabet(deviceTokenLength, base32.Alphabet, rand.SecureRand)
	return code
}
