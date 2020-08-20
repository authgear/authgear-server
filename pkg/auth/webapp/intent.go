package webapp

import (
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

type Intent struct {
	// OldStateID indicates the old state ID.
	// If it is not empty, some fields from the existing state would be restored.
	OldStateID string
	// RedirectURI indicates the location to redirect after the interaction finishes.
	RedirectURI string
	// KeepState indicates whether the state should be kept after the interaction finishes.
	// It is useful for interaction that has a dead end, such as forgot / reset password.
	// If it is true, then the state is attached to RedirectURI.
	KeepState bool
	UILocales string
	Intent    interaction.Intent
}
