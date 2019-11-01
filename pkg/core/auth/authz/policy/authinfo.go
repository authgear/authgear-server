package policy

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
)

func requireAuthenticated(r *http.Request, ctx auth.ContextGetter) error {
	authInfo, _ := ctx.AuthInfo()
	if authInfo == nil {
		return authz.ErrNotAuthenticated
	}

	return nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authz.PolicyFunc = requireAuthenticated
)
