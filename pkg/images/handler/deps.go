package handler

import (
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	httputil.DependencySet,
	wire.Bind(new(JSONResponseWriter), new(*httputil.JSONResponseWriter)),

	NewGetHandlerLogger,
	wire.Struct(new(GetHandler), "*"),
	NewPostHandlerLogger,
	wire.Struct(new(PostHandler), "*"),
)
