package flows

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/dependency/interaction"
	corerand "github.com/authgear/authgear-server/pkg/core/rand"
	"github.com/authgear/authgear-server/pkg/core/skyerr"
)

var ErrStateNotFound = errors.New("state not found")

var (
	stateIDAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	stateIDLength   = 32
)

// State is the state of a flow of an interaction.
type State struct {
	// ID is a cryptographically random string.
	ID string `json:"id"`

	// Interaction is the interaction of this flow.
	Interaction *interaction.Interaction `json:"interaction"`

	// FIXME(webapp): Clear error correctly.
	// Error is either reset to nil or set to non-nil in every POST request.
	Error *skyerr.APIError `json:"error,omitempty"`

	// Extra is used to persist extra data across the interaction.
	Extra map[string]interface{} `json:"extra,omitempty"`
}

const (
	ExtraSSOAction      string = "sso_action"
	ExtraSSONonce       string = "sso_nonce"
	ExtraSSORedirectURI string = "sso_redirect_uri"
	ExtraUserID         string = "user_id"

	// ExtraGivenLoginID indicates the given login ID by the user. It is a string.
	ExtraGivenLoginID string = "https://authgear.com/claims/given_login_id"
)

func NewState() *State {
	return &State{
		ID:    corerand.StringWithAlphabet(stateIDLength, stateIDAlphabet, corerand.SecureRand),
		Extra: make(map[string]interface{}),
	}
}

func (s *State) SetError(err error) {
	s.Error = skyerr.AsAPIError(err)
}
