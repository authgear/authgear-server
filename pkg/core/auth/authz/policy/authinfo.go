package policy

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func RequireAuthenticated(r *http.Request, ctx auth.ContextGetter) error {
	authInfo := ctx.AuthInfo()
	if authInfo == nil {
		return skyerr.NewError(skyerr.NotAuthenticated, "require authenticated user")
	}

	if authInfo.TokenValidSince != nil {
		tokenValidSince := authInfo.TokenValidSince

		// Not all types of access token support this field. The token is
		// still considered if it does not have an issue time.
		//
		// Due to precision, the issue time of the token can be before
		// AuthInfo.TokenValidSince. We consider the token still valid
		// if the token is issued within 1 second before tokenValidSince.
		token := ctx.Token()
		if !token.IssuedAt.IsZero() &&
			!token.IssuedAt.After(tokenValidSince.Add(-1*time.Second)) {
			return skyerr.NewError(skyerr.NotAuthenticated, "require authenticated user")
		}
	}

	return nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authz.PolicyFunc = RequireAuthenticated
)
