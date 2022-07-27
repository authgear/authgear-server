package identity

import (
	"github.com/authgear/authgear-server/pkg/lib/webauthn"
)

type PasskeySpec struct {
	AttestationResponse []byte                    `json:"attestation_response,omitempty"`
	AssertionResponse   []byte                    `json:"assertion_response,omitempty"`
	CreationOptions     *webauthn.CreationOptions `json:"creation_options,omitempty"`
}
