package interaction

import (
	"encoding/json"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// Interaction represents an interaction with authenticators/identities, and authentication process.
type Interaction struct {
	Token       string    `json:"token"`
	CreatedAt   time.Time `json:"created_at"`
	ExpireAt    time.Time `json:"expire_at"`
	SessionID   string    `json:"session_id,omitempty"`
	SessionType string    `json:"session_type,omitempty"`
	ClientID    string    `json:"client_id,omitempty"`

	Intent Intent           `json:"-"`
	Error  *skyerr.APIError `json:"error,omitempty"`

	UserID                 string             `json:"user_id"`
	Identity               *identity.Ref      `json:"identity"`
	PrimaryAuthenticator   *authenticator.Ref `json:"primary_authenticator"`
	SecondaryAuthenticator *authenticator.Ref `json:"secondary_authenticator"`

	State                map[string]string     `json:"state,omitempty"`
	PendingAuthenticator *authenticator.Info   `json:"pending_authenticator,omitempty"`
	UpdateIdentities     []*identity.Info      `json:"update_identities,omitempty"`
	UpdateAuthenticators []*authenticator.Info `json:"update_authenticators,omitempty"`
	NewIdentities        []*identity.Info      `json:"new_identities,omitempty"`
	NewAuthenticators    []*authenticator.Info `json:"new_authenticators,omitempty"`
	RemoveIdentities     []*identity.Info      `json:"remove_identities,omitempty"`
	RemoveAuthenticators []*authenticator.Info `json:"remove_authenticators,omitempty"`

	// Extra is used to persist extra data across the interaction.
	Extra map[string]string `json:"extra,omitempty"`

	// The following fields are for checking programming errors.
	saved     bool
	committed bool
}

func newInteraction(clientID string, intent Intent) *Interaction {
	return &Interaction{
		ClientID: clientID,
		Intent:   intent,
		State:    map[string]string{},
		Extra:    map[string]string{},
	}
}

func (i *Interaction) IsNewIdentity(id string) bool {
	for _, identity := range i.NewIdentities {
		if identity.ID == id {
			return true
		}
	}
	return false
}

func (i *Interaction) IsNewAuthenticator(id string) bool {
	for _, authenticator := range i.NewAuthenticators {
		if authenticator.ID == id {
			return true
		}
	}
	return false
}

func (i *Interaction) MarshalJSON() ([]byte, error) {
	type interaction Interaction
	type jsonInteraction struct {
		*interaction
		Intent     Intent     `json:"intent"`
		IntentType IntentType `json:"intent_type"`
	}
	ji := jsonInteraction{
		interaction: (*interaction)(i),
		Intent:      i.Intent,
		IntentType:  i.Intent.Type(),
	}
	return json.Marshal(ji)
}

func (i *Interaction) UnmarshalJSON(data []byte) error {
	type interaction Interaction
	type jsonInteraction struct {
		*interaction
		Intent     json.RawMessage `json:"intent"`
		IntentType IntentType      `json:"intent_type"`
	}
	ji := &jsonInteraction{interaction: (*interaction)(i)}
	if err := json.Unmarshal(data, ji); err != nil {
		return err
	}

	i.Intent = NewIntent(ji.IntentType)
	if err := json.Unmarshal(ji.Intent, i.Intent); err != nil {
		return err
	}

	return nil
}
