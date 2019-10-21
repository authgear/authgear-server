package policy

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authz"

	"github.com/skygeario/skygear-server/pkg/core/auth"
)

func DenyNoAccessKey(r *http.Request, ctx auth.ContextGetter) error {
	key := ctx.AccessKey()
	if key.IsNoAccessKey() {
		return authz.AccessKeyNotAccepted.New("API key required")
	}

	return nil
}

func RequireMasterKey(r *http.Request, ctx auth.ContextGetter) error {
	key := ctx.AccessKey()
	if !key.IsMasterKey() {
		return authz.AccessKeyNotAccepted.New("Master key required")
	}

	return nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authz.PolicyFunc = DenyNoAccessKey
	_ authz.PolicyFunc = RequireMasterKey
)
