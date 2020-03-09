package policy

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// DenyDisabledUser denies disabled user.
// It is not an error if the request does not have an associated user.
// If you want to enforce enabled user, use RequireValidUser.
func DenyDisabledUser(r *http.Request, ctx auth.ContextGetter) error {
	authInfo, _ := ctx.AuthInfo()
	// FIXME(time): Switch to TimeProvider
	now := time.Now().UTC()
	if authInfo != nil && authInfo.IsDisabled(now) {
		details := skyerr.Details{}
		if authInfo.DisabledExpiry != nil {
			details["expiry"] = authInfo.DisabledExpiry.Format(time.RFC3339)
		}
		if authInfo.DisabledMessage != "" {
			details["message"] = authInfo.DisabledMessage
		}

		return authz.UserDisabled.NewWithInfo("user is disabled", details)
	}
	return nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authz.PolicyFunc = DenyDisabledUser
)
