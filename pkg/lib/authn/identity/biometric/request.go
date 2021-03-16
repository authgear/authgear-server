package biometric

import (
	"github.com/lestrrat-go/jwx/jwk"
)

// nolint:gosec
const RequestTokenType = "vnd.authgear.biometric-request"

type RequestAction string

const (
	RequestActionSetup        RequestAction = "setup"
	RequestActionAuthenticate RequestAction = "authenticate"
)

type Request struct {
	Key        jwk.Key                `json:"-"`
	KeyID      string                 `json:"-"`
	DeviceInfo map[string]interface{} `json:"device_info"`
	Challenge  string                 `json:"challenge"`
	Action     RequestAction          `json:"action"`
}
