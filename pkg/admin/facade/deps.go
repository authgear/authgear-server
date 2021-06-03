package facade

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(UserFacade), "*"),
	wire.Struct(new(IdentityFacade), "*"),
	wire.Struct(new(AuthenticatorFacade), "*"),
	wire.Struct(new(VerificationFacade), "*"),
	wire.Struct(new(SessionFacade), "*"),
	wire.Struct(new(AuditLogFacade), "*"),
)
