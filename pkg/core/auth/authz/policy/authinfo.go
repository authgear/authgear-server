package policy

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

func requireAuthenticated(r *http.Request) error {
	user := authn.GetAuthInfo(r.Context())
	session := authn.GetSession(r.Context())
	if user == nil || session == nil {
		return authz.ErrNotAuthenticated
	}

	return nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authz.PolicyFunc = requireAuthenticated
)
