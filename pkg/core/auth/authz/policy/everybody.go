package policy

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/handler/context"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type Everybody struct {
	allow bool
}

func (p Everybody) IsAllowed(r *http.Request, ctx context.AuthContext) error {
	if !p.allow {
		// TODO:
		// return proper error code
		return skyerr.NewError(skyerr.UnexpectedError, "everybody is denied")
	}

	return nil
}
