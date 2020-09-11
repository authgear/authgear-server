package httputil

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	NewJSONResponseWriterLogger,
	wire.Struct(new(JSONResponseWriter), "*"),
)
