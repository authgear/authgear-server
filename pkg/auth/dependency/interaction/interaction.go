package interaction

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// Interaction represents an interaction with authenticators/identities, and authentication process.
type Interaction struct {
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
	ClientID  string    `json:"client_id,omitempty"`

	Intent Intent           `json:"intent"`
	Error  *skyerr.APIError `json:"error,omitempty"`

	UserID                   string             `json:"user_id"`
	IdentityID               string             `json:"identity_id,omitempty"`
	Identity                 *IdentityInfo      `json:"-"`
	PrimaryAuthenticatorID   string             `json:"primary_authenticator_id,omitempty"`
	PrimaryAuthenticator     *AuthenticatorInfo `json:"-"`
	SecondaryAuthenticatorID string             `json:"secondary_authenticator_id,omitempty"`
	SecondaryAuthenticator   *AuthenticatorInfo `json:"-"`

	PendingIdentity      *IdentityInfo       `json:"pending_identity,omitempty"`
	PendingAuthenticator *AuthenticatorInfo  `json:"pending_authenticator,omitempty"`
	NewIdentiies         []IdentityInfo      `json:"new_identities,omitempty"`
	NewAuthenticators    []AuthenticatorInfo `json:"new_authenticators,omitempty"`
}
