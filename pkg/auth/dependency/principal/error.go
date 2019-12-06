package principal

import (
	"errors"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var ErrNotFound = errors.New("principal not found")
var ErrAlreadyExists = errors.New("principal already exists")
var ErrMultipleResultsFound = errors.New("multiple principals found")

var CurrentIdentityBeingDeleted = skyerr.Invalid.WithReason("CurrentIdentityBeingDeleted")
var ErrCurrentIdentityBeingDeleted = CurrentIdentityBeingDeleted.New("must not delete current identity")
