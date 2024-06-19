package viewmodels

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	NewBaseLogger,
	wire.Struct(new(BaseViewModeler), "*"),
	wire.Struct(new(SettingsViewModeler), "*"),
	wire.Struct(new(SettingsProfileViewModeler), "*"),
	wire.Struct(new(AlternativeStepsViewModeler), "*"),
	wire.Struct(new(AuthenticationViewModeler), "*"),
	wire.Struct(new(ChangePasswordViewModeler), "*"),
	wire.Struct(new(AuthflowViewModeler), "*"),
	wire.Struct(new(InlinePreviewAuthflowBranchViewModeler), "*"),
)
