package policy

import (
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
)

var RequireValidUser = AllOf(
	authz.PolicyFunc(requireAuthenticated),
	authz.PolicyFunc(DenyDisabledUser),
)

var RequireValidUserOrMasterKey = AnyOf(
	authz.PolicyFunc(RequireMasterKey),
	RequireValidUser,
)
