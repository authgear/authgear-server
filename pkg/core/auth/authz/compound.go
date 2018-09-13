package authz

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/handler"
)

type AllOfPolicy struct {
	policies []Policy
}

func NewAllOfPolicy(policies ...Policy) Policy {
	return AllOfPolicy{policies: policies}
}

func (p AllOfPolicy) IsAllowed(r *http.Request, ctx handler.AuthContext) error {
	for _, policy := range p.policies {
		if err := policy.IsAllowed(r, ctx); err != nil {
			return err
		}
	}

	return nil
}
