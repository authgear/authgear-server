package handler

import (
	"net/http"

	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/gateway"
)

func NewGatewayHandler(gatewayDependency gateway.DependencyMap) http.Handler {
	return server.FactoryToHandler(&GatewayHandlerFactory{
		gatewayDependency,
	})
}

type GatewayHandlerFactory struct {
	Dependency gateway.DependencyMap
}

func (f *GatewayHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &GatewayHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h
}

type GatewayHandler struct {
}

func (h *GatewayHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	gearEndpoint := r.Header.Get(coreHttp.HeaderGearEndpoint)
	if gearEndpoint != "" {
		handleGear(rw, r)
	} else {
		handleDeploymentRoute(rw, r)
	}
}
