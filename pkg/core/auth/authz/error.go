package authz

import (
	skyerr "github.com/skygeario/skygear-server/pkg/core/xskyerr"
)

var (
	NotAuthenticated     = skyerr.Unauthorized.WithReason("NotAuthenticated")
	AccessKeyNotAccepted = skyerr.Unauthorized.WithReason("AccessKeyNotAccepted")
	UserDisabled         = skyerr.Forbidden.WithReason("UserDisabled")
)

func NewNotAuthenticatedError() error {
	return NotAuthenticated.New("authentication required")
}
