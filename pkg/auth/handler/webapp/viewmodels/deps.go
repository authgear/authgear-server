package viewmodels

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(BaseViewModeler), "*"),
	wire.Struct(new(SettingsViewModeler), "*"),
)
