package policy

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

// DenyDisabledUser denies disabled user.
// It is not an error if the request does not have an associated user.
// If you want to enforce enabled user, use RequireValidUser.
func DenyDisabledUser(r *http.Request) error {
	user := authn.GetUser(r.Context())
	if user != nil && user.IsDisabled {
		return authz.UserDisabled.New("user is disabled")
	}
	return nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authz.PolicyFunc = DenyDisabledUser
)
