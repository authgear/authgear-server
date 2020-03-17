package policy

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
)

type All struct {
	policies []authz.Policy
}

func AllOf(policies ...authz.Policy) authz.Policy {
	return All{policies: policies}
}

func (p All) IsAllowed(r *http.Request) error {
	for _, policy := range p.policies {
		if err := policy.IsAllowed(r); err != nil {
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

func (p Any) IsAllowed(r *http.Request) error {
	// return the last authz error
	var err error
	for _, policy := range p.policies {
		err = policy.IsAllowed(r)
		if err == nil {
			return nil
		}
	}

	return err
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authz.Policy = &All{}
)
