package principal

import (
	"errors"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var ErrNotFound = errors.New("principal not found")
var ErrAlreadyExists = errors.New("principal already exists")
var ErrMultipleResultsFound = errors.New("multiple principals found")

// ErrCurrentIdentityBeingDeleted is shared by sso/unlink.go and loginid/remove.go
var ErrCurrentIdentityBeingDeleted = skyerr.Invalid.WithReason("CurrentIdentityBeingDeleted").New("must not delete current identity")
