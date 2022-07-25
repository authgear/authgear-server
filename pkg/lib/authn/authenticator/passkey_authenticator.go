package authenticator

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type Passkey struct {
	ID                  string                         `json:"id"`
	UserID              string                         `json:"user_id"`
	CreatedAt           time.Time                      `json:"created_at"`
	UpdatedAt           time.Time                      `json:"updated_at"`
	Kind                string                         `json:"kind"`
	IsDefault           bool                           `json:"is_default"`
	CredentialID        string                         `json:"credential_id"`
	CreationOptions     *model.WebAuthnCreationOptions `json:"creation_options,omitempty"`
	AttestationResponse []byte                         `json:"attestation_response,omitempty"`
	// SignCount of 0 means sign count is not supported by the authenticator.
	// So we do not include omitempty here.
	SignCount int64 `json:"sign_count"`
}

func (a *Passkey) ToInfo() *Info {
	return &Info{
		ID:        a.ID,
		UserID:    a.UserID,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
		Type:      model.AuthenticatorTypePasskey,
		Kind:      Kind(a.Kind),
		IsDefault: a.IsDefault,

		Passkey: a,
	}
}
