package middleware

import (
	"fmt"
	"net/http"
	"regexp"

	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	gatewayConfig "github.com/skygeario/skygear-server/pkg/gateway/config"
	gatewayModel "github.com/skygeario/skygear-server/pkg/gateway/model"
	"github.com/skygeario/skygear-server/pkg/gateway/store"
)

var gearPathRegex = regexp.MustCompile(`^/_([^\/]*)`)

// TenantAuthzMiddleware is middleware to check if the current app can access
// gear
type TenantAuthzMiddleware struct {
	Store         store.GatewayStore
	Configuration gatewayConfig.Configuration
}

// Handle reject the request if the current app doesn't have permission to
// access gear
func (a TenantAuthzMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := gatewayModel.GatewayContextFromContext(r.Context())
		app := ctx.App
		domain := ctx.Domain

		// Tenant authorization
		var gear gatewayModel.Gear
		if domain.Assignment == gatewayModel.AssignmentTypeMicroservices {
			// fallback route to gear by path
			gear = gatewayModel.Gear(getGearName(r.URL.Path))
		} else {
			gear = gatewayModel.Gear(domain.Assignment)
		}
		if gear == "" {
			// microservices
			next.ServeHTTP(w, r)
			return
		}

		gearVersion := app.GetGearVersion(gear)
		if !app.CanAccessGear(gear) {
			http.Error(w, fmt.Sprintf("%s is not support in current app plan", gear), http.StatusForbidden)
			return
		}

		url, err := a.Configuration.GetGearURL(gear, gearVersion)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if url == "" {
			http.Error(w, fmt.Sprintf("%s gear %s environment is not supported", gear, gearVersion), http.StatusVariantAlsoNegotiates)
			return
		}

		r.Header.Set(coreHttp.HeaderGear, string(gear))
		r.Header.Set(coreHttp.HeaderGearVersion, string(gearVersion))
		r.Header.Set(coreHttp.HeaderGearEndpoint, url)

		next.ServeHTTP(w, r)
	})
}

func getGearName(path string) string {
	result := gearPathRegex.FindStringSubmatch(path)
	if len(result) == 2 {
		return result[1]
	}

	return ""
}
