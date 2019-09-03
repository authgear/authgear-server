package policy

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func RequireAuthenticated(r *http.Request, ctx auth.ContextGetter) error {
	authInfo := ctx.AuthInfo()
	if authInfo == nil {
		return skyerr.NewNotAuthenticatedErr()
	}

	return nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authz.PolicyFunc = RequireAuthenticated
)
