package userinfo

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(UserInfoService), "*"),
	wire.Struct(new(Sink), "*"),
)
