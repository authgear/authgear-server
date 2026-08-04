package event

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

// PrepareBlockingEventOptions carries optional pre-computed inputs for
// preparing a blocking event.
type PrepareBlockingEventOptions struct {
	// ResolvedUser, when non-nil, populates the payload's resolve:"user" field
	// instead of reading the user from the database. Callers set it when they
	// have already read the same user within the same request.
	ResolvedUser *model.User
}
