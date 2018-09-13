package authz

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type RolePolicy struct {
	role  string
	allow bool
}

func (p RolePolicy) IsAllowed(r *http.Request, ctx handler.AuthContext) error {
	if ctx.AuthInfo == nil {
		return skyerr.NewError(skyerr.UnexpectedAuthInfoNotFound, "user authentication info not found")
	}

	containsRole := false
	for _, role := range ctx.AuthInfo.Roles {
		if role == p.role {
			containsRole = true
			break
		}
	}

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
