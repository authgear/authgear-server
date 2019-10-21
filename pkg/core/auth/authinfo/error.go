package authinfo

import (
	skyerr "github.com/skygeario/skygear-server/pkg/core/xskyerr"
)

var UserNotFound = skyerr.NotFound.WithReason("UserNotFound")
var ErrNotFound = UserNotFound.New("user not found")
