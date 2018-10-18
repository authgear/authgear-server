package policy

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
)

type All struct {
	policies []authz.Policy
}

func AllOf(policies ...authz.Policy) authz.Policy {
	return All{policies: policies}
}

func (p All) IsAllowed(r *http.Request, ctx auth.ContextGetter) error {
	for _, policy := range p.policies {
		if err := policy.IsAllowed(r, ctx); err != nil {
			return err
		}
	}

	return nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authz.Policy = &All{}
)
