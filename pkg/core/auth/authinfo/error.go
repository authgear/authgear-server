package authinfo

import (
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var UserNotFound = skyerr.NotFound.WithReason("UserNotFound")
var ErrNotFound = UserNotFound.New("user not found")
