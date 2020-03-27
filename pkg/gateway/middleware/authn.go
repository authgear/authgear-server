package middleware

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"

	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	coreMiddleware "github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
	"github.com/skygeario/skygear-server/pkg/gateway/config"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

// AuthnMiddleware call auth gear resolve endpoint and injects auth info headers
// into the request
type AuthnMiddleware struct {
	GatewayConfiguration config.Configuration `dependency:"GatewayConfiguration"`
}

type AuthnMiddlewareFactory struct{}

func (f AuthnMiddlewareFactory) NewInjectableMiddleware() coreMiddleware.InjectableMiddleware {
	return &AuthnMiddleware{}
}

func (m *AuthnMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggerFactory := logging.NewFactoryFromRequest(r,
			logging.NewDefaultLogHook(nil),
			sentry.NewLogHookFromContext(r.Context()),
		)
		logger := loggerFactory.NewLogger("authn-middleware")

		ctx := model.GatewayContextFromContext(r.Context())
		gear := ctx.Gear
		// auth info headers are not needed for auth gear
		if gear == model.AuthGear {
			next.ServeHTTP(w, r)
			return
		}

		// get resolve endpoint from config
		var err error
		u, err := url.Parse(m.GatewayConfiguration.Auth.Live)
		if err != nil {
			logger.WithError(err).Error("invalid auth endpoint")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		u.Path = "/_auth/session/resolve"

		// make resolve endpoint request
		resolveReq, _ := http.NewRequest("GET", u.String(), nil)
		resolveReq.Header = r.Header.Clone()

		client := &http.Client{}
		resolveResp, err := client.Do(resolveReq)
		if err != nil {
			logger.WithError(err).Error("failed to call auth resolve endpoint")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer resolveResp.Body.Close()

		if resolveResp.StatusCode >= 400 && resolveResp.StatusCode < 500 {
			// if resolve response is 4xx, return the resolve response
			pipeResponse(w, resolveResp)
			return
		} else if resolveResp.StatusCode != 200 {
			logger.WithFields(logrus.Fields{
				"status_code":   resolveResp.StatusCode,
				"auth_endpoint": u.String(),
			}).Error("failed to call auth resolve endpoint")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// remove all X-Skygear-* headers expects tenant config for gears
		for key := range r.Header {
			// should not remove tenant config if request to gear
			if isTenantConfigHeader(key) {
				if gear != "" {
					continue
				}
			}
			if isSkygearHeader(key) {
				r.Header.Del(key)
			}
		}

		// copy resolve endpoint X-Skygear-* headers to request
		for key, values := range resolveResp.Header {
			if isSkygearHeader(key) {
				for _, v := range values {
					r.Header.Add(key, v)
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func isSkygearHeader(key string) bool {
	return strings.HasPrefix(strings.ToLower(key), "x-skygear")
}

func isTenantConfigHeader(key string) bool {
	return strings.EqualFold(key, corehttp.HeaderTenantConfig)
}

func pipeResponse(rw http.ResponseWriter, response *http.Response) {
	for key, values := range response.Header {
		for _, v := range values {
			rw.Header().Add(key, v)
		}
	}
	rw.WriteHeader(response.StatusCode)
	io.Copy(rw, response.Body)
}
