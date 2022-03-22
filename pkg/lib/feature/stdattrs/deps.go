package stdattrs

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(Service), "*"),
	wire.Struct(new(ServiceNoEvent), "*"),
	wire.Struct(new(PictureTransformer), "*"),
	wire.Bind(new(Transformer), new(*PictureTransformer)),
)
