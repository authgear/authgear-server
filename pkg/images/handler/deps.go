package handler

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewGetHandlerLogger,
	wire.Struct(new(GetHandler), "*"),
	NewPostHandlerLogger,
	wire.Struct(new(PostHandler), "*"),
)
