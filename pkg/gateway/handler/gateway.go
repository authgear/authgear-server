package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/gateway"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

func NewGatewayHandler(gatewayDependency gateway.DependencyMap) http.Handler {
	return server.FactoryToHandler(&GatewayHandlerFactory{
		gatewayDependency,
	})
}

type GatewayHandlerFactory struct {
	Dependency gateway.DependencyMap
}

func (f *GatewayHandlerFactory) NewHandler(r *http.Request) http.Handler {
	ctx := model.GatewayContextFromContext(r.Context())
	gear := ctx.Gear
	var factory handler.Factory
	if gear != "" {
		factory = &GearHandlerFactory{f.Dependency}
	} else {
		factory = &DeploymentRouteHandlerFactory{f.Dependency}

	}
	return factory.NewHandler(r)
}
