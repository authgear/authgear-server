package oauthsession

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

type T struct {
	AuthorizationRequest protocol.AuthorizationRequest `json:"authorization_request,omitempty"`
	SettingsActionID     string                        `json:"settings_action_id,omitempty"`
	SettingsActionResult *SettingsActionResult         `json:"settings_action_result,omitempty"`

	// IdentitySpecs is for supporting include_identity_attributes_in_id_token.
	IdentitySpecs []*identity.Spec `json:"identity_specs,omitzero"`
}

type Entry struct {
	ID string `json:"id,omitempty"`
	T  T      `json:"t,omitempty"`
}

func NewEntry(t T) *Entry {
	id := rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)
	return &Entry{
		ID: id,
		T:  t,
	}
}
