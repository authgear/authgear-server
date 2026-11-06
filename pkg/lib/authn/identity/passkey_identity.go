package identity

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/passkeyprovider"
)

type Passkey struct {
	ID                  string                         `json:"id"`
	CreatedAt           time.Time                      `json:"created_at"`
	UpdatedAt           time.Time                      `json:"updated_at"`
	UserID              string                         `json:"user_id"`
	CredentialID        string                         `json:"credential_id"`
	CreationOptions     *model.WebAuthnCreationOptions `json:"creation_options,omitempty"`
	AttestationResponse []byte                         `json:"attestation_response,omitempty"`
	// AAGUID is read from AttestationResponse, and is not stored on its own.
	AAGUID string `json:"aaguid,omitempty"`
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

// ProviderName returns the name of the provider that stores the passkey,
// or an empty string when it is unknown.
func (i *Passkey) ProviderName() string {
	name, _ := passkeyprovider.Name(i.AAGUID)
	return name
}
