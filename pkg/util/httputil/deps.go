package httputil

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	NewJSONResponseWriterLogger,
	wire.Struct(new(JSONResponseWriter), "*"),
	wire.Struct(new(FlashMessage), "*"),
	wire.Struct(new(TutorialCookie), "*"),
	MakeHTTPOrigin,
)
