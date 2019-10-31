package policy

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
)

// denyNotVerifiedUser denies not verified user.
func denyNotVerifiedUser(r *http.Request, ctx auth.ContextGetter) error {
	authInfo, _ := ctx.AuthInfo()

	if authInfo != nil && !authInfo.Verified {
		return authz.UserNotVerified.New("user is not verified")
	}

	return nil
}

var (
	_ authz.PolicyFunc = denyNotVerifiedUser
)
