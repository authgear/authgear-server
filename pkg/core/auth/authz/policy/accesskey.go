package policy

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authz"

	"github.com/skygeario/skygear-server/pkg/core/auth"
)

func RequireClient(r *http.Request, ctx auth.ContextGetter) error {
	key := auth.GetAccessKey(r.Context())
	if key.Client == nil {
		return authz.AccessKeyNotAccepted.New("API key required")
	}

	return nil
}

func RequireMasterKey(r *http.Request, ctx auth.ContextGetter) error {
	key := auth.GetAccessKey(r.Context())
	if !key.IsMasterKey {
		return authz.AccessKeyNotAccepted.New("Master key required")
	}

	return nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authz.PolicyFunc = RequireClient
	_ authz.PolicyFunc = RequireMasterKey
)
