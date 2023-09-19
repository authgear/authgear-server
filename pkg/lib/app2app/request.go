package app2app

import (
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// nolint:gosec
const RequestTokenType = "vnd.authgear.app2app-request"

type Request struct {
	Key       jwk.Key `json:"-"`
	Challenge string  `json:"challenge"`
}
