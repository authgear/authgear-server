package policy

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type Everybody struct {
	allow bool
}

func (p Everybody) IsAllowed(r *http.Request, ctx auth.ContextGetter) error {
	if !p.allow {
		// TODO:
		// return proper error code
		return skyerr.NewError(skyerr.UnexpectedError, "everybody is denied")
	}

	return nil
}
