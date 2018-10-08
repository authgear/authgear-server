package policy

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/handler/context"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func RequireAuthenticated(r *http.Request, ctx context.AuthContext) error {
	if ctx.AuthInfo == nil {
		return skyerr.NewError(skyerr.NotAuthenticated, "require authenticated user")
	}

	if ctx.AuthInfo.TokenValidSince != nil {
		tokenValidSince := *ctx.AuthInfo.TokenValidSince

		// Not all types of access token support this field. The token is
		// still considered if it does not have an issue time.
		//
		// Due to precision, the issue time of the token can be before
		// AuthInfo.TokenValidSince. We consider the token still valid
		// if the token is issued within 1 second before tokenValidSince.
		if !ctx.Token.IssuedAt.IsZero() &&
			!ctx.Token.IssuedAt.After(tokenValidSince.Add(-1*time.Second)) {
			return skyerr.NewError(skyerr.NotAuthenticated, "require authenticated user")
		}
	}

	return nil
}
