package app2app

import (
	"github.com/lestrrat-go/jwx/jwk"
)

// nolint:gosec
const TokenType = "vnd.authgear.app2app-device-key"

type Token struct {
	Key       jwk.Key `json:"-"`
	Challenge string  `json:"challenge"`
}
