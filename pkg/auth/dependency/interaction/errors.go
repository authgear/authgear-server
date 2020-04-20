package interaction

import (
	"errors"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var ErrInteractionNotFound = errors.New("interaction not found")

var ErrInvalidStep = errors.New("step is invalid for current interaction state")

var InvalidCredentials = skyerr.Unauthorized.WithReason("InvalidCredentials")

var ErrInvalidCredentials = InvalidCredentials.New("invalid credentials")
