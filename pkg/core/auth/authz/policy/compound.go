package policy

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/handler/context"
)

type All struct {
	policies []authz.Policy
}

func AllOf(policies ...authz.Policy) authz.Policy {
	return All{policies: policies}
}

func (p All) IsAllowed(r *http.Request, ctx context.AuthContext) error {
	for _, policy := range p.policies {
		if err := policy.IsAllowed(r, ctx); err != nil {
			return err
		}
	}

	return nil
}
