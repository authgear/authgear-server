package policy

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
)

func DenyInvalidSession(r *http.Request, ctx auth.ContextGetter) error {
	_, err := ctx.Session()
	if err == session.ErrSessionNotFound {
		return err
	}
	// ignore any other error
	return nil
}

var (
	_ authz.PolicyFunc = DenyInvalidSession
)
