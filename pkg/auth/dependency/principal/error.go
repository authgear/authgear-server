package principal

import (
	"errors"

	skyerr "github.com/skygeario/skygear-server/pkg/core/xskyerr"
)

var ErrNotFound = errors.New("principal not found")
var ErrAlreadyExists = errors.New("principal already exists")

var CurrentIdentityBeingDeleted = skyerr.Invalid.WithReason("CurrentIdentityBeingDeleted")
var ErrCurrentIdentityBeingDeleted = CurrentIdentityBeingDeleted.New("must not delete current identity")
