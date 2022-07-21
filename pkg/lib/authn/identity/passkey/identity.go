package passkey

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/webauthn"
)

type Identity struct {
	ID                  string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	UserID              string
	CredentialID        string
	CreationOptions     *webauthn.CreationOptions
	AttestationResponse []byte
}
