package handler

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/gateway"
	"github.com/skygeario/skygear-server/pkg/gateway/config"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

var gearPathRegex = regexp.MustCompile(`^/_([^\/]*)`)

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
	GatewayConfiguration config.Configuration `dependency:"GatewayConfiguration"`
}

func (h *GatewayHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := model.GatewayContextFromContext(r.Context())
	domain := ctx.Domain
	gear := getGearToRoute(&domain, r)

	if gear != "" {
		handleGear(gear, h.GatewayConfiguration, rw, r)
	} else {
		handleDeploymentRoute(rw, r)
	}
}

func getGearToRoute(domain *model.Domain, r *http.Request) model.Gear {
	if domain.Assignment == model.AssignmentTypeDefault {
		host := r.Host
		if host == domain.Domain {
			// microservices
			// fallback route to gear if necessary
			return model.Gear(getGearName(r.URL.Path))
		}
		// get gear from host
		parts := strings.Split(host, ".")
		return model.GetGear(parts[0])
	}
	if domain.Assignment == model.AssignmentTypeMicroservices {
		// fallback route to gear by path
		// return empty string if it is not matched
		return model.Gear(getGearName(r.URL.Path))
	}
	return model.Gear(domain.Assignment)
}

func getGearName(path string) string {
	result := gearPathRegex.FindStringSubmatch(path)
	if len(result) == 2 {
		return result[1]
	}

	return ""
}
