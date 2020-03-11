package handler

import (
	"net/http"

	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
)

func NewGatewayHandler() http.HandlerFunc {
	return http.HandlerFunc(handleGateway)
}

func handleGateway(rw http.ResponseWriter, r *http.Request) {
	gearEndpoint := r.Header.Get(coreHttp.HeaderGearEndpoint)
	if gearEndpoint != "" {
		handleGear(rw, r)
	} else {
		handleDeploymentRoute(rw, r)
	}
}
