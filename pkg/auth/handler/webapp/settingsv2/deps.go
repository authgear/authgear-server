package settingsv2

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(SettingsV2Handler), "*"),
)
