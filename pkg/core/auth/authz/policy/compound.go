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

type Any struct {
	policies []authz.Policy
}

func AnyOf(policies ...authz.Policy) authz.Policy {
	return Any{policies: policies}
}

func (p Any) IsAllowed(r *http.Request, ctx auth.ContextGetter) error {
	errs := []error{}

	for _, policy := range p.policies {
		var err error
		if err = policy.IsAllowed(r, ctx); err == nil {
			return nil
		}

		errs = append(errs, err)
	}

	// return the first error
	return errs[0]
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authz.Policy = &All{}
)
