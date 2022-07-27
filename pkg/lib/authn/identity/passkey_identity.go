package identity

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/webauthn"
)

type Passkey struct {
	ID                  string                    `json:"id"`
	CreatedAt           time.Time                 `json:"created_at"`
	UpdatedAt           time.Time                 `json:"updated_at"`
	UserID              string                    `json:"user_id"`
	CredentialID        string                    `json:"credential_id"`
	CreationOptions     *webauthn.CreationOptions `json:"creation_options,omitempty"`
	AttestationResponse []byte                    `json:"attestation_response,omitempty"`
}

func (i *Passkey) ToInfo() *Info {
	return &Info{
		ID:        i.ID,
		UserID:    i.UserID,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
		Type:      model.IdentityTypePasskey,

		Passkey: i,
	}
}
