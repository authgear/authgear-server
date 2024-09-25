//go:build wireinject
// +build wireinject

package images

import (
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/images/deps"
	"github.com/authgear/authgear-server/pkg/images/handler"
	imagesservice "github.com/authgear/authgear-server/pkg/images/service"
	"github.com/authgear/authgear-server/pkg/lib/images"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/presign"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/vipsutil"
)

func newPanicMiddleware(p *deps.RootProvider) httproute.Middleware {
	panic(wire.Build(
		deps.RootDependencySet,
		middleware.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.PanicMiddleware)),
	))
}

func newSentryMiddleware(p *deps.RootProvider) httproute.Middleware {
	panic(wire.Build(
		deps.RootDependencySet,
		wire.Struct(new(middleware.SentryMiddleware), "*"),
		wire.Bind(new(httproute.Middleware), new(*middleware.SentryMiddleware)),
	))
}

func newCORSMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		middleware.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.CORSMiddleware)),
	))
}

func newGetHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.DependencySet,
		handler.DependencySet,
		wire.Bind(new(handler.VipsDaemon), new(*vipsutil.Daemon)),
		wire.Bind(new(handler.DirectorMaker), new(*imagesservice.ImagesCloudStorageService)),
		wire.Bind(new(http.Handler), new(*handler.GetHandler)),
	))
}

func newPostHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.DependencySet,
		handler.DependencySet,
		images.DependencySet,
		wire.Bind(new(handler.JSONResponseWriter), new(*httputil.JSONResponseWriter)),
		wire.Bind(new(handler.PresignProvider), new(*presign.Provider)),
		wire.Bind(new(handler.ImagesStore), new(*images.Store)),
		wire.Bind(new(handler.PostHandlerCloudStorageService), new(*imagesservice.ImagesCloudStorageService)),
		wire.Bind(new(http.Handler), new(*handler.PostHandler)),
	))
}
