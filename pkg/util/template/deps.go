package template

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(Resolver), "*"),
	wire.Struct(new(Engine), "*"),

	wire.Bind(new(EngineTemplateResolver), new(*Resolver)),
)
