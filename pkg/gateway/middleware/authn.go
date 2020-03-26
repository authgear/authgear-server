package middleware

import (
	"net/http"
	"net/url"
	"strings"

	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	coreMiddleware "github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

// AuthnMiddleware call auth gear resolve endpoint and injects auth info headers
// into the request
type AuthnMiddleware struct {
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
		authHost := ctx.AuthHost
		// auth info headers are not needed for auth gear
		if gear == model.AuthGear {
			next.ServeHTTP(w, r)
			return
		}
		u := &url.URL{
			Scheme: corehttp.GetProto(r),
			Host:   authHost,
			Path:   "/_auth/session/resolve",
		}
		resolveReq, _ := http.NewRequest("GET", u.String(), nil)
		resolveReq.Header = r.Header.Clone()

		client := &http.Client{}
		resolveResp, err := client.Do(resolveReq)
		if err != nil {
			logger.WithError(err).Error("failed to call auth resolve endpoint")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// remove all X-Skygear-* headers
		for key := range r.Header {
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
