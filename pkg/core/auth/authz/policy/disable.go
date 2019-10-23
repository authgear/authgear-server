package policy

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	skyerr "github.com/skygeario/skygear-server/pkg/core/xskyerr"
)

// DenyDisabledUser denies disabled user.
// It is not an error if the request does not have an associated user.
// If you want to enforce enabled user, use both RequireAuthenticated and DenyDisabledUser.
func DenyDisabledUser(r *http.Request, ctx auth.ContextGetter) error {
	authInfo, _ := ctx.AuthInfo()
	if authInfo != nil && authInfo.IsDisabled() {
		details := skyerr.Details{}
		if authInfo.DisabledExpiry != nil {
			details["expiry"] = skyerr.APIErrorString(authInfo.DisabledExpiry.Format(time.RFC3339))
		}
		if authInfo.DisabledMessage != "" {
			details["message"] = skyerr.APIErrorString(authInfo.DisabledMessage)
		}

		return authz.UserDisabled.NewWithDetails("user is disabled", details)
	}
	return nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authz.PolicyFunc = DenyDisabledUser
)
