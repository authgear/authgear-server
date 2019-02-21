package policy

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type Everybody struct {
	Allow bool
}

func (p Everybody) IsAllowed(r *http.Request, ctx auth.ContextGetter) error {
	if !p.Allow {
		// TODO:
		// return proper error code
		return skyerr.NewError(skyerr.UnexpectedError, "everybody is denied")
	}

	return nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authz.Policy = &Everybody{}
)
