package service

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(AppService), "*"),
	wire.Struct(new(AdminAPIService), "*"),
	wire.Struct(new(AuthzService), "*"),

	wire.Bind(new(AppAuthzService), new(*AuthzService)),
)
