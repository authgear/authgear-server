package policy

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// DenyDisabledUser denies disabled user.
// It is not an error if the request does not have an associated user.
// If you want to enforce enabled user, use RequireValidUser.
func DenyDisabledUser(r *http.Request) error {
	user := authn.GetUser(r.Context())
	// FIXME(time): Switch to TimeProvider
	now := time.Now().UTC()
	if user != nil && user.IsDisabled(now) {
		details := skyerr.Details{}
		if user.DisabledExpiry != nil {
			details["expiry"] = user.DisabledExpiry.Format(time.RFC3339)
		}
		if user.DisabledMessage != "" {
			details["message"] = user.DisabledMessage
		}

		return authz.UserDisabled.NewWithInfo("user is disabled", details)
	}
	return nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authz.PolicyFunc = DenyDisabledUser
)
