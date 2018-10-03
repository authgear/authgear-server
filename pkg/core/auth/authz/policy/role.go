package policy

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/handler/context"

	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type Role struct {
	role  string
	allow bool
}

func NewAllowRole(role string) authz.Policy {
	return Role{role: role, allow: true}
}

func NewDenyRole(role string) authz.Policy {
	return Role{role: role, allow: false}
}

func (p Role) IsAllowed(r *http.Request, ctx context.AuthContext) error {
	if ctx.AuthInfo == nil {
		return skyerr.NewError(skyerr.UnexpectedAuthInfoNotFound, "user authentication info not found")
	}

	containsRole := ctx.AuthInfo.HasAnyRoles([]string{p.role})

	if p.allow && !containsRole {
		// TODO:
		// return proper error code
		return skyerr.NewError(skyerr.UnexpectedError, "Role "+p.role+" is required")
	}

	if !p.allow && containsRole {
		// TODO:
		// return proper error code
		return skyerr.NewError(skyerr.UnexpectedError, "Unexpected role "+p.role)
	}

	return nil
}
