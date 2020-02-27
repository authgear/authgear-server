package policy

import (
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
)

var RequireValidUser = AllOf(
	authz.PolicyFunc(requireAuthenticated),
	authz.PolicyFunc(DenyDisabledUser),
)

var RequireVerifiedUser = AllOf(
	authz.PolicyFunc(requireAuthenticated),
	authz.PolicyFunc(DenyDisabledUser),
	authz.PolicyFunc(denyNotVerifiedUser),
)

var RequireValidUserOrMasterKey = AnyOf(
	authz.PolicyFunc(RequireMasterKey),
	AllOf(
		authz.PolicyFunc(DenyNoAccessKey),
		RequireValidUser,
	),
)
