package policy

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type Everybody struct {
	allow bool
}

func (p Everybody) IsAllowed(r *http.Request, ctx handler.AuthContext) error {
	if !p.allow {
		// TODO:
		// return proper error code
		return skyerr.NewError(skyerr.UnexpectedError, "everybody is denied")
	}

	return nil
}
