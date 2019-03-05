package policy

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
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

func (p Role) IsAllowed(r *http.Request, ctx auth.ContextGetter) error {
	authInfo := ctx.AuthInfo()
	if authInfo == nil {
		return skyerr.NewError(skyerr.UnexpectedAuthInfoNotFound, "user authentication info not found")
	}

	containsRole := authInfo.HasAnyRoles([]string{p.role})

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

func RequireAdminRole(r *http.Request, ctx auth.ContextGetter) error {
	err := skyerr.NewError(
		skyerr.PermissionDenied,
		"no permission to perform this action",
	)

	roles := ctx.Roles()
	if len(roles) == 0 {
		return err
	}

	isAdmin := false
	for _, v := range roles {
		isAdmin = isAdmin || v.IsAdmin
	}

	if !isAdmin {
		return err
	}

	return nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authz.Policy     = &Role{}
	_ authz.PolicyFunc = RequireAdminRole
)
