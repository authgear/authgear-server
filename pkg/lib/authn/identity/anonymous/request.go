package anonymous

import (
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// nolint:gosec
const RequestTokenType = "vnd.authgear.anonymous-request"

type RequestAction string

const (
	RequestActionAuth    RequestAction = "auth"
	RequestActionPromote RequestAction = "promote"
)

type Request struct {
	Key       jwk.Key       `json:"-"`
	KeyID     string        `json:"-"`
	Challenge string        `json:"challenge"`
	Action    RequestAction `json:"action"`
}
