package policy

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// DenyDisabledUser denies disabled user.
// It is not an error if the request does not have an associated user.
// If you want to enforce enabled user, use both RequireAuthenticated and DenyDisabledUser.
func DenyDisabledUser(r *http.Request, ctx auth.ContextGetter) error {
	authInfo, _ := ctx.AuthInfo()
	if authInfo != nil && authInfo.Disabled {
		return skyerr.NewError(skyerr.UserDisabled, "user disabled")
	}
	return nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authz.PolicyFunc = DenyDisabledUser
)
