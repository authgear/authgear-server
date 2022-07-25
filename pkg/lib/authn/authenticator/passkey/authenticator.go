package passkey

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/webauthn"
)

type Authenticator struct {
	ID                  string
	IsDefault           bool
	Kind                string
	UserID              string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	CredentialID        string
	CreationOptions     *webauthn.CreationOptions
	AttestationResponse []byte
	SignCount           int64
}
