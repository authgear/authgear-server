package webapp

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

type Intent struct {
	// StateID indicates the fixed state ID. If it is empty, one is generated.
	// It is intended for OAuth.
	StateID string
	// RedirectURI indicates the location to redirect after the interaction finishes.
	RedirectURI string
	// KeepState indicates whether the state should be kept after the interaction finishes.
	// It is useful for interaction that has a dead end, such as forgot / reset password.
	// If it is true, then the state is attached to RedirectURI.
	KeepState bool
	Intent    newinteraction.Intent
	// StateExtra indicates the extra values to be set on state.
	StateExtra map[string]interface{}
}
