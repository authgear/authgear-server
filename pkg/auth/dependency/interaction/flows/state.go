package flows

import (
	"errors"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/dependency/interaction"
	corerand "github.com/authgear/authgear-server/pkg/core/rand"
	"github.com/authgear/authgear-server/pkg/core/skyerr"
)

var ErrStateNotFound = errors.New("state not found")

var (
	stateIDAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	stateIDLength   = 32
)

// State is a particular state instance of a interaction flow.
// State is immutable in the sense that every mutation creates a new state with a different InstanceID.
type State struct {
	// FlowID is the unique ID for a flow.
	// It is a constant value through out a flow.
	// It is used to keep track of which instances belong to a particular flow.
	// When one instance is committed, any other instances sharing the same FlowID become invalid.
	FlowID string `json:"flow_id"`

	// InstanceID is a unique ID for a particular instance of a flow.
	InstanceID string `json:"instance_id"`

	// Interaction is the interaction of this
	Interaction *interaction.Interaction `json:"interaction,omitempty"`

	// Error is the error associated with this state.
	Error *skyerr.APIError `json:"error,omitempty"`

	// Extra is used to persist extra data across the interaction.
	Extra map[string]interface{} `json:"extra,omitempty"`

	// readOnly is a flag to indicate that this state was read for read-only.
	// If this state is passed to UpdateState, panic will occur.
	readOnly bool
}

const (
	// FIXME(webapp): Remove the following fields when we eagerly create interaction for OAuth.
	ExtraSSOAction string = "sso_action"
	ExtraSSONonce  string = "sso_nonce"
	ExtraSSOUserID string = "sso_user_id"

	// ExtraGivenLoginID indicates the given login ID by the user. It is a string.
	ExtraGivenLoginID string = "https://authgear.com/claims/given_login_id"

	// ExtraRedirectURI indicates the redirect URI. It is a string.
	ExtraRedirectURI string = "https://authgear.com/claims/redirect_uri"

	ExtraLoginIDKey       string = "https://authgear.com/claims/login_id_key"
	ExtraLoginIDType      string = "https://authgear.com/claims/login_id_type"
	ExtraOldLoginID       string = "https://authgear.com/claims/old_login_id"
	ExtraLoginIDInputType string = "https://authgear.com/claims/login_id_input_type"

	// ExtraAnonymousUserID indicates the interaction is for promoting the anonymous user.
	ExtraAnonymousUserID string = "https://authgear.com/claims/anonymous_user_id"
)

func NewState() *State {
	flowID := corerand.StringWithAlphabet(stateIDLength, stateIDAlphabet, corerand.SecureRand)
	instanceID := corerand.StringWithAlphabet(stateIDLength, stateIDAlphabet, corerand.SecureRand)
	return &State{
		FlowID:     flowID,
		InstanceID: instanceID,
		Extra:      make(map[string]interface{}),
	}
}

// RedirectURI returns the redirect URI associated with s.
func (s *State) RedirectURI(input *url.URL) *url.URL {
	u := *input

	q := u.Query()
	q.Set("x_sid", s.InstanceID)

	u.Scheme = ""
	u.Opaque = ""
	u.Host = ""
	u.User = nil
	u.RawQuery = q.Encode()

	return &u
}
