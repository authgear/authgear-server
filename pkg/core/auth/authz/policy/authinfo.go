package policy

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func RequireAuthenticated(r *http.Request, ctx auth.ContextGetter) error {
	authInfo, err := ctx.AuthInfo()
	if authInfo == nil {
		if err == session.ErrSessionNotFound {
			return err
		}
		return skyerr.NewNotAuthenticatedErr()
	}

	return nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authz.PolicyFunc = RequireAuthenticated
)
