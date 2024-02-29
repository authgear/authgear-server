package transport

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(UIParamMiddleware), "*"),
	wire.Struct(new(GraphQLHandler), "*"),
	wire.Struct(new(PresignImagesUploadHandler), "*"),
	wire.Struct(new(UserImportHandler), "*"),
	NewPresignImagesUploadHandlerLogger,
)
