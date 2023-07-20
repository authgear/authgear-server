package facade

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(Coordinator), "*"),
	wire.Struct(new(AuthenticatorFacade), "*"),
	wire.Struct(new(IdentityFacade), "*"),
	wire.Struct(new(UserFacade), "*"),
	wire.Struct(new(WorkflowVerificationFacade), "*"),
	wire.Struct(new(AdminVerificationFacade), "*"),
	wire.Struct(new(MFAFacade), "*"),
)
