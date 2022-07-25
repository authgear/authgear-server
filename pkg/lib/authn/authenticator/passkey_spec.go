package authenticator

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

type PasskeySpec struct {
	AttestationResponse []byte                         `json:"attestation_response,omitempty"`
	AssertionResponse   []byte                         `json:"assertion_response,omitempty"`
	CreationOptions     *model.WebAuthnCreationOptions `json:"creation_options,omitempty"`
}
