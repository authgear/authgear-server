package handler

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(GetHandler), "*"),
	wire.Struct(new(PostHandler), "*"),
)
