package loader

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	NewUserLoader,
	NewIdentityLoader,
	NewAuthenticatorLoader,
	NewRoleLoader,
	NewGroupLoader,
	NewAuditLogLoader,
	NewResourceLoader,
)
