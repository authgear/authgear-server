package httputil

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(FlashMessage), "*"),
	wire.Struct(new(TutorialCookie), "*"),
	MakeHTTPOrigin,
	GetRequestURL,
)
